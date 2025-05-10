package repository

import (
	"testing"

	"github.com/Mousa96/chatting-service/internal/websocket/models"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestMemoryRepository_AddClient(t *testing.T) {
	// Create a new repository
	repo := NewMemoryRepository()
	
	// Create a test client
	client := &models.Client{
		UserID:     1,
		Connection: &websocket.Conn{},
		Send:       make(chan []byte, 256),
	}
	
	// Test adding the client
	err := repo.AddClient(client)
	assert.NoError(t, err)
	
	// Verify the client was added correctly
	savedClient, err := repo.GetClient(1)
	assert.NoError(t, err)
	assert.Equal(t, client, savedClient)
	
	// Verify user status was set to online
	status, err := repo.GetUserStatus(1)
	assert.NoError(t, err)
	assert.Equal(t, models.StatusOnline, status)
}

func TestMemoryRepository_RemoveClient(t *testing.T) {
	// Create a new repository
	repo := NewMemoryRepository()
	
	// Create and add a test client
	client := &models.Client{
		UserID:     1,
		Connection: &websocket.Conn{},
		Send:       make(chan []byte, 256),
	}
	repo.AddClient(client)
	
	// Test removing the client
	err := repo.RemoveClient(1)
	assert.NoError(t, err)
	
	// Verify the client was removed
	_, err = repo.GetClient(1)
	assert.Error(t, err)
	
	// Verify user status was set to offline
	status, err := repo.GetUserStatus(1)
	assert.NoError(t, err)
	assert.Equal(t, models.StatusOffline, status)
}

func TestMemoryRepository_GetClient_NotFound(t *testing.T) {
	// Create a new repository
	repo := NewMemoryRepository()
	
	// Try to get a non-existent client
	_, err := repo.GetClient(999)
	assert.Error(t, err)
}

func TestMemoryRepository_UpdateUserStatus(t *testing.T) {
	// Create a new repository
	repo := NewMemoryRepository()
	
	// Create and add a test client
	client := &models.Client{
		UserID:     1,
		Connection: &websocket.Conn{},
		Send:       make(chan []byte, 256),
	}
	repo.AddClient(client)
	
	// Test updating user status
	err := repo.UpdateUserStatus(1, models.StatusAway)
	assert.NoError(t, err)
	
	// Verify the status was updated
	status, err := repo.GetUserStatus(1)
	assert.NoError(t, err)
	assert.Equal(t, models.StatusAway, status)
}

func TestMemoryRepository_GetAllClients(t *testing.T) {
	// Create a new repository
	repo := NewMemoryRepository()
	
	// Create and add multiple test clients
	client1 := &models.Client{
		UserID:     1,
		Connection: &websocket.Conn{},
		Send:       make(chan []byte, 256),
	}
	client2 := &models.Client{
		UserID:     2,
		Connection: &websocket.Conn{},
		Send:       make(chan []byte, 256),
	}
	
	repo.AddClient(client1)
	repo.AddClient(client2)
	
	// Get all clients
	clients, err := repo.GetAllClients()
	assert.NoError(t, err)
	assert.Len(t, clients, 2)
	
	// Verify the clients are correct
	foundClient1 := false
	foundClient2 := false
	
	for _, c := range clients {
		if c.UserID == 1 {
			foundClient1 = true
		}
		if c.UserID == 2 {
			foundClient2 = true
		}
	}
	
	assert.True(t, foundClient1)
	assert.True(t, foundClient2)
} 