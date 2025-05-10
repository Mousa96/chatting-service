// Package service provides the business logic for message operations
package service

import (
	"mime/multipart"

	"github.com/Mousa96/chatting-service/internal/message/models"
)

// Service defines the message operations interface
type Service interface {
	// SendMessage creates and sends a new message
	SendMessage(senderID int, req *models.CreateMessageRequest) (*models.Message, error)
	// GetConversation retrieves the conversation history between two users
	GetConversation(userID1, userID2 int) ([]models.Message, error)
	// UploadMedia handles media upload
	UploadMedia(userID int, file *multipart.FileHeader) (string, error)
	// BroadcastMessage broadcasts a message to all users
	BroadcastMessage(senderID int, req *models.BroadcastMessageRequest) ([]*models.Message, error)
	// GetMessageHistory retrieves the message history for a user
	GetMessageHistory(userID int) ([]models.Message, error)
	// UpdateMessageStatus updates the status of a message
	UpdateMessageStatus(messageID int, status models.MessageStatus, userID int) error
	// GetMessageByID retrieves a message by its ID
	GetMessageByID(messageID int) (*models.Message, error)
}
