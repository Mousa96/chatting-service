package repository

import "github.com/Mousa96/chatting-service/internal/auth/models"

type Repository interface {
    Create(user *models.User) error
    GetByUsername(username string) (*models.User, error)
} 