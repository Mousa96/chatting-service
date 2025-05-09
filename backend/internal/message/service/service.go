// Package service implements the messaging business logic
package service

import (
	"github.com/Mousa96/chatting-service/internal/message/models"
	"github.com/Mousa96/chatting-service/internal/message/repository"
)

// MessageService provides the implementation of the Service interface
type MessageService struct {
	messageRepo repository.Repository
}

// NewMessageService creates a new MessageService instance
func NewMessageService(messageRepo repository.Repository) Service {
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
