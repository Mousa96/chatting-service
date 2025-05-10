// Package service provides the business logic for WebSocket operations
package service

import (
	messageModels "github.com/Mousa96/chatting-service/internal/message/models"
	"github.com/Mousa96/chatting-service/internal/websocket/models"
	"github.com/gorilla/websocket"
)

// Service defines the WebSocket service interface
type Service interface {
	// HandleConnection handles a new WebSocket connection
	HandleConnection(conn *websocket.Conn, userID int) error
	
	// CloseConnection closes a user's WebSocket connection
	CloseConnection(userID int) error
	
	// SendEvent sends an event to a specific user
	SendEvent(receiverID int, event *models.Event) error
	
	// BroadcastEvent sends an event to all connected users
	BroadcastEvent(event *models.Event) error
	
	// NotifyMessageSent notifies a user about a new message
	NotifyMessageSent(message *messageModels.Message) error
	
	// NotifyStatusChange notifies about message status changes (read/delivered)
	NotifyStatusChange(messageID int, status messageModels.MessageStatus) error
	
	// NotifyTypingStatus sends typing notifications
	NotifyTypingStatus(senderID, receiverID int, isTyping bool) error
	
	// UpdateUserStatus changes a user's online status
	UpdateUserStatus(userID int, status models.UserStatus) error
	
	// GetUserStatus gets a user's online status
	GetUserStatus(userID int) (models.UserStatus, error)
	
	// GetConnectedUsers gets all connected user IDs
	GetConnectedUsers() ([]int, error)
}
