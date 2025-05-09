package repository

import "github.com/Mousa96/chatting-service/internal/message/models"

type Repository interface {
    Create(msg *models.Message) error
    GetConversation(userID1, userID2 int) ([]models.Message, error)
} 