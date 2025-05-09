package service

import (
	"testing"

	"github.com/Mousa96/chatting-service/internal/auth/models"
	"github.com/Mousa96/chatting-service/internal/auth/repository"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	repo := repository.NewTestUserRepository()
	jwtKey := []byte("test-key")
	authService := NewAuthService(repo, jwtKey)

	tests := []struct {
		name        string
		request     *models.CreateUserRequest
		expectedErr bool
	}{
		{
			name: "Valid registration",
			request: &models.CreateUserRequest{
				Username: "testuser",
				Password: "testpass123",
			},
			expectedErr: false,
		},
		{
			name: "Duplicate username",
			request: &models.CreateUserRequest{
				Username: "testuser",
				Password: "anotherpass",
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := authService.Register(tt.request)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotEmpty(t, response.Token)
				assert.Equal(t, tt.request.Username, response.User.Username)
				assert.NotZero(t, response.User.ID)
				assert.NotEmpty(t, response.User.PasswordHash)
				assert.NotEqual(t, tt.request.Password, response.User.PasswordHash)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	repo := repository.NewTestUserRepository()
	jwtKey := []byte("test-key")
	authService := NewAuthService(repo, jwtKey)

	// Create a test user first
	validUser := &models.CreateUserRequest{
		Username: "testuser",
		Password: "testpass123",
	}
	if _, err := authService.Register(validUser); err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	tests := []struct {
		name        string
		request     *models.LoginRequest
		expectedErr bool
	}{
		{
			name: "Valid login",
			request: &models.LoginRequest{
				Username: "testuser",
				Password: "testpass123",
			},
			expectedErr: false,
		},
		{
			name: "Wrong password",
			request: &models.LoginRequest{
				Username: "testuser",
				Password: "wrongpass",
			},
			expectedErr: true,
		},
		{
			name: "User not found",
			request: &models.LoginRequest{
				Username: "nonexistent",
				Password: "testpass123",
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := authService.Login(tt.request)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotEmpty(t, response.Token)
				assert.Equal(t, tt.request.Username, response.User.Username)
			}
		})
	}
}
