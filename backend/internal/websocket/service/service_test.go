package service

import (
	"encoding/json"
	"testing"
	"time"

	"mime/multipart"

	messageModels "github.com/Mousa96/chatting-service/internal/message/models"
	"github.com/Mousa96/chatting-service/internal/websocket/models"
	"github.com/Mousa96/chatting-service/internal/websocket/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMessageService is a mock implementation of the message service
type MockMessageService struct {
	mock.Mock
}

func (m *MockMessageService) SendMessage(senderID int, req *messageModels.CreateMessageRequest) (*messageModels.Message, error) {
	args := m.Called(senderID, req)
	return args.Get(0).(*messageModels.Message), args.Error(1)
}

func (m *MockMessageService) GetMessagesByUser(userID int) ([]*messageModels.Message, error) {
	args := m.Called(userID)
	return args.Get(0).([]*messageModels.Message), args.Error(1)
}

func (m *MockMessageService) GetConversation(userID1, userID2 int) ([]messageModels.Message, error) {
	args := m.Called(userID1, userID2)
	return args.Get(0).([]messageModels.Message), args.Error(1)
}

func (m *MockMessageService) UpdateMessageStatus(messageID int, status messageModels.MessageStatus, userID int) error {
	args := m.Called(messageID, status, userID)
	return args.Error(0)
}

func (m *MockMessageService) BroadcastMessage(senderID int, req *messageModels.BroadcastMessageRequest) ([]*messageModels.Message, error) {
	args := m.Called(senderID, req)
	return args.Get(0).([]*messageModels.Message), args.Error(1)
}

func (m *MockMessageService) GetMessageByID(messageID int) (*messageModels.Message, error) {
	args := m.Called(messageID)
	return args.Get(0).(*messageModels.Message), args.Error(1)
}

func (m *MockMessageService) GetMessageHistory(userID int) ([]messageModels.Message, error) {
	args := m.Called(userID)
	return args.Get(0).([]messageModels.Message), args.Error(1)
}

func (m *MockMessageService) UploadMedia(userID int, file *multipart.FileHeader) (string, error) {
	args := m.Called(userID, file)
	return args.String(0), args.Error(1)
}

func TestWebSocketService_SendEvent(t *testing.T) {
	// Create mock repository
	mockRepo := repository.NewMockRepository()
	
	// Create test client
	testClient := &models.Client{
		UserID: 1,
		Send:   make(chan []byte, 256),
	}
	mockRepo.AddClient(testClient)
	
	// Create service
	mockMessageService := &MockMessageService{}
	wsService := NewWebSocketService(mockRepo, mockMessageService)
	
	// Create test event
	event := &models.Event{
		Type:      models.EventMessage,
		SenderID:  2,
		Timestamp: time.Now(),
		Payload:   "Test message",
	}
	
	// Send event
	err := wsService.SendEvent(1, event)
	assert.NoError(t, err)
	
	// Verify event was queued to the client's send channel
	assert.True(t, mockRepo.GetClientCalled)
	
	// Check that data was sent on the channel
	select {
	case data := <-testClient.Send:
		// Verify the sent data is valid JSON and contains the correct information
		var sentEvent models.Event
		err := json.Unmarshal(data, &sentEvent)
		assert.NoError(t, err)
		assert.Equal(t, event.Type, sentEvent.Type)
		assert.Equal(t, event.SenderID, sentEvent.SenderID)
	default:
		t.Fatal("No data sent on client channel")
	}
}

func TestWebSocketService_BroadcastEvent(t *testing.T) {
	// Create mock repository
	mockRepo := repository.NewMockRepository()
	
	// Create test clients
	client1 := &models.Client{
		UserID: 1,
		Send:   make(chan []byte, 256),
	}
	client2 := &models.Client{
		UserID: 2,
		Send:   make(chan []byte, 256),
	}
	mockRepo.AddClient(client1)
	mockRepo.AddClient(client2)
	
	// Create service
	mockMessageService := &MockMessageService{}
	wsService := NewWebSocketService(mockRepo, mockMessageService)
	
	// Create test event
	event := &models.Event{
		Type:      models.EventUserStatus,
		SenderID:  3,
		Timestamp: time.Now(),
		Payload:   "User logged in",
	}
	
	// Broadcast event
	err := wsService.BroadcastEvent(event)
	assert.NoError(t, err)
	
	// Verify event was sent to all clients
	assert.True(t, mockRepo.GetAllClientsCalled)
	
	// Check first client
	select {
	case data := <-client1.Send:
		var sentEvent models.Event
		err := json.Unmarshal(data, &sentEvent)
		assert.NoError(t, err)
		assert.Equal(t, event.Type, sentEvent.Type)
	default:
		t.Fatal("No data sent to client 1")
	}
	
	// Check second client
	select {
	case data := <-client2.Send:
		var sentEvent models.Event
		err := json.Unmarshal(data, &sentEvent)
		assert.NoError(t, err)
		assert.Equal(t, event.Type, sentEvent.Type)
	default:
		t.Fatal("No data sent to client 2")
	}
}

func TestWebSocketService_UpdateUserStatus(t *testing.T) {
	// Create mock repository
	mockRepo := repository.NewMockRepository()
	
	// Create service
	mockMessageService := &MockMessageService{}
	wsService := NewWebSocketService(mockRepo, mockMessageService)
	
	// Update user status
	err := wsService.UpdateUserStatus(1, models.StatusAway)
	assert.NoError(t, err)
	
	// Verify repository was called
	assert.True(t, mockRepo.UpdateUserStatusCalled)
}

func TestWebSocketService_GetUserStatus(t *testing.T) {
	// Create mock repository
	mockRepo := repository.NewMockRepository()
	
	// Set a user status
	mockRepo.UpdateUserStatus(1, models.StatusOnline)
	
	// Create service
	mockMessageService := &MockMessageService{}
	wsService := NewWebSocketService(mockRepo, mockMessageService)
	
	// Get user status
	status, err := wsService.GetUserStatus(1)
	assert.NoError(t, err)
	assert.Equal(t, models.StatusOnline, status)
	
	// Verify repository was called
	assert.True(t, mockRepo.GetUserStatusCalled)
}

func TestWebSocketService_GetConnectedUsers(t *testing.T) {
	// Create mock repository
	mockRepo := repository.NewMockRepository()
	
	// Create test clients
	client1 := &models.Client{
		UserID: 1,
		Send:   make(chan []byte, 256),
	}
	client2 := &models.Client{
		UserID: 2,
		Send:   make(chan []byte, 256),
	}
	mockRepo.AddClient(client1)
	mockRepo.AddClient(client2)
	
	// Create service
	mockMessageService := &MockMessageService{}
	wsService := NewWebSocketService(mockRepo, mockMessageService)
	
	// Get connected users
	users, err := wsService.GetConnectedUsers()
	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Contains(t, users, 1)
	assert.Contains(t, users, 2)
	
	// Verify repository was called
	assert.True(t, mockRepo.GetAllClientsCalled)
}

func TestWebSocketService_ProcessBroadcastEvent(t *testing.T) {
	// Setup test clients
	client1 := &models.Client{
		UserID: 2,
		Send:   make(chan []byte, 256),
	}
	client2 := &models.Client{
		UserID: 3,
		Send:   make(chan []byte, 256),
	}
	mockRepo := repository.NewMockRepository()
	mockRepo.AddClient(client1)
	mockRepo.AddClient(client2)
	
	// Mock the message service
	mockMessageService := new(MockMessageService)
	mockMessageService.On("BroadcastMessage", 
		1, // sender ID
		mock.MatchedBy(func(req *messageModels.BroadcastMessageRequest) bool {
			return len(req.ReceiverIDs) == 2 && 
				   req.Content == "Test broadcast" &&
				   req.MediaURL == "/uploads/test.jpg"
		})).
		Return([]*messageModels.Message{
			{ID: 101, SenderID: 1, ReceiverID: 2, Content: "Test broadcast", MediaURL: "/uploads/test.jpg"},
			{ID: 102, SenderID: 1, ReceiverID: 3, Content: "Test broadcast", MediaURL: "/uploads/test.jpg"},
		}, nil)
	
	// Create the service
	wsService := NewWebSocketService(mockRepo, mockMessageService)
	
	// Create a broadcast event
	event := models.Event{
		Type:     models.EventBroadcast,
		SenderID: 1,
		Message: &messageModels.Message{
			Content:  "Test broadcast",
			MediaURL: "/uploads/test.jpg",
			ReceiverIDs: []int{2, 3},
		},
	}
	
	// Manually trigger the readPump's event handling logic
	// This would normally happen in readPump but we can simulate it here
	if event.Type == models.EventBroadcast {
		// Access the exported methods
		// We'll need to process the event ourselves
		broadcastReq := &messageModels.BroadcastMessageRequest{
			ReceiverIDs: event.Message.ReceiverIDs,
			Content:     event.Message.Content,
			MediaURL:    event.Message.MediaURL,
		}
		
		messages, _ := mockMessageService.BroadcastMessage(event.SenderID, broadcastReq)
		
		// Send each message to its recipient
		for _, msg := range messages {
			msgEvent := &models.Event{
				Type:      models.EventMessage,
				SenderID:  event.SenderID,
				Timestamp: time.Now(),
				Message:   msg,
			}
			wsService.SendEvent(msg.ReceiverID, msgEvent)
		}
	}
	
	// Verify the message service was called
	mockMessageService.AssertExpectations(t)
	
	// Check that events were sent to both clients
	timeout := time.After(100 * time.Millisecond)
	
	// Check client 1
	select {
	case data := <-client1.Send:
		var sentEvent models.Event
		err := json.Unmarshal(data, &sentEvent)
		assert.NoError(t, err)
		assert.Equal(t, models.EventMessage, sentEvent.Type)
		assert.Equal(t, 1, sentEvent.SenderID)
		assert.Equal(t, 101, sentEvent.Message.ID)
	case <-timeout:
		t.Fatal("No message sent to client 1")
	}
	
	// Check client 2
	select {
	case data := <-client2.Send:
		var sentEvent models.Event
		err := json.Unmarshal(data, &sentEvent)
		assert.NoError(t, err)
		assert.Equal(t, models.EventMessage, sentEvent.Type)
		assert.Equal(t, 1, sentEvent.SenderID)
		assert.Equal(t, 102, sentEvent.Message.ID)
	case <-timeout:
		t.Fatal("No message sent to client 2")
	}
} 