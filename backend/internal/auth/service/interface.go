package service

import "github.com/Mousa96/chatting-service/internal/auth/models"

type Service interface {
    Register(req *models.CreateUserRequest) (*models.AuthResponse, error)
    Login(req *models.LoginRequest) (*models.AuthResponse, error)
} 