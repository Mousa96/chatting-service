// Package repository provides data access interfaces and implementations for messages
package repository

import "github.com/Mousa96/chatting-service/internal/message/models"

// Repository defines the data access interface for message operations
type Repository interface {
	// Create stores a new message in the database
	Create(msg *models.Message) error
	// GetConversation retrieves all messages between two users
	GetConversation(userID1, userID2 int) ([]models.Message, error)
	// GetMessageHistory retrieves the message history for a user
	GetMessageHistory(userID int) ([]models.Message, error)
	// GetMessageByID retrieves a message by its ID
	GetMessageByID(messageID int) (*models.Message, error)
	// UpdateMessageStatus updates the status of a message
	UpdateMessageStatus(messageID int, status models.MessageStatus) error
}
