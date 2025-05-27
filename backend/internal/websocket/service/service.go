package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Mousa96/chatting-service/internal/message/models"
	"github.com/Mousa96/chatting-service/internal/message/service"
	"github.com/Mousa96/chatting-service/internal/middleware"
	websocketModels "github.com/Mousa96/chatting-service/internal/websocket/models"
	"github.com/gorilla/websocket"
)

// Add new constant for user status event
const (
	EventUserStatus = "user_status"
)

type WebSocketService struct {
	upgrader  websocket.Upgrader
	clients ClientList
	userClients    map[int]*Client 
	sync.RWMutex
	handlers map[string]EventHandler
	messageService service.Service
	jwtKey         []byte
	pendingMessages map[int][]*models.Message // userID -> pending messages
    pendingMutex    sync.RWMutex
}

type SendResult struct {
	Success    bool
	UserOnline bool
	Error      error
}

func NewWebSocketService(messageService service.Service, jwtKey []byte) *WebSocketService {
	m :=&WebSocketService{
		clients: make(ClientList),
		handlers: make(map[string]EventHandler),
		upgrader: websocket.Upgrader{
			ReadBufferSize: 1024,
			WriteBufferSize: 1024,
			CheckOrigin: checkOrigin,
		},
		userClients: make(map[int]*Client),
		messageService: messageService,
		jwtKey: jwtKey,
		pendingMessages: make(map[int][]*models.Message),

	}
	m.setupEventHandlers()
	return m
}

func (s *WebSocketService) setupEventHandlers() {
	s.handlers[websocketModels.EventSendMessage] = sendMessage
	s.handlers[websocketModels.EventBroadcastMessage] = broadcastMessage
	s.handlers[websocketModels.EventMessageRead] = HandleMessageRead
	s.handlers[websocketModels.EventGetOnlineUsers] = handleGetOnlineUsers
}

func sendMessage(event *websocketModels.Event, c *Client) error {
	fmt.Println("Sending message:", event)
	var sendMessageEvent websocketModels.SendMessageEvent
	if err := json.Unmarshal(event.Payload, &sendMessageEvent); err != nil {
		return fmt.Errorf("error unmarshalling event: %v", err)
	}
	
	// Save the message to the database
	savedMessage, err := c.wsService.messageService.SendMessage(c.userID, &models.CreateMessageRequest{
		ReceiverID: sendMessageEvent.To,
		Content: sendMessageEvent.Message,
		MediaURL: sendMessageEvent.MediaURL,
	})
	if err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}

	messagePayload := websocketModels.MessagePayload{
		ID:         savedMessage.ID,
		SenderID:   savedMessage.SenderID,
		ReceiverID: savedMessage.ReceiverID,
		Content:    savedMessage.Content,
		MediaURL:   savedMessage.MediaURL,
		Status:     websocketModels.StatusSent,
		CreatedAt:  savedMessage.CreatedAt.Format(time.RFC3339),
	}
	
	messageEvent := websocketModels.Event{
		Type: websocketModels.EventReceiveMessage,
		Payload: mustMarshal(messagePayload),
	}

	// Send confirmation to sender (this should always succeed since sender is connected)
	senderResult := c.wsService.sendMessageToClient(c.userID, messageEvent)
	if senderResult.Error != nil {
		log.Printf("Failed to send confirmation to sender %d: %v", c.userID, senderResult.Error)
		// Don't return error here - message was saved, just confirmation failed
	}

	// Send to recipient
	recipientResult := c.wsService.sendMessageToClient(sendMessageEvent.To, messageEvent)
	if recipientResult.Error != nil {
		log.Printf("Failed to send message to recipient %d: %v", sendMessageEvent.To, recipientResult.Error)
		// Don't return error - message was saved, recipient just has connection issues
	} else if !recipientResult.UserOnline {
		log.Printf("Recipient %d is offline, adding to pending queue", sendMessageEvent.To)
        // ADD THIS: Add to pending queue when user is offline
        c.wsService.addToPendingQueue(savedMessage)
	}

	// Mark as delivered if message was successfully sent
	if recipientResult.Success {
		c.wsService.markAsDelivered(savedMessage.ID, sendMessageEvent.To)
	}

	return nil
}

