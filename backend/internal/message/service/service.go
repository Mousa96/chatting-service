// Package service implements the messaging business logic
package service

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/Mousa96/chatting-service/internal/message/models"
	"github.com/Mousa96/chatting-service/internal/message/repository"
	"github.com/Mousa96/chatting-service/internal/storage"
)

// MessageService provides the implementation of the Service interface
type MessageService struct {
	messageRepo repository.Repository
	storage     storage.Storage
}

// NewMessageService creates a new MessageService instance
func NewMessageService(messageRepo repository.Repository, storage storage.Storage) Service {
	return &MessageService{
		messageRepo: messageRepo,
		storage:     storage,
	}
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

func (s *MessageService) UploadMedia(userID int, file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)

	// Upload using storage interface
	url, err := s.storage.Upload(filename, src, file.Header.Get("Content-Type"))
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return url, nil
}

func (s *MessageService) BroadcastMessage(senderID int, req *models.BroadcastMessageRequest) ([]*models.Message, error) {
	if len(req.ReceiverIDs) == 0 {
		return nil, fmt.Errorf("receiver IDs cannot be empty")
	}
	if req.Content == "" {
		return nil, fmt.Errorf("message content cannot be empty")
	}

	var messages []*models.Message

	// Create a message for each receiver
	for _, receiverID := range req.ReceiverIDs {
		msg := &models.Message{
			SenderID:   senderID,
			ReceiverID: receiverID,
			Content:    req.Content,
			MediaURL:   req.MediaURL,
		}

		if err := s.messageRepo.Create(msg); err != nil {
			return nil, fmt.Errorf("failed to send message to user %d: %w", receiverID, err)
		}

		messages = append(messages, msg)
	}

	return messages, nil
}
