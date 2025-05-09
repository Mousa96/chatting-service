// Package repository provides data access interfaces and implementations for messages
package repository

import "github.com/Mousa96/chatting-service/internal/message/models"

// Repository defines the data access interface for message operations
type Repository interface {
	// Create stores a new message in the database
	Create(msg *models.Message) error
	// GetConversation retrieves all messages between two users
	GetConversation(userID1, userID2 int) ([]models.Message, error)
}