func broadcastMessage(event *websocketModels.Event, c *Client) error {
	var broadcastMessageEvent websocketModels.BroadcastMessageEvent
	if err := json.Unmarshal(event.Payload, &broadcastMessageEvent); err != nil {
		return fmt.Errorf("error unmarshalling event: %v", err)
	}

	// Save the messages to the database
	savedMessages, err := c.wsService.messageService.BroadcastMessage(c.userID, &models.BroadcastMessageRequest{
		ReceiverIDs: broadcastMessageEvent.ReceiverIDs,
		Content:     broadcastMessageEvent.Message,
		MediaURL:    broadcastMessageEvent.MediaURL,
	})
	if err != nil {
		return fmt.Errorf("error broadcasting message: %v", err)
	}

	successCount := 0
	offlineCount := 0
	errorCount := 0

	// Send individual messages to each recipient
	for _, message := range savedMessages {
		messagePayload := websocketModels.MessagePayload{
			ID:         message.ID,
			SenderID:   message.SenderID,
			ReceiverID: message.ReceiverID,
			Content:    message.Content,
			MediaURL:   message.MediaURL,
			Status:     websocketModels.StatusSent,
			CreatedAt:  message.CreatedAt.Format(time.RFC3339),
		}

		messageEvent := websocketModels.Event{
			Type: websocketModels.EventReceiveMessage,
			Payload: mustMarshal(messagePayload),
		}

		// Send to sender (confirmation)
		senderResult := c.wsService.sendMessageToClient(c.userID, messageEvent)
		if senderResult.Error != nil {
			log.Printf("Failed to send broadcast confirmation to sender %d: %v", c.userID, senderResult.Error)
		}

		// Send to recipient
		recipientResult := c.wsService.sendMessageToClient(message.ReceiverID, messageEvent)
		if recipientResult.Error != nil {
			log.Printf("Failed to send broadcast to user %d: %v", message.ReceiverID, recipientResult.Error)
			errorCount++
		} else if !recipientResult.UserOnline {
			log.Printf("Recipient %d is offline, adding to pending queue", message.ReceiverID)
			// ADD THIS: Add to pending queue when user is offline
			c.wsService.addToPendingQueue(message)
		} else if recipientResult.Success {
			successCount++
			// Auto-mark as delivered if recipient received it
			c.wsService.markAsDelivered(message.ID, message.ReceiverID)
		}
	}
	
	log.Printf("Broadcast from user %d: %d delivered, %d offline, %d errors out of %d total", 
		c.userID, successCount, offlineCount, errorCount, len(savedMessages))
	return nil
}


