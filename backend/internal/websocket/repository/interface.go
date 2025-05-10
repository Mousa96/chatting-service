// Package repository provides data access for WebSocket operations
package repository

import (
	"github.com/Mousa96/chatting-service/internal/websocket/models"
)

// Repository defines the WebSocket data access interface
type Repository interface {
	// AddClient adds a new WebSocket client
	AddClient(client *models.Client) error
	
	// RemoveClient removes a WebSocket client
	RemoveClient(userID int) error
	
	// GetClient retrieves a client by user ID
	GetClient(userID int) (*models.Client, error)
	
	// GetAllClients retrieves all active WebSocket clients
	GetAllClients() ([]*models.Client, error)
	
	// UpdateUserStatus updates a user's online status
	UpdateUserStatus(userID int, status models.UserStatus) error
	
	// GetUserStatus gets a user's current online status
	GetUserStatus(userID int) (models.UserStatus, error)
}
