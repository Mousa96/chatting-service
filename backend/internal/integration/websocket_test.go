package integration

import (
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	messageModels "github.com/Mousa96/chatting-service/internal/message/models"
	"github.com/Mousa96/chatting-service/internal/middleware"
	"github.com/Mousa96/chatting-service/internal/websocket/handler"
	wsModels "github.com/Mousa96/chatting-service/internal/websocket/models"
	"github.com/Mousa96/chatting-service/internal/websocket/repository"
	"github.com/Mousa96/chatting-service/internal/websocket/service"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// MockMessageService for testing
type MockMessageService struct{}

func (m *MockMessageService) SendMessage(senderID int, req *messageModels.CreateMessageRequest) (*messageModels.Message, error) {
	return &messageModels.Message{
		SenderID: senderID,
		ReceiverID: req.ReceiverID,
		Content: req.Content,
	}, nil
}

func (m *MockMessageService) GetMessagesByUser(userID int) (interface{}, error) {
	return nil, nil
}

func (m *MockMessageService) GetConversation(userID1, userID2 int) ([]messageModels.Message, error) {
	return []messageModels.Message{}, nil
}

func (m *MockMessageService) UpdateMessageStatus(messageID int, status messageModels.MessageStatus, userID int) error {
	return nil
}

func (m *MockMessageService) NotifyMessageSent(message interface{}) error {
	return nil
}

func (m *MockMessageService) NotifyStatusChange(messageID int, status interface{}) error {
	return nil
}

func (m *MockMessageService) BroadcastMessage(senderID int, req *messageModels.BroadcastMessageRequest) ([]*messageModels.Message, error) {
	return []*messageModels.Message{}, nil
}

func (m *MockMessageService) UploadMedia(userID int, file *multipart.FileHeader) (string, error) {
	return "http://example.com/test-media.jpg", nil
}

func (m *MockMessageService) GetMessageByID(messageID int) (*messageModels.Message, error) {
	return &messageModels.Message{
		ID: messageID,
		SenderID: 1,
		ReceiverID: 2,
		Content: "Test message",
	}, nil
}

func (m *MockMessageService) GetMessageHistory(userID int) ([]messageModels.Message, error) {
	return []messageModels.Message{}, nil
}

// Helper to create a test server with WebSocket support
func setupWebSocketTestServer(t *testing.T) *httptest.Server {
	// Create repository
	wsRepo := repository.NewMemoryRepository()
	
	// Create service with mock message service
	mockMsgService := &MockMessageService{}
	wsSvc := service.NewWebSocketService(wsRepo, mockMsgService)
	
	// Create handler
	wsHandler := handler.NewWebSocketHandler(wsSvc)
	
	// Create test handler that adds a fake auth context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Fake authentication by embedding user ID 1 in a context
		ctx := r.Context()
		ctx = context.WithValue(ctx, middleware.UserIDKey, 1)
		
		// Call the actual handler with modified context
		wsHandler.HandleConnection(w, r.WithContext(ctx))
	})
	
	// Create test server
	return httptest.NewServer(testHandler)
}

func TestWebSocketConnection(t *testing.T) {
	// Set up test server
	server := setupWebSocketTestServer(t)
	defer server.Close()
	
	// Convert http URL to ws URL
	wsURL := "ws" + server.URL[4:]
	
	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer conn.Close()
	
	// Send a test message
	testMessage := map[string]interface{}{
		"type":        "message",
		"receiver_id": 2,
		"content":     "Test message",
	}
	err = conn.WriteJSON(testMessage)
	assert.NoError(t, err)
	
	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(time.Second))
	
	// We should receive a ping or some response
	messageType, _, err := conn.ReadMessage()
	if err != nil {
		t.Logf("Read error (may be expected): %v", err)
	} else {
		assert.True(t, messageType == websocket.TextMessage || messageType == websocket.PingMessage)
	}
}

func TestWebSocketUserStatus(t *testing.T) {
	// Create repository
	wsRepo := repository.NewMemoryRepository()
	
	// Create service with mock message service
	mockMsgService := &MockMessageService{}
	wsSvc := service.NewWebSocketService(wsRepo, mockMsgService)
	
	// Create handler
	wsHandler := handler.NewWebSocketHandler(wsSvc)
	
	// Create a test request for the GetUserStatus endpoint
	req := httptest.NewRequest("GET", "/ws/status?user_id=1", nil)
	
	// Add user ID to context (simulate auth middleware)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 2)
	req = req.WithContext(ctx)
	
	// Create a response recorder
	rr := httptest.NewRecorder()
	
	// Call the handler
	wsHandler.GetUserStatus(rr, req)
	
	// Check the status code
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestWebSocketConnectedUsers(t *testing.T) {
	// Create repository
	wsRepo := repository.NewMemoryRepository()
	
	// Add a test client to the repository
	client := &wsModels.Client{
		UserID: 1,
		Send:   make(chan []byte, 256),
	}
	wsRepo.AddClient(client)
	
	// Create service with mock message service
	mockMsgService := &MockMessageService{}
	wsSvc := service.NewWebSocketService(wsRepo, mockMsgService)
	
	// Create handler
	wsHandler := handler.NewWebSocketHandler(wsSvc)
	
	// Create a test request for the GetConnectedUsers endpoint
	req := httptest.NewRequest("GET", "/ws/users", nil)
	
	// Add user ID to context (simulate auth middleware)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 2)
	req = req.WithContext(ctx)
	
	// Create a response recorder
	rr := httptest.NewRecorder()
	
	// Call the handler
	wsHandler.GetConnectedUsers(rr, req)
	
	// Check the status code
	assert.Equal(t, http.StatusOK, rr.Code)
} 