func (s *WebSocketService) markAsDelivered(messageID int, recipientID int) {
	// Update status to delivered in database
	err := s.messageService.UpdateMessageStatus(messageID, models.StatusDelivered, recipientID)
	if err != nil {
		log.Printf("Failed to update message status to delivered: %v", err)
		return
	}

	// Get the message to find sender
	message, err := s.messageService.GetMessageByID(messageID)
	if err != nil {
		log.Printf("Error getting message: %v", err)
		return
	}

	// Send status change notification
	statusChangeEvent := websocketModels.Event{
		Type: websocketModels.EventStatusChange,
		Payload: mustMarshal(websocketModels.StatusChangeEvent{
			MessageID: messageID,
			Status:    websocketModels.StatusDelivered,
			UserID:    recipientID,
		}),
	}
	
	// Notify sender that message was delivered
	senderResult := s.sendMessageToClient(message.SenderID, statusChangeEvent)
	if senderResult.Error != nil {
		log.Printf("Failed to notify sender %d of delivery: %v", message.SenderID, senderResult.Error)
	} else if !senderResult.UserOnline {
		log.Printf("Sender %d is offline, delivery notification not sent", message.SenderID)
	}
}
// Add message to pending queue when user is offline
func (s *WebSocketService) addToPendingQueue(message *models.Message) {
    s.pendingMutex.Lock()
    defer s.pendingMutex.Unlock()
    
    userID := message.ReceiverID
    s.pendingMessages[userID] = append(s.pendingMessages[userID], message)
    log.Printf("Added message %d to pending queue for user %d", message.ID, userID)
}
// Process pending messages when user comes online
func (s *WebSocketService) processPendingMessages(userID int) {
    s.pendingMutex.Lock()
    pending := s.pendingMessages[userID]
    delete(s.pendingMessages, userID) // Clear the queue
    s.pendingMutex.Unlock()
    
    if len(pending) == 0 {
        return
    }
    
    log.Printf("Processing %d pending messages for user %d", len(pending), userID)
    
    for _, message := range pending {
        messagePayload := websocketModels.MessagePayload{
            ID:         message.ID,
            SenderID:   message.SenderID,
            ReceiverID: message.ReceiverID,
            Content:    message.Content,
            MediaURL:   message.MediaURL,
            Status:     websocketModels.StatusDelivered,
            CreatedAt:  message.CreatedAt.Format(time.RFC3339),
        }
        
        messageEvent := websocketModels.Event{
            Type: websocketModels.EventReceiveMessage,
            Payload: mustMarshal(messagePayload),
        }
        
        // Send to user
        result := s.sendMessageToClient(userID, messageEvent)
        if result.Success {
            // Mark as delivered
            s.markAsDelivered(message.ID, userID)
        }
    }
}
// FIXED: Remove duplicate status notifications
func (s *WebSocketService) MarkMessageAsRead(messageID, userID int) {
	// Update status in database
	err := s.messageService.UpdateMessageStatus(messageID, models.StatusRead, userID)
	if err != nil {
		log.Printf("Error updating message %d to read: %v", messageID, err)
		return
	}

	// Get message to find sender
	message, err := s.messageService.GetMessageByID(messageID)
	if err != nil {
		log.Printf("Error getting message: %v", err)
		return
	}

	// Send unified status change event
	statusChangeEvent := websocketModels.Event{
		Type: websocketModels.EventStatusChange,
		Payload: mustMarshal(websocketModels.StatusChangeEvent{
			MessageID: messageID,
			Status:    websocketModels.StatusRead,
			UserID:    userID,
		}),
	}

	// FIXED: Only notify sender ONCE, not both sender and reader
	senderResult := s.sendMessageToClient(message.SenderID, statusChangeEvent)
	if senderResult.Error != nil {
		log.Printf("Failed to notify sender of read receipt: %v", senderResult.Error)
	} else if !senderResult.UserOnline {
		log.Printf("Sender %d is offline, read receipt not delivered", message.SenderID)
	}
}

func HandleMessageRead(event *websocketModels.Event, c *Client) error {
	var readPayload struct {
		MessageID int `json:"message_id"`
	}
	
	if err := json.Unmarshal(event.Payload, &readPayload); err != nil {
		return err
	}

	c.wsService.MarkMessageAsRead(readPayload.MessageID, c.userID)
	return nil
}

func handleGetOnlineUsers(event *websocketModels.Event, c *Client) error {
    onlineUsers := c.wsService.getOnlineUserIDs()
    
    // Send current online users to the requesting client
    for _, userID := range onlineUsers {
        if userID != c.userID { // Don't send own status
            statusEvent := websocketModels.Event{
                Type: websocketModels.EventUserStatus,
                Payload: mustMarshal(websocketModels.UserStatusEvent{
                    UserID: userID,
                    Status: "online",
                }),
            }
            
            select {
            case c.egress <- statusEvent:
            default:
                // Buffer full, skip
            }
        }
    }
    
    return nil
}

func (s *WebSocketService) getOnlineUserIDs() []int {
    s.RLock()
    defer s.RUnlock()
    
    var onlineUsers []int
    for userID := range s.userClients {
        onlineUsers = append(onlineUsers, userID)
    }
    return onlineUsers
}

