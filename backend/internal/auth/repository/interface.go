// Package repository provides data access interfaces and implementations
package repository

import "github.com/Mousa96/chatting-service/internal/auth/models"

// Repository defines the data access interface for user operations
type Repository interface {
	// Create stores a new user in the database
	Create(user *models.User) error
	// GetByUsername retrieves a user by their username
	GetByUsername(username string) (*models.User, error)
}
