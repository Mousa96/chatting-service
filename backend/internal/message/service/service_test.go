package service

import (
	"io"
	"testing"

	"github.com/Mousa96/chatting-service/internal/message/models"
	"github.com/Mousa96/chatting-service/internal/message/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendMessage(t *testing.T) {
	repo := repository.NewTestMessageRepository()
	mockStorage := new(mockStorage)
	messageService := NewMessageService(repo, mockStorage)

	tests := []struct {
		name        string
		senderID    int
		request     *models.CreateMessageRequest
		expectedErr bool
	}{
		{
			name:     "Valid message",
			senderID: 1,
			request: &models.CreateMessageRequest{
				ReceiverID: 2,
				Content:    "Hello!",
			},
			expectedErr: false,
		},
		{
			name:     "Message with media",
			senderID: 1,
			request: &models.CreateMessageRequest{
				ReceiverID: 2,
				Content:    "Check this out",
				MediaURL:   "http://example.com/image.jpg",
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := messageService.SendMessage(tt.senderID, tt.request)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, msg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, msg)
				assert.Equal(t, tt.senderID, msg.SenderID)
				assert.Equal(t, tt.request.ReceiverID, msg.ReceiverID)
				assert.Equal(t, tt.request.Content, msg.Content)
				assert.Equal(t, tt.request.MediaURL, msg.MediaURL)
				assert.NotZero(t, msg.ID)
				assert.NotZero(t, msg.CreatedAt)
			}
		})
	}
}

func TestGetConversation(t *testing.T) {
	repo := repository.NewTestMessageRepository()
	mockStorage := new(mockStorage)
	messageService := NewMessageService(repo, mockStorage)

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
		{
			SenderID:   1,
			ReceiverID: 3,
			Content:    "Hello 3",
		},
	}

	for _, msg := range messages {
		if err := repo.Create(&msg); err != nil {
			t.Fatalf("Failed to create test message: %v", err)
		}
	}

	tests := []struct {
		name        string
		userID1     int
		userID2     int
		expectedLen int
	}{
		{
			name:        "Get conversation between user 1 and 2",
			userID1:     1,
			userID2:     2,
			expectedLen: 2,
		},
		{
			name:        "Get conversation between user 1 and 3",
			userID1:     1,
			userID2:     3,
			expectedLen: 1,
		},
		{
			name:        "Get empty conversation",
			userID1:     2,
			userID2:     3,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := messageService.GetConversation(tt.userID1, tt.userID2)

			assert.NoError(t, err)
			assert.Len(t, messages, tt.expectedLen)

			for _, msg := range messages {
				assert.True(t,
					(msg.SenderID == tt.userID1 && msg.ReceiverID == tt.userID2) ||
						(msg.SenderID == tt.userID2 && msg.ReceiverID == tt.userID1),
				)
			}
		})
	}
}

func TestBroadcastMessage(t *testing.T) {
	repo := repository.NewTestMessageRepository()
	mockStorage := new(mockStorage)
	messageService := NewMessageService(repo, mockStorage)

	tests := []struct {
		name        string
		senderID    int
		req         *models.BroadcastMessageRequest
		wantErr     bool
		expectedLen int
	}{
		{
			name:     "Valid broadcast",
			senderID: 1,
			req: &models.BroadcastMessageRequest{
				ReceiverIDs: []int{2, 3, 4},
				Content:     "Hello everyone!",
			},
			wantErr:     false,
			expectedLen: 3,
		},
		{
			name:     "Empty receivers",
			senderID: 1,
			req: &models.BroadcastMessageRequest{
				ReceiverIDs: []int{},
				Content:     "Hello!",
			},
			wantErr: true,
		},
		{
			name:     "Empty content",
			senderID: 1,
			req: &models.BroadcastMessageRequest{
				ReceiverIDs: []int{2, 3},
				Content:     "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := messageService.BroadcastMessage(tt.senderID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, messages, tt.expectedLen)

			for i, msg := range messages {
				assert.Equal(t, tt.senderID, msg.SenderID)
				assert.Equal(t, tt.req.ReceiverIDs[i], msg.ReceiverID)
				assert.Equal(t, tt.req.Content, msg.Content)
				assert.Equal(t, tt.req.MediaURL, msg.MediaURL)
			}
		})
	}
}

// Add mock storage
type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) Upload(filename string, content io.Reader, contentType string) (string, error) {
	args := m.Called(filename, content, contentType)
	return args.String(0), args.Error(1)
}

func (m *mockStorage) Delete(filename string) error {
	args := m.Called(filename)
	return args.Error(0)
}
