package service

import (
	"fmt"
	"io"
	"testing"

	"github.com/Mousa96/chatting-service/internal/message/models"
	"github.com/Mousa96/chatting-service/internal/message/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

type mockRepo struct {
	messages []models.Message
}

func (m *mockRepo) Create(msg *models.Message) error {
	m.messages = append(m.messages, *msg)
	return nil
}

func (m *mockRepo) GetConversation(userID1, userID2 int) ([]models.Message, error) {
	var result []models.Message
	for _, msg := range m.messages {
		if (msg.SenderID == userID1 && msg.ReceiverID == userID2) ||
			(msg.SenderID == userID2 && msg.ReceiverID == userID1) {
			result = append(result, msg)
		}
	}
	return result, nil
}

// Add missing methods
func (m *mockRepo) GetConversationPaginated(userID1, userID2, page, pageSize int) ([]models.Message, *models.Pagination, error) {
	messages, _ := m.GetConversation(userID1, userID2)
	return messages, &models.Pagination{
		CurrentPage: page,
		PageSize:    pageSize,
		TotalItems:  len(messages),
		TotalPages:  1,
	}, nil
}

func (m *mockRepo) GetMessageHistory(userID int) ([]models.Message, error) {
	var result []models.Message
	for _, msg := range m.messages {
		if msg.SenderID == userID || msg.ReceiverID == userID {
			result = append(result, msg)
		}
	}
	return result, nil
}

func (m *mockRepo) GetMessageHistoryPaginated(userID, page, pageSize int) ([]models.Message, *models.Pagination, error) {
	messages, _ := m.GetMessageHistory(userID)
	return messages, &models.Pagination{
		CurrentPage: page,
		PageSize:    pageSize,
		TotalItems:  len(messages),
		TotalPages:  1,
	}, nil
}

func (m *mockRepo) GetMessageByID(messageID int) (*models.Message, error) {
	for _, msg := range m.messages {
		if msg.ID == messageID {
			return &msg, nil
		}
	}
	return nil, fmt.Errorf("message not found")
}

func (m *mockRepo) UpdateMessageStatus(messageID int, status models.MessageStatus) error {
	for i := range m.messages {
		if m.messages[i].ID == messageID {
			m.messages[i].Status = status
			return nil
		}
	}
	return fmt.Errorf("message not found")
}

func (m *mockRepo) GetMessagesByUser(userID int) ([]models.Message, error) {
	var result []models.Message
	for _, msg := range m.messages {
		if msg.SenderID == userID || msg.ReceiverID == userID {
			result = append(result, msg)
		}
	}
	return result, nil
}

func TestGetConversation(t *testing.T) {
	repo := &mockRepo{}
	mockStorage := new(mockStorage)
	messageService := NewMessageService(repo, mockStorage)

	// Setup test messages
	messages := []models.Message{
		{ID: 1, SenderID: 1, ReceiverID: 2, Content: "Hello from 1"},
		{ID: 2, SenderID: 2, ReceiverID: 1, Content: "Hi from 2"},
		{ID: 3, SenderID: 1, ReceiverID: 3, Content: "Hello 3"},
	}

	for _, msg := range messages {
		err := repo.Create(&msg)
		require.NoError(t, err)
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

func TestGetMessageHistory(t *testing.T) {
	repo := &mockRepo{}
	mockStorage := new(mockStorage)
	messageService := NewMessageService(repo, mockStorage)

	// Setup test messages
	messages := []models.Message{
		{ID: 1, SenderID: 1, ReceiverID: 2, Content: "Message 1"},
		{ID: 2, SenderID: 2, ReceiverID: 1, Content: "Message 2"},
	}

	for _, msg := range messages {
		err := repo.Create(&msg)
		require.NoError(t, err)
	}

	tests := []struct {
		name        string
		userID      int
		expectedLen int
		wantErr     bool
	}{
		{
			name:        "Get user 1's messages",
			userID:      1,
			expectedLen: 2,
			wantErr:     false,
		},
		{
			name:        "Get user 2's messages",
			userID:      2,
			expectedLen: 2,
			wantErr:     false,
		},
		{
			name:        "Get non-existent user's messages",
			userID:      999,
			expectedLen: 0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := messageService.GetMessageHistory(tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, messages, tt.expectedLen)

			// Verify all messages involve the user
			for _, msg := range messages {
				assert.True(t, msg.SenderID == tt.userID || msg.ReceiverID == tt.userID)
			}
		})
	}
}

func TestUpdateMessageStatus(t *testing.T) {
	repo := repository.NewTestMessageRepository()
	mockStorage := new(mockStorage)
	messageService := NewMessageService(repo, mockStorage)

	// Create a test message
	msg := &models.Message{
		SenderID:   1,
		ReceiverID: 2,
		Content:    "Test message",
		Status:     models.StatusSent,
	}
	err := repo.Create(msg)
	require.NoError(t, err)

	tests := []struct {
		name          string
		messageID     int
		newStatus     models.MessageStatus
		userID        int
		expectedError bool
	}{
		{
			name:          "Valid status update to delivered",
			messageID:     msg.ID,
			newStatus:     models.StatusDelivered,
			userID:        2, // Receiver ID
			expectedError: false,
		},
		{
			name:          "Valid status update to read",
			messageID:     msg.ID,
			newStatus:     models.StatusRead,
			userID:        2, // Receiver ID
			expectedError: false,
		},
		{
			name:          "Invalid status value",
			messageID:     msg.ID,
			newStatus:     "invalid_status",
			userID:        2,
			expectedError: true,
		},
		{
			name:          "Unauthorized user",
			messageID:     msg.ID,
			newStatus:     models.StatusDelivered,
			userID:        3, // Not the receiver
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := messageService.UpdateMessageStatus(tt.messageID, tt.newStatus, tt.userID)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				msg, err := repo.GetMessageByID(tt.messageID)
				assert.NoError(t, err)
				assert.Equal(t, tt.newStatus, msg.Status)
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
