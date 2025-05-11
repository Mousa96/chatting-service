// Package repository provides test implementations of the Repository interface
package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/Mousa96/chatting-service/internal/message/models"
)

// TestMessageRepository provides an in-memory implementation of Repository for testing
type TestMessageRepository struct {
	messages map[int]*models.Message
	mu       sync.RWMutex
	nextID   int
}

// NewTestMessageRepository creates a new instance of TestMessageRepository
func NewTestMessageRepository() *TestMessageRepository {
	return &TestMessageRepository{
		messages: make(map[int]*models.Message),
		nextID:   1,
	}
}

func (r *TestMessageRepository) Create(msg *models.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	msg.ID = r.nextID
	msg.CreatedAt = time.Now()
	r.messages[msg.ID] = msg
	r.nextID++

	return nil
}

func (r *TestMessageRepository) GetConversation(userID1, userID2 int) ([]models.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var conversation []models.Message
	for _, msg := range r.messages {
		if (msg.SenderID == userID1 && msg.ReceiverID == userID2) ||
			(msg.SenderID == userID2 && msg.ReceiverID == userID1) {
			conversation = append(conversation, *msg)
		}
	}

	return conversation, nil
}

func (r *TestMessageRepository) GetMessageHistory(userID int) ([]models.Message, error) {
	var messages []models.Message
	for _, msg := range r.messages {
		if msg.SenderID == userID || msg.ReceiverID == userID {
			messages = append(messages, *msg)
		}
	}
	return messages, nil
}

func (r *TestMessageRepository) GetMessageByID(messageID int) (*models.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if msg, exists := r.messages[messageID]; exists {
		return msg, nil
	}
	return nil, fmt.Errorf("message not found")
}

func (r *TestMessageRepository) UpdateMessageStatus(messageID int, status models.MessageStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if msg, exists := r.messages[messageID]; exists {
		msg.Status = status
		msg.UpdatedAt = time.Now()
		return nil
	}
	return fmt.Errorf("message not found")
}

// GetConversationPaginated retrieves the conversation with pagination
func (r *TestMessageRepository) GetConversationPaginated(userID1, userID2, page, pageSize int) ([]models.Message, *models.Pagination, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// First get all messages in conversation
	var conversation []models.Message
	for _, msg := range r.messages {
		if (msg.SenderID == userID1 && msg.ReceiverID == userID2) ||
			(msg.SenderID == userID2 && msg.ReceiverID == userID1) {
			conversation = append(conversation, *msg)
		}
	}

	totalItems := len(conversation)
	pagination := models.NewPagination(page, pageSize, totalItems)
	
	// Apply pagination
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= totalItems {
		return []models.Message{}, pagination, nil
	}
	if end > totalItems {
		end = totalItems
	}
	
	// Return paginated slice
	if start < end {
		return conversation[start:end], pagination, nil
	}
	return []models.Message{}, pagination, nil
}

// GetMessageHistoryPaginated retrieves messages with pagination
func (r *TestMessageRepository) GetMessageHistoryPaginated(userID, page, pageSize int) ([]models.Message, *models.Pagination, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// First get all messages
	var messages []models.Message
	for _, msg := range r.messages {
		if msg.SenderID == userID || msg.ReceiverID == userID {
			messages = append(messages, *msg)
		}
	}

	totalItems := len(messages)
	pagination := models.NewPagination(page, pageSize, totalItems)
	
	// Apply pagination
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= totalItems {
		return []models.Message{}, pagination, nil
	}
	if end > totalItems {
		end = totalItems
	}
	
	// Return paginated slice
	if start < end {
		return messages[start:end], pagination, nil
	}
	return []models.Message{}, pagination, nil
}

// GetMessagesByUser retrieves all messages involving a user
func (r *TestMessageRepository) GetMessagesByUser(userID int) ([]models.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var messages []models.Message
	for _, msg := range r.messages {
		if msg.SenderID == userID || msg.ReceiverID == userID {
			messages = append(messages, *msg)
		}
	}
	return messages, nil
}
