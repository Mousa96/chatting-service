// Package service implements the WebSocket business logic
package service

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	messageModels "github.com/Mousa96/chatting-service/internal/message/models"
	"github.com/Mousa96/chatting-service/internal/message/service"
	"github.com/Mousa96/chatting-service/internal/websocket/models"
	"github.com/Mousa96/chatting-service/internal/websocket/repository"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB for text and media metadata
)

// WebSocketService provides the implementation of the Service interface
type WebSocketService struct {
	repo          repository.Repository
	messageService service.Service
	
	// Message throttling
	throttleLimit   int
	throttleWindow  time.Duration
	messageCounters map[int][]time.Time
	throttleMutex   sync.Mutex
}

// NewWebSocketService creates a new WebSocketService instance
func NewWebSocketService(repo repository.Repository, messageService service.Service) Service {
	return &WebSocketService{
		repo:          repo,
		messageService: messageService,
	}
}

// NewWebSocketServiceWithThrottling creates a new WebSocketService instance with throttling
func NewWebSocketServiceWithThrottling(
	repo repository.Repository,
	msgService service.Service,
	limit int,
	window time.Duration,
) *WebSocketService {
	return &WebSocketService{
		repo:            repo,
		messageService:  msgService,
		throttleLimit:   limit,
		throttleWindow:  window,
		messageCounters: make(map[int][]time.Time),
		throttleMutex:   sync.Mutex{},
	}
}

// HandleConnection manages a new WebSocket connection for a user
func (s *WebSocketService) HandleConnection(conn *websocket.Conn, userID int) error {
	client := &models.Client{
		UserID:     userID,
		Connection: conn,
		Send:       make(chan []byte, 256), // Buffer for outbound messages
	}

	// Register client
	if err := s.repo.AddClient(client); err != nil {
		return fmt.Errorf("failed to register client: %w", err)
	}

	// Notify others that user is online
	statusEvent := &models.Event{
		Type:      models.EventUserStatus,
		Timestamp: time.Now(),
		Payload: models.UserStatusEvent{
			UserID: userID,
			Status: models.StatusOnline,
		},
	}
	s.BroadcastEvent(statusEvent)

	// Start goroutines for pumping messages
	go s.readPump(client)
	go s.writePump(client)

	return nil
}

// CloseConnection closes a user's WebSocket connection
func (s *WebSocketService) CloseConnection(userID int) error {
	client, err := s.repo.GetClient(userID)
	if err != nil {
		return err
	}

	// Close the client's send channel to signal the writePump to exit
	close(client.Send)

	// Remove client from repository
	if err := s.repo.RemoveClient(userID); err != nil {
		return err
	}

	// Notify others that user is offline
	statusEvent := &models.Event{
		Type:      models.EventUserStatus,
		Timestamp: time.Now(),
		Payload: models.UserStatusEvent{
			UserID: userID,
			Status: models.StatusOffline,
		},
	}
	s.BroadcastEvent(statusEvent)

	return nil
}

