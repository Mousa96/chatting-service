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

	"github.com/Mousa96/chatting-service/internal/message/models"
	"github.com/Mousa96/chatting-service/internal/middleware"
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

			req := httptest.NewRequest(http.MethodGet, "/api/messages/conversation?user_id="+tt.otherUserID, nil)
			
			// Add user ID to context
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
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

func TestUploadMedia(t *testing.T) {
	tests := []struct {
		name         string
		fileContent  []byte
		filename     string
		contentType  string
		expectedCode int
	}{
		{
			name: "Valid image upload",
			// JPEG file header magic bytes
			fileContent: []byte{
				0xFF, 0xD8, 0xFF, 0xE0, // JPEG SOI and APP0 marker
				0x00, 0x10, 0x4A, 0x46, // APP0 length and "JF"
				0x49, 0x46, 0x00, 0x01, // "IF" and version
				0x01, 0x02, 0x03, 0x04, // Some image data
			},
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

			// Create multipart form
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, err := writer.CreateFormFile("media", tt.filename)
			assert.NoError(t, err)
			_, err = part.Write(tt.fileContent)
			assert.NoError(t, err)
			writer.Close()

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/messages/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			
			// Add user ID to context
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
			req = req.WithContext(ctx)
			
			rr := httptest.NewRecorder()

			if tt.expectedCode == http.StatusOK {
				mockService.On("UploadMedia", 1, mock.AnythingOfType("*multipart.FileHeader")).
					Return("https://storage.example.com/"+tt.filename, nil)
			}

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

			mockService.AssertExpectations(t)
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
