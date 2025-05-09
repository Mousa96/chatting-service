package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Mousa96/chatting-service/internal/message/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockService struct {
	mock.Mock
}

func (m *mockService) SendMessage(senderID int, req *models.CreateMessageRequest) (*models.Message, error) {
	args := m.Called(senderID, req)
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *mockService) GetConversation(userID1, userID2 int) ([]models.Message, error) {
	args := m.Called(userID1, userID2)
	return args.Get(0).([]models.Message), args.Error(1)
}

func TestSendMessage(t *testing.T) {
	mockSvc := new(mockService)
	handler := NewMessageHandler(mockSvc)

	tests := []struct {
		name         string
		userID       int
		request      models.CreateMessageRequest
		expectedMsg  *models.Message
		expectedCode int
	}{
		{
			name:   "Valid message",
			userID: 1,
			request: models.CreateMessageRequest{
				ReceiverID: 2,
				Content:    "Hello!",
			},
			expectedMsg: &models.Message{
				SenderID:   1,
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
			expectedMsg: &models.Message{
				SenderID:   1,
				ReceiverID: 2,
				Content:    "Check this out",
				MediaURL:   "http://example.com/image.jpg",
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectation
			mockSvc.On("SendMessage", tt.userID, &tt.request).Return(tt.expectedMsg, nil)

			// Create request with body
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/messages", bytes.NewBuffer(body))

			// Add userID to context using the same key as middleware
			ctx := context.WithValue(req.Context(), userIDContextKey, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.SendMessage(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedCode == http.StatusOK {
				var response models.Message
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMsg.SenderID, response.SenderID)
				assert.Equal(t, tt.expectedMsg.ReceiverID, response.ReceiverID)
				assert.Equal(t, tt.expectedMsg.Content, response.Content)
				assert.Equal(t, tt.expectedMsg.MediaURL, response.MediaURL)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestGetConversation(t *testing.T) {
	mockSvc := new(mockService)
	handler := NewMessageHandler(mockSvc)

	messages := []models.Message{
		{SenderID: 1, ReceiverID: 2, Content: "Hello"},
		{SenderID: 2, ReceiverID: 1, Content: "Hi"},
	}

	tests := []struct {
		name         string
		userID       int
		otherUserID  string
		expectedID   int
		messages     []models.Message
		expectedCode int
	}{
		{
			name:         "Valid conversation",
			userID:       1,
			otherUserID:  "2",
			expectedID:   2,
			messages:     messages,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid user ID",
			userID:       1,
			otherUserID:  "invalid",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Empty conversation",
			userID:       1,
			otherUserID:  "3",
			expectedID:   3,
			messages:     []models.Message{},
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedCode == http.StatusOK {
				mockSvc.On("GetConversation", tt.userID, tt.expectedID).Return(tt.messages, nil)
			}

			req := httptest.NewRequest(http.MethodGet, "/messages/conversation?user_id="+tt.otherUserID, nil)
			ctx := context.WithValue(req.Context(), userIDContextKey, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.GetConversation(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedCode == http.StatusOK {
				var response []models.Message
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.messages, response)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}
