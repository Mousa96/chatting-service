package service

import (
	"github.com/Mousa96/chatting-service/internal/models"
	"github.com/Mousa96/chatting-service/internal/repository"
)

type MessageService struct {
	messageRepo *repository.MessageRepository
}

func NewMessageService(messageRepo *repository.MessageRepository) *MessageService {
	return &MessageService{messageRepo: messageRepo}
}

func (s *MessageService) SendMessage(senderID int, req *models.CreateMessageRequest) (*models.Message, error) {
	msg := &models.Message{
		SenderID:   senderID,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
		MediaURL:   req.MediaURL,
	}

	if err := s.messageRepo.Create(msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *MessageService) GetConversation(userID1, userID2 int) ([]models.Message, error) {
	return s.messageRepo.GetConversation(userID1, userID2)
} 