package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"mime/multipart"

	"errors"
	"time"

	"github.com/Mousa96/chatting-service/internal/message/models"
	"github.com/Mousa96/chatting-service/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockService struct {
	mock.Mock
	mockError      error
	mockMessages   []models.Message
}

func (m *mockService) SendMessage(senderID int, req *models.CreateMessageRequest) (*models.Message, error) {
	args := m.Called(senderID, req)
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *mockService) GetConversation(userID, otherUserID int) ([]models.Message, error) {
	if m.mockError != nil {
		return []models.Message{}, m.mockError
	}
	return m.mockMessages, nil
}

func (m *mockService) UploadMedia(userID int, file *multipart.FileHeader) (string, error) {
	args := m.Called(userID, file)
	return args.String(0), args.Error(1)
}

func (m *mockService) BroadcastMessage(senderID int, req *models.BroadcastMessageRequest) ([]*models.Message, error) {
	args := m.Called(senderID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *mockService) GetMessageHistory(userID int) ([]models.Message, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Message), args.Error(1)
}

func (m *mockService) UpdateMessageStatus(messageID int, status models.MessageStatus, userID int) error {
	args := m.Called(messageID, status, userID)
	return args.Error(0)
}

func (m *mockService) GetMessageByID(messageID int) (*models.Message, error) {
	args := m.Called(messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *mockService) GetConversationPaginated(userID1, userID2, page, pageSize int) ([]models.Message, *models.Pagination, error) {
	args := m.Called(userID1, userID2, page, pageSize)
	return args.Get(0).([]models.Message), args.Get(1).(*models.Pagination), args.Error(2)
}

func (m *mockService) GetMessageHistoryPaginated(userID, page, pageSize int) ([]models.Message, *models.Pagination, error) {
	args := m.Called(userID, page, pageSize)
	return args.Get(0).([]models.Message), args.Get(1).(*models.Pagination), args.Error(2)
}

func TestSendMessage(t *testing.T) {
	tests := []struct {
		name string
		req  models.CreateMessageRequest
	}{
		{
			name: "Valid message",
			req: models.CreateMessageRequest{
				ReceiverID: 2,
				Content:    "Hello!",
			},
		},
		{
			name: "Message with media",
			req: models.CreateMessageRequest{
				ReceiverID: 2,
				Content:    "Check this out",
				MediaURL:   "http://example.com/image.jpg",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockService)
			handler := NewMessageHandler(mockService)

			// Create request with context containing user ID
			body, _ := json.Marshal(tt.req)
			req := httptest.NewRequest(http.MethodPost, "/api/messages", bytes.NewBuffer(body))
			
			// Add user ID to context
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
			req = req.WithContext(ctx)
			
			rr := httptest.NewRecorder()
			
			// Set up expectations
			mockService.On("SendMessage", 1, &tt.req).Return(&models.Message{
				ID:         1,
				SenderID:   1,
				ReceiverID: tt.req.ReceiverID,
				Content:    tt.req.Content,
				MediaURL:   tt.req.MediaURL,
			}, nil)

			handler.SendMessage(rr, req)
			
			assert.Equal(t, http.StatusOK, rr.Code)

			var response models.Message
			err := json.NewDecoder(rr.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, 1, response.SenderID)
			assert.Equal(t, tt.req.ReceiverID, response.ReceiverID)
			assert.Equal(t, tt.req.Content, response.Content)
			assert.Equal(t, tt.req.MediaURL, response.MediaURL)

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetConversation(t *testing.T) {
	mockService := new(mockService)
	handler := NewMessageHandler(mockService)

	t.Run("Success", func(t *testing.T) {
		mockMessages := []models.Message{
			{ID: 1, SenderID: 1, ReceiverID: 2, Content: "Hello", CreatedAt: time.Now()},
		}
		mockService.On("GetConversation", 1, 2).Return(mockMessages, nil).Once()
		
		req := httptest.NewRequest(http.MethodGet, "/api/messages/conversation?user_id=2", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
		req = req.WithContext(ctx)
		
		rr := httptest.NewRecorder()
		handler.GetConversation(rr, req)
		
		assert.Equal(t, http.StatusOK, rr.Code)
		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "messages")
	})

	t.Run("EmptyConversation", func(t *testing.T) {
		mockService.On("GetConversation", 1, 3).Return([]models.Message{}, nil).Once()
		
		req := httptest.NewRequest(http.MethodGet, "/api/messages/conversation?user_id=3", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
		req = req.WithContext(ctx)
		
		rr := httptest.NewRecorder()
		handler.GetConversation(rr, req)
		
		assert.Equal(t, http.StatusOK, rr.Code)
		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "messages")
		assert.Empty(t, response["messages"])
	})

	t.Run("ServerError", func(t *testing.T) {
		// Set the mock error for this test case
		mockService.mockError = errors.New("database error")
		
		req := httptest.NewRequest(http.MethodGet, "/api/messages/conversation?user_id=4", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
		req = req.WithContext(ctx)
		
		rr := httptest.NewRecorder()
		handler.GetConversation(rr, req)
		
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestUploadMedia(t *testing.T) {
	tests := []struct {
		name         string
		fileContent  []byte
		filename     string
		contentType  string
		expectedCode int
	}{
		{
			name:         "Valid image upload",
			fileContent:  []byte{0xFF, 0xD8, 0xFF, 0xE0}, // JPEG header
			filename:     "test.jpg",
			contentType:  "image/jpeg",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid file type",
			fileContent:  []byte("text content"),
			filename:     "test.txt",
			contentType:  "text/plain",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Empty file",
			fileContent:  []byte{},
			filename:     "empty.jpg",
			contentType:  "image/jpeg",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockService)
			handler := NewMessageHandler(mockService)

			// Create multipart form data
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, err := writer.CreateFormFile("file", tt.filename) // Changed "media" to "file"
			require.NoError(t, err)
			_, err = part.Write(tt.fileContent)
			require.NoError(t, err)
			err = writer.Close()
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/messages/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
			req = req.WithContext(ctx)

			if tt.expectedCode == http.StatusOK {
				mockService.On("UploadMedia", 1, mock.AnythingOfType("*multipart.FileHeader")).
					Return("https://example.com/"+tt.filename, nil)
			}

			rr := httptest.NewRecorder()
			handler.UploadMedia(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedCode == http.StatusOK {
				var response struct {
					URL string `json:"url"`
				}
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Contains(t, response.URL, tt.filename)
			}
		})
	}
}

func TestBroadcastMessage(t *testing.T) {
	tests := []struct {
		name         string
		userID      int
		request     models.BroadcastMessageRequest
		setupMock   func(*mockService)
		expectedCode int
	}{
		{
			name:    "Valid broadcast",
			userID: 1,
			request: models.BroadcastMessageRequest{
				ReceiverIDs: []int{2, 3, 4},
				Content:    "Hello everyone!",
			},
			setupMock: func(ms *mockService) {
				ms.On("BroadcastMessage", 1, mock.MatchedBy(func(req *models.BroadcastMessageRequest) bool {
					return len(req.ReceiverIDs) == 3 && req.Content == "Hello everyone!"
				})).Return([]*models.Message{
					{ID: 1, SenderID: 1, ReceiverID: 2, Content: "Hello everyone!"},
					{ID: 2, SenderID: 1, ReceiverID: 3, Content: "Hello everyone!"},
					{ID: 3, SenderID: 1, ReceiverID: 4, Content: "Hello everyone!"},
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:    "Empty receivers",
			userID: 1,
			request: models.BroadcastMessageRequest{
				ReceiverIDs: []int{},
				Content:    "Hello!",
			},
			setupMock: func(ms *mockService) {
				ms.On("BroadcastMessage", 1, mock.Anything).
					Return(nil, fmt.Errorf("receiver IDs cannot be empty"))
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockService)
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}
			handler := NewMessageHandler(mockService)

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/api/messages/broadcast", bytes.NewBuffer(body))
			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))
			
			rr := httptest.NewRecorder()
			handler.BroadcastMessage(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetMessageHistory(t *testing.T) {
	mockService := new(mockService)
	handler := NewMessageHandler(mockService)

	// Test case: Valid request
	t.Run("Valid_request", func(t *testing.T) {
		// Set up mock messages
		mockMessages := []models.Message{
			{ID: 1, SenderID: 1, ReceiverID: 2, Content: "Hello", CreatedAt: time.Now()},
		}
		mockPagination := &models.Pagination{
			CurrentPage: 1,
			PageSize:    10,
			TotalItems:  1,
			TotalPages:  1,
		}
		
		// Set up mock service expectation
		mockService.On("GetMessageHistoryPaginated", 1, 1, 10).Return(mockMessages, mockPagination, nil).Once()
		
		// Create request with context containing userID
		req := httptest.NewRequest(http.MethodGet, "/api/messages/history", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1) // This is the critical line
		req = req.WithContext(ctx)
		
		// Execute request
		rr := httptest.NewRecorder()
		handler.GetMessageHistory(rr, req)
		
		// Assert status code
		assert.Equal(t, http.StatusOK, rr.Code)
		
		// Verify JSON response
		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		// Verify response structure
		assert.Contains(t, response, "messages")
		assert.Contains(t, response, "pagination")
		
		// Verify service was called correctly
		mockService.AssertExpectations(t)
	})

	// Test case: Empty history
	t.Run("Empty_history", func(t *testing.T) {
		// Empty messages, but with valid pagination
		mockPagination := &models.Pagination{
			CurrentPage: 1,
			PageSize:    10,
			TotalItems:  0,
			TotalPages:  0,
		}
		
		// Set up mock service expectation
		mockService.On("GetMessageHistoryPaginated", 1, 1, 10).Return([]models.Message{}, mockPagination, nil).Once()
		
		// Create request with context containing userID
		req := httptest.NewRequest(http.MethodGet, "/api/messages/history", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1) // This is the critical line
		req = req.WithContext(ctx)
		
		// Execute request
		rr := httptest.NewRecorder()
		handler.GetMessageHistory(rr, req)
		
		// Assert status code
		assert.Equal(t, http.StatusOK, rr.Code)
		
		// Verify JSON response
		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		// Verify response structure
		assert.Contains(t, response, "messages")
		assert.Contains(t, response, "pagination")
		assert.Empty(t, response["messages"])
		
		// Verify service was called correctly
		mockService.AssertExpectations(t)
	})
}
