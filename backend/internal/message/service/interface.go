package service

import "github.com/Mousa96/chatting-service/internal/message/models"

type Service interface {
    SendMessage(senderID int, req *models.CreateMessageRequest) (*models.Message, error)
    GetConversation(userID1, userID2 int) ([]models.Message, error)
} 