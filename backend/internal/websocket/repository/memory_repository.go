// Package repository implements the WebSocket repository interface
package repository

import (
	"fmt"
	"sync"

	"github.com/Mousa96/chatting-service/internal/websocket/models"
)

// MemoryRepository provides an in-memory implementation of Repository
type MemoryRepository struct {
	clients     map[int]*models.Client
	userStatus  map[int]models.UserStatus
	mu          sync.RWMutex
}

// NewMemoryRepository creates a new MemoryRepository instance
func NewMemoryRepository() Repository {
	return &MemoryRepository{
		clients:    make(map[int]*models.Client),
		userStatus: make(map[int]models.UserStatus),
	}
}

// AddClient adds a new WebSocket client
func (r *MemoryRepository) AddClient(client *models.Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clients[client.UserID] = client
	r.userStatus[client.UserID] = models.StatusOnline

	return nil
}

// RemoveClient removes a WebSocket client
func (r *MemoryRepository) RemoveClient(userID int) error {
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
func (r *MemoryRepository) GetClient(userID int) (*models.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	client, exists := r.clients[userID]
	if !exists {
		return nil, fmt.Errorf("client not found: user ID %d", userID)
	}

	return client, nil
}

// GetAllClients retrieves all active WebSocket clients
func (r *MemoryRepository) GetAllClients() ([]*models.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clients := make([]*models.Client, 0, len(r.clients))
	for _, client := range r.clients {
		clients = append(clients, client)
	}

	return clients, nil
}

// UpdateUserStatus updates a user's online status
func (r *MemoryRepository) UpdateUserStatus(userID int, status models.UserStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.userStatus[userID] = status
	return nil
}

// GetUserStatus gets a user's current online status
func (r *MemoryRepository) GetUserStatus(userID int) (models.UserStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	status, exists := r.userStatus[userID]
	if !exists {
		return models.StatusOffline, nil // Default to offline if not found
	}

	return status, nil
}
