// Package repository provides test implementations of the Repository interface
package repository

import (
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
