package repository

import (
	"fmt"
	"sync"

	"github.com/Mousa96/chatting-service/internal/websocket/models"
)

// MockRepository is a mock implementation of the Repository interface for testing
type MockRepository struct {
	clients     map[int]*models.Client
	userStatus  map[int]models.UserStatus
	mu          sync.RWMutex
	
	// Track function calls for assertions
	AddClientCalled    bool
	RemoveClientCalled bool
	GetClientCalled    bool
	GetAllClientsCalled bool
	UpdateUserStatusCalled bool
	GetUserStatusCalled bool
}

// NewMockRepository creates a new MockRepository instance
func NewMockRepository() *MockRepository {
	return &MockRepository{
		clients:    make(map[int]*models.Client),
		userStatus: make(map[int]models.UserStatus),
	}
}

// AddClient adds a new WebSocket client
func (r *MockRepository) AddClient(client *models.Client) error {
	r.AddClientCalled = true
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clients[client.UserID] = client
	r.userStatus[client.UserID] = models.StatusOnline

	return nil
}

// RemoveClient removes a WebSocket client
func (r *MockRepository) RemoveClient(userID int) error {
	r.RemoveClientCalled = true
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.clients[userID]; !exists {
		return fmt.Errorf("client not found: user ID %d", userID)
	}

	delete(r.clients, userID)
	r.userStatus[userID] = models.StatusOffline

	return nil
}

// GetClient retrieves a client by user ID
func (r *MockRepository) GetClient(userID int) (*models.Client, error) {
	r.GetClientCalled = true
	r.mu.RLock()
	defer r.mu.RUnlock()

	client, exists := r.clients[userID]
	if !exists {
		return nil, fmt.Errorf("client not found: user ID %d", userID)
	}

	return client, nil
}

// GetAllClients retrieves all active WebSocket clients
func (r *MockRepository) GetAllClients() ([]*models.Client, error) {
	r.GetAllClientsCalled = true
	r.mu.RLock()
	defer r.mu.RUnlock()

	clients := make([]*models.Client, 0, len(r.clients))
	for _, client := range r.clients {
		clients = append(clients, client)
	}

	return clients, nil
}

// UpdateUserStatus updates a user's online status
func (r *MockRepository) UpdateUserStatus(userID int, status models.UserStatus) error {
	r.UpdateUserStatusCalled = true
	r.mu.Lock()
	defer r.mu.Unlock()

	r.userStatus[userID] = status
	return nil
}

// GetUserStatus gets a user's current online status
func (r *MockRepository) GetUserStatus(userID int) (models.UserStatus, error) {
	r.GetUserStatusCalled = true
	r.mu.RLock()
	defer r.mu.RUnlock()

	status, exists := r.userStatus[userID]
	if !exists {
		return models.StatusOffline, nil
	}

	return status, nil
} 