func (s *WebSocketService) routeEvent(event *websocketModels.Event, c *Client) error {
	if handler, ok := s.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			log.Printf("error handling event: %v", err)
			return err
		}
		return nil
	} else {
		log.Printf("no handler for event: %v", event)
		return fmt.Errorf("no handler for event: %v", event)
	}
}

func (s *WebSocketService) ServeWs(w http.ResponseWriter, r *http.Request) {
	// get the token from the query string
	token := r.URL.Query().Get("token")
    if token == "" {
        http.Error(w, "Missing authentication token", http.StatusUnauthorized)
        return
    }
	// validate the token
	userID, err := middleware.ValidateTokenAndGetUserID(token, string(s.jwtKey))
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	// upgrade the connection to a websocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error during connection upgrade:", err)
		return
	}

	client := NewClient(conn, s, userID)
	s.addClient(client)

	// start the client read and write processes
	go client.readMessages()
	go client.writeMessages()
}

func (s *WebSocketService) addClient(client *Client) {
    s.Lock()
    defer s.Unlock()

    // Close existing connection if user reconnects
    if existingClient, exists := s.userClients[client.userID]; exists {
        existingClient.connection.Close()
        delete(s.clients, existingClient)
    }

    s.clients[client] = true
    s.userClients[client.userID] = client
    
    // Broadcast online status after adding client
    go s.broadcastUserStatus(client.userID, "online")
    
    // Send current online users to the new client
    go func() {
        onlineUsers := s.getOnlineUserIDs()
        for _, userID := range onlineUsers {
            if userID != client.userID {
                statusEvent := websocketModels.Event{
                    Type: websocketModels.EventUserStatus,
                    Payload: mustMarshal(websocketModels.UserStatusEvent{
                        UserID: userID,
                        Status: "online",
                    }),
                }
                client.egress <- statusEvent
            }
        }
    }()

    // Process pending messages for user who just came online
    go s.processPendingMessages(client.userID)
}

func (s *WebSocketService) removeClient(client *Client) {
    s.Lock()
    defer s.Unlock()

    if _, ok := s.clients[client]; ok {
        client.connection.Close()
        delete(s.clients, client)
        delete(s.userClients, client.userID)
        
        // Broadcast offline status after removing client
        go s.broadcastUserStatus(client.userID, "offline")
    }
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	log.Println("origin:", origin)
	switch origin {
	case "http://localhost:8080":
		return true
	default:
		return false
	}
}

func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("Error marshaling payload: %v", err)
        return json.RawMessage(`{"error": "marshal failed"}`)
	}
	return json.RawMessage(data)
}

func (s *WebSocketService) sendMessageToClient(userID int, event websocketModels.Event) SendResult {
	s.RLock()
	client, ok := s.userClients[userID]
	s.RUnlock()
	if !ok {
		// User is offline - this is normal, not an error
		log.Printf("User %d is offline, message not delivered via WebSocket", userID)
		return SendResult{
			Success:    false,
			UserOnline: false,
			Error:      nil, // No error - user is just offline
		}
	}
	
	// Try to send the message
	select {
	case client.egress <- event:
		return SendResult{
			Success:    true,
			UserOnline: true,
			Error:      nil,
		}
	default:
		// Client's message buffer is full - this is an actual error
		log.Printf("Message buffer full for user %d, closing connection", userID)
		return SendResult{
			Success:    false,
			UserOnline: true, // They were online but connection is problematic
			Error:      fmt.Errorf("client message buffer is full"),
		}
	}
}

func (s *WebSocketService) broadcastUserStatus(userID int, status string) {
    statusEvent := websocketModels.Event{
        Type: websocketModels.EventUserStatus,
        Payload: mustMarshal(websocketModels.UserStatusEvent{
            UserID: userID,
            Status: status,
        }),
    }
    
    s.RLock()
    for client := range s.clients {
        if client.userID != userID {
            select {
            case client.egress <- statusEvent:
            default:
                log.Printf("Failed to send status update to user %d", client.userID)
            }
        }
    }
    s.RUnlock()
}