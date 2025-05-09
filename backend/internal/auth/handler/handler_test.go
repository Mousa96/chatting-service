package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Mousa96/chatting-service/internal/auth/models"
	"github.com/Mousa96/chatting-service/internal/auth/repository"
	"github.com/Mousa96/chatting-service/internal/auth/service"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	repo := repository.NewTestUserRepository()
	jwtKey := []byte("test-key")
	authService := service.NewAuthService(repo, jwtKey)
	handler := NewAuthHandler(authService)

	tests := []struct {
		name          string
		request       models.CreateUserRequest
		expectedCode  int
	}{
		{
			name: "Valid registration",
			request: models.CreateUserRequest{
				Username: "testuser",
				Password: "testpass123",
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid request - missing password",
			request: models.CreateUserRequest{
				Username: "testuser",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Duplicate username",
			request: models.CreateUserRequest{
				Username: "testuser",
				Password: "testpass123",
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			handler.Register(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedCode == http.StatusOK {
				var response models.AuthResponse
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Token)
				assert.Equal(t, tt.request.Username, response.User.Username)
				assert.NotZero(t, response.User.ID)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	repo := repository.NewTestUserRepository()
	jwtKey := []byte("test-key")
	authService := service.NewAuthService(repo, jwtKey)
	handler := NewAuthHandler(authService)

	// Create a test user first
	validUser := models.CreateUserRequest{
		Username: "testuser",
		Password: "testpass123",
	}
	authService.Register(&validUser)

	tests := []struct {
		name          string
		request       models.LoginRequest
		expectedCode  int
	}{
		{
			name: "Valid login",
			request: models.LoginRequest{
				Username: "testuser",
				Password: "testpass123",
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid credentials",
			request: models.LoginRequest{
				Username: "testuser",
				Password: "wrongpass",
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "User not found",
			request: models.LoginRequest{
				Username: "nonexistent",
				Password: "testpass123",
			},
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			handler.Login(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedCode == http.StatusOK {
				var response models.AuthResponse
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Token)
				assert.Equal(t, tt.request.Username, response.User.Username)
			}
		})
	}
} 