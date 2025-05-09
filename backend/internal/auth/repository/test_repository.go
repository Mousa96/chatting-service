package repository

import (
	"errors"
	"sync"
	"time"

	"github.com/Mousa96/chatting-service/internal/auth/models"
)

type Username string

type TestUserRepository struct {
	users  map[Username]*models.User
	mu     sync.RWMutex
	nextID int
}

func NewTestUserRepository() *TestUserRepository {
	return &TestUserRepository{
		users:  make(map[Username]*models.User),
		nextID: 1,
	}
}

func (r *TestUserRepository) Create(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	username := Username(user.Username)
	if _, exists := r.users[username]; exists {
		return errors.New("username already exists")
	}

	user.ID = r.nextID
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	r.users[username] = user
	r.nextID++

	return nil
}

func (r *TestUserRepository) GetByUsername(username string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[Username(username)]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}
