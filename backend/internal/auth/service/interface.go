// Package service provides the business logic for authentication operations
package service

import "github.com/Mousa96/chatting-service/internal/auth/models"

// Service defines the authentication operations interface
type Service interface {
	// Register creates a new user account and returns an authentication token
	Register(req *models.CreateUserRequest) (*models.AuthResponse, error)
	// Login authenticates a user and returns an authentication token
	Login(req *models.LoginRequest) (*models.AuthResponse, error)
}
