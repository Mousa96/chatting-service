// Package service provides the business logic for message operations
package service

import "github.com/Mousa96/chatting-service/internal/message/models"

// Service defines the message operations interface
type Service interface {
	// SendMessage creates and sends a new message
	SendMessage(senderID int, req *models.CreateMessageRequest) (*models.Message, error)
	// GetConversation retrieves the conversation history between two users
	GetConversation(userID1, userID2 int) ([]models.Message, error)
}
