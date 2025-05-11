// Package repository provides data access interfaces and implementations for messages
package repository

import (
	"github.com/Mousa96/chatting-service/internal/message/models"
)

// Repository defines the message repository operations
type Repository interface {
	// Create stores a new message
	Create(message *models.Message) error
	
	// GetMessagesByUser retrieves all messages involving a user
	GetMessagesByUser(userID int) ([]models.Message, error)
	
	// GetConversation retrieves the conversation between two users
	GetConversation(userID1, userID2 int) ([]models.Message, error)
	
	// GetConversationPaginated retrieves the conversation with pagination
	GetConversationPaginated(userID1, userID2, page, pageSize int) ([]models.Message, *models.Pagination, error)
	
	// GetMessageHistory retrieves all messages for a user in chronological order
	GetMessageHistory(userID int) ([]models.Message, error)
	
	// GetMessageHistoryPaginated retrieves messages with pagination
	GetMessageHistoryPaginated(userID, page, pageSize int) ([]models.Message, *models.Pagination, error)
	
	// UpdateMessageStatus updates the status of a message
	UpdateMessageStatus(messageID int, status models.MessageStatus) error
	
	// GetMessageByID retrieves a message by its ID
	GetMessageByID(messageID int) (*models.Message, error)
}
