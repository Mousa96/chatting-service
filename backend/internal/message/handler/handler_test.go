package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Mousa96/chatting-service/internal/message/models"
	"github.com/Mousa96/chatting-service/internal/message/repository"
	"github.com/Mousa96/chatting-service/internal/message/service"
	"github.com/stretchr/testify/assert"
)

type contextKey string
const userIDKey = contextKey("user_id")

func TestSendMessage(t *testing.T) {
	repo := repository.NewTestMessageRepository()
	messageService := service.NewMessageService(repo)
	handler := NewMessageHandler(messageService)

	tests := []struct {
		name         string
		userID       int
		request      models.CreateMessageRequest
		expectedCode int
	}{
		{
			name:   "Valid message",
			userID: 1,
			request: models.CreateMessageRequest{
				ReceiverID: 2,
				Content:    "Hello!",
			},
			expectedCode: http.StatusOK,
		},
		{
			name:   "Message with media",
			userID: 1,
			request: models.CreateMessageRequest{
				ReceiverID: 2,
				Content:    "Check this out",
				MediaURL:   "http://example.com/image.jpg",
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/api/messages", bytes.NewBuffer(body))

			// Add user_id to context
			ctx := context.WithValue(req.Context(), userIDKey, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.SendMessage(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedCode == http.StatusOK {
				var response models.Message
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.userID, response.SenderID)
				assert.Equal(t, tt.request.ReceiverID, response.ReceiverID)
				assert.Equal(t, tt.request.Content, response.Content)
				assert.Equal(t, tt.request.MediaURL, response.MediaURL)
			}
		})
	}
}

func TestGetConversation(t *testing.T) {
	repo := repository.NewTestMessageRepository()
	messageService := service.NewMessageService(repo)
	handler := NewMessageHandler(messageService)

	// Setup test messages
	messages := []models.Message{
		{
			SenderID:   1,
			ReceiverID: 2,
			Content:    "Hello from 1",
		},
		{
			SenderID:   2,
			ReceiverID: 1,
			Content:    "Hi from 2",
		},
	}

	for _, msg := range messages {
		if err := repo.Create(&msg); err != nil {
			t.Fatalf("Failed to create test message: %v", err)
		}
	}

	tests := []struct {
		name         string
		userID       int
		otherUserID  string
		expectedCode int
		expectedMsgs int
	}{
		{
			name:         "Valid conversation",
			userID:       1,
			otherUserID:  "2",
			expectedCode: http.StatusOK,
			expectedMsgs: 2,
		},
		{
			name:         "Invalid user ID",
			userID:       1,
			otherUserID:  "invalid",
			expectedCode: http.StatusBadRequest,
			expectedMsgs: 0,
		},
		{
			name:         "Empty conversation",
			userID:       1,
			otherUserID:  "3",
			expectedCode: http.StatusOK,
			expectedMsgs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/messages/conversation?user_id="+tt.otherUserID, nil)

			// Add user_id to context
			ctx := context.WithValue(req.Context(), userIDKey, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.GetConversation(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedCode == http.StatusOK {
				var response []models.Message
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Len(t, response, tt.expectedMsgs)
			}
		})
	}
}
