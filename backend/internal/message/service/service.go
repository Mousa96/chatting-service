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
	// Modified validation to properly handle media-only messages
	if req.Content == "" && req.MediaURL == "" {
		return nil, fmt.Errorf("message must have either content or media")
	}

	msg := &models.Message{
		SenderID:   senderID,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
		MediaURL:   req.MediaURL,
		CreatedAt:  time.Now(),
		Status:     models.StatusSent,
	}

	if err := s.messageRepo.Create(msg); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
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

	// Generate unique filename - remove the leading /uploads/
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
	
	// Modified validation to properly handle media-only messages
	if req.Content == "" && req.MediaURL == "" {
		return nil, fmt.Errorf("message must have either content or media")
	}

	var messages []*models.Message

	// Create a message for each receiver
	for _, receiverID := range req.ReceiverIDs {
		msg := &models.Message{
			SenderID:   senderID,
			ReceiverID: receiverID,
			Content:    req.Content,
			MediaURL:   req.MediaURL,
			CreatedAt:  time.Now(),
		}

		if err := s.messageRepo.Create(msg); err != nil {
			return nil, fmt.Errorf("failed to send message to user %d: %w", receiverID, err)
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

func (s *MessageService) GetMessageHistory(userID int) ([]models.Message, error) {
	return s.messageRepo.GetMessageHistory(userID)
}

func (s *MessageService) UpdateMessageStatus(messageID int, status models.MessageStatus, userID int) error {
	// Validate status first
	if !status.IsValid() {
		return fmt.Errorf("invalid status: %s", status)
	}

	// Verify message exists
	message, err := s.messageRepo.GetMessageByID(messageID)
	if err != nil {
		return fmt.Errorf("failed to find message: %w", err)
	}

	// Only the receiver should be able to update message status
	if message.ReceiverID != userID {
		return fmt.Errorf("not authorized to update this message status")
	}

	// Update the status
	if err := s.messageRepo.UpdateMessageStatus(messageID, status); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

// GetMessageByID retrieves a message by its ID
func (s *MessageService) GetMessageByID(messageID int) (*models.Message, error) {
	message, err := s.messageRepo.GetMessageByID(messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message by ID: %w", err)
	}
	return message, nil
}

// GetMessageHistoryPaginated retrieves message history with pagination
func (s *MessageService) GetMessageHistoryPaginated(userID, page, pageSize int) ([]models.Message, *models.Pagination, error) {
	return s.messageRepo.GetMessageHistoryPaginated(userID, page, pageSize)
}

// GetConversationPaginated retrieves conversation with pagination
func (s *MessageService) GetConversationPaginated(userID1, userID2, page, pageSize int) ([]models.Message, *models.Pagination, error) {
	return s.messageRepo.GetConversationPaginated(userID1, userID2, page, pageSize)
}