// SendEvent sends an event to a specific user
func (s *WebSocketService) SendEvent(receiverID int, event *models.Event) error {
	client, err := s.repo.GetClient(receiverID)
	if err != nil {
		return err // User is not connected
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	select {
	case client.Send <- data:
		return nil
	default:
		// Channel is full or closed - connection might be dead
		s.CloseConnection(receiverID)
		return fmt.Errorf("failed to send event, connection closed")
	}
}

// BroadcastEvent sends an event to all connected users
func (s *WebSocketService) BroadcastEvent(event *models.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	clients, err := s.repo.GetAllClients()
	if err != nil {
		return fmt.Errorf("failed to get clients: %w", err)
	}

	for _, client := range clients {
		// Skip sender if the event has a sender ID
		if event.SenderID > 0 && client.UserID == event.SenderID {
			continue
		}

		select {
		case client.Send <- data:
		default:
			// Non-blocking send - if channel is full, close connection
			s.CloseConnection(client.UserID)
		}
	}

	return nil
}

// NotifyMessageSent notifies a user about a new message
func (s *WebSocketService) NotifyMessageSent(message *messageModels.Message) error {
	event := &models.Event{
		Type:      models.EventMessage,
		SenderID:  message.SenderID,
		Timestamp: time.Now(),
		Message:   message,
	}

	return s.SendEvent(message.ReceiverID, event)
}

// NotifyStatusChange notifies about message status changes (read/delivered)
func (s *WebSocketService) NotifyStatusChange(messageID int, status messageModels.MessageStatus) error {
	// First, get the message to know who to notify
	message, err := s.messageService.GetMessageByID(messageID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	event := &models.Event{
		Type:      models.EventStatusChange,
		SenderID:  message.ReceiverID, // The receiver is updating the status
		Timestamp: time.Now(),
		Payload: models.StatusUpdateEvent{
			MessageID: messageID,
			Status:    status,
		},
	}

	// Notify the sender of the message about the status change
	return s.SendEvent(message.SenderID, event)
}

// NotifyTypingStatus sends typing indicators between users
func (s *WebSocketService) NotifyTypingStatus(senderID, receiverID int, isTyping bool) error {
	event := &models.Event{
		Type:      models.EventTyping,
		SenderID:  senderID,
		Timestamp: time.Now(),
		Payload: models.TypingEvent{
			UserID:     senderID,
			IsTyping:   isTyping,
			ReceiverID: receiverID,
		},
	}

	return s.SendEvent(receiverID, event)
}

// UpdateUserStatus changes a user's online status
func (s *WebSocketService) UpdateUserStatus(userID int, status models.UserStatus) error {
	if err := s.repo.UpdateUserStatus(userID, status); err != nil {
		return err
	}

	// Notify others about the status change
	event := &models.Event{
		Type:      models.EventUserStatus,
		SenderID:  userID,
		Timestamp: time.Now(),
		Payload: models.UserStatusEvent{
			UserID: userID,
			Status: status,
		},
	}

	return s.BroadcastEvent(event)
}

// GetUserStatus gets a user's current online status
func (s *WebSocketService) GetUserStatus(userID int) (models.UserStatus, error) {
	return s.repo.GetUserStatus(userID)
}

// GetConnectedUsers returns a list of all currently connected users
func (s *WebSocketService) GetConnectedUsers() ([]int, error) {
	clients, err := s.repo.GetAllClients()
	if err != nil {
		return nil, err
	}

	userIDs := make([]int, len(clients))
	for i, client := range clients {
		userIDs[i] = client.UserID
	}

	return userIDs, nil
}

// readPump pumps messages from the WebSocket connection to the hub
func (s *WebSocketService) readPump(client *models.Client) {
	defer func() {
		s.CloseConnection(client.UserID)
	}()

	client.Connection.SetReadLimit(maxMessageSize)
	client.Connection.SetReadDeadline(time.Now().Add(pongWait))
	client.Connection.SetPongHandler(func(string) error {
		client.Connection.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := client.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, 
				websocket.CloseGoingAway, 
				websocket.CloseAbnormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			break
		}

		// Use the new throttling-aware message handler
		throttled := s.handleRawMessage(client.UserID, message)
		if throttled {
			// If the message was throttled, we've already sent an error response
			continue
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (s *WebSocketService) writePump(client *models.Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.Connection.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Connection.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				client.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Connection.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.Connection.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessageEvent processes message events from clients
func (s *WebSocketService) handleMessageEvent(senderID int, event *models.Event) {
	// Extract the message from the event
	if event.Message == nil {
		log.Println("received message event with no message")
		return
	}

	// Save to database via message service
	createReq := &messageModels.CreateMessageRequest{
		ReceiverID: event.Message.ReceiverID,
		Content:    event.Message.Content,
		MediaURL:   event.Message.MediaURL,
	}
	savedMsg, err := s.messageService.SendMessage(senderID, createReq)
	if err != nil {
		log.Printf("failed to save message: %v", err)
		return
	}

	// Update the event with the saved message (which now has an ID)
	event.Message = savedMsg

	// Forward the message to the recipient
	event.SenderID = senderID
	if event.Message.ReceiverID > 0 {
		s.SendEvent(event.Message.ReceiverID, event)
	}
}

// handleTypingEvent processes typing events from clients
func (s *WebSocketService) handleTypingEvent(senderID int, event *models.Event) {
	typingEvent, ok := event.Payload.(map[string]interface{})
	if !ok {
		log.Println("invalid typing event payload")
		return
	}

	// Extract receiver ID and isTyping from the payload
	receiverID, ok := typingEvent["receiver_id"].(float64)
	if !ok {
		log.Println("invalid receiver_id in typing event")
		return
	}

	isTyping, ok := typingEvent["is_typing"].(bool)
	if !ok {
		log.Println("invalid is_typing in typing event")
		return
	}

	// Forward the typing event to the recipient
	s.NotifyTypingStatus(senderID, int(receiverID), isTyping)
}

// handleStatusChangeEvent processes message status change events
func (s *WebSocketService) handleStatusChangeEvent(senderID int, event *models.Event) {
	statusEvent, ok := event.Payload.(map[string]interface{})
	if !ok {
		log.Println("invalid status change event payload")
		return
	}

	// Extract message ID and status from the payload
	messageID, ok := statusEvent["message_id"].(float64)
	if !ok {
		log.Println("invalid message_id in status change event")
		return
	}

	statusStr, ok := statusEvent["status"].(string)
	if !ok {
		log.Println("invalid status in status change event")
		return
	}

	// Update the message status in the database
	status := messageModels.MessageStatus(statusStr)
	if !status.IsValid() {
		log.Println("invalid message status:", statusStr)
		return
	}

	// This would ideally call the existing message service
	log.Printf("Updating message %d status to %s", int(messageID), status)
}

// handleBroadcastEvent processes broadcast events from clients
func (s *WebSocketService) handleBroadcastEvent(senderID int, event *models.Event) {
	// Extract the message from the event
	if event.Message == nil {
		log.Println("received broadcast event with no message")
		return
	}
	
	// Extract receiver IDs from the payload
	var receiverIDs []int
	
	// Check if receiver_ids is in the Message struct
	if event.Message.ReceiverIDs != nil && len(event.Message.ReceiverIDs) > 0 {
		receiverIDs = event.Message.ReceiverIDs
	} else {
		log.Println("broadcast event has no valid recipients")
		return
	}
	
	// Create broadcast request
	broadcastReq := &messageModels.BroadcastMessageRequest{
		ReceiverIDs: receiverIDs,
		Content:     event.Message.Content,
		MediaURL:    event.Message.MediaURL,
	}
	
	// Save messages to database
	messages, err := s.messageService.BroadcastMessage(senderID, broadcastReq)
	if err != nil {
		log.Printf("failed to save broadcast messages: %v", err)
		return
	}
	
	// Send to each recipient
	for _, msg := range messages {
		broadcastEvent := &models.Event{
			Type:      models.EventMessage, // Recipients get it as a normal message
			SenderID:  senderID,
			Timestamp: time.Now(),
			Message:   msg,
		}
		
		// Only send to connected users
		s.SendEvent(msg.ReceiverID, broadcastEvent)
	}
}

// isThrottled checks if a message should be throttled
func (s *WebSocketService) isThrottled(userID int) bool {
	s.throttleMutex.Lock()
	defer s.throttleMutex.Unlock()
	
	// If no limit is set, don't throttle
	if s.throttleLimit <= 0 {
		return false
	}
	
	now := time.Now()
	
	// Get the user's message timestamps
	timestamps, exists := s.messageCounters[userID]
	if !exists {
		// First message from this user
		s.messageCounters[userID] = []time.Time{now}
		return false
	}
	
	// Filter out timestamps outside the window
	validTimestamps := []time.Time{}
	for _, ts := range timestamps {
		if now.Sub(ts) <= s.throttleWindow {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	
	// Check if the user is over the limit
	if len(validTimestamps) >= s.throttleLimit {
		return true
	}
	
	// Add the current timestamp and update
	s.messageCounters[userID] = append(validTimestamps, now)
	return false // Message was not throttled
}

// handleRawMessage processes a raw websocket message with throttling
func (s *WebSocketService) handleRawMessage(userID int, message []byte) bool {
	// Check if the user is being throttled
	if s.isThrottled(userID) {
		// User is sending too many messages
		log.Printf("User %d is being throttled - too many messages", userID)
		
		// Send a throttling message to the user
		errorEvent := &models.Event{
			Type:    models.EventError,
			Message: &messageModels.Message{Content: "You are sending messages too quickly. Please slow down."},
		}
		
		errorJSON, _ := json.Marshal(errorEvent)
		client, err := s.repo.GetClient(userID)
		if err == nil && client != nil {
			client.Send <- errorJSON
		}
		
		return true // Message was throttled
	}
	
	// Parse the message
	var event models.Event
	if err := json.Unmarshal(message, &event); err != nil {
		log.Printf("Error parsing WebSocket message: %v", err)
		return false
	}
	
	// Handle the event based on its type
	switch event.Type {
	case models.EventMessage:
		s.handleMessageEvent(userID, &event)
	case models.EventTyping:
		s.handleTypingEvent(userID, &event)
	case models.EventStatusChange:
		s.handleStatusChangeEvent(userID, &event)
	case models.EventBroadcast:
		s.handleBroadcastEvent(userID, &event)
	default:
		log.Printf("Unknown event type: %s", event.Type)
	}
	
	return false // Message was not throttled
}
