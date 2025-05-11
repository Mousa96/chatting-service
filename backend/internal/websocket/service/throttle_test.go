package service

import (
	"encoding/json"
	"testing"
	"time"

	messageModels "github.com/Mousa96/chatting-service/internal/message/models"
	"github.com/Mousa96/chatting-service/internal/websocket/models"
	"github.com/Mousa96/chatting-service/internal/websocket/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMessageThrottling(t *testing.T) {
	// Create mock repository
	mockRepo := repository.NewMemoryRepository()
	
	// Create test client
	testClient := &models.Client{
		UserID: 1,
		Send:   make(chan []byte, 256),
	}
	mockRepo.AddClient(testClient)
	
	// Create mock message service
	mockMessageService := new(MockMessageService)
	
	// Counter to track number of processed messages
	messageCount := 0
	
	// Configure the mock message service with a simpler approach
	// This properly mocks the SendMessage method with fixed return values
	mockMessageService.On("SendMessage", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			// Increment counter when the method is called
			messageCount++
		}).
		Return(&messageModels.Message{
			ID:         1,
			SenderID:   1,
			ReceiverID: 2,
			Content:    "Test message",
		}, nil)
	
	// Create service with throttling enabled (5 messages per second)
	wsService := NewWebSocketServiceWithThrottling(mockRepo, mockMessageService, 5, time.Second)
	
	// Send messages quickly
	const totalMessages = 10
	events := make([]*models.Event, totalMessages)
	
	for i := 0; i < totalMessages; i++ {
		events[i] = &models.Event{
			Type:     models.EventMessage,
			SenderID: 1,
			Message: &messageModels.Message{
				ReceiverID: 2,
				Content:    "Test message",
			},
		}
	}
	
	// Process events in quick succession
	for i, event := range events {
		// Convert to JSON as would happen in the readPump
		data, err := json.Marshal(event)
		assert.NoError(t, err)
		
		// Manually call handleRawMessage to simulate receiving the message
		throttled := wsService.handleRawMessage(1, data)
		
		// First 5 should not be throttled, the rest should be
		if i < 5 {
			assert.False(t, throttled, "Message %d should not be throttled", i+1)
		} else {
			assert.True(t, throttled, "Message %d should be throttled", i+1)
		}
	}
	
	// Verify that only 5 messages were processed
	assert.Equal(t, 5, messageCount, "Only 5 messages should be processed")
} 