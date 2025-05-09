package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Mousa96/chatting-service/internal/auth/models"
	"github.com/stretchr/testify/assert"
)

func TestAuthFlow(t *testing.T) {
	// Setup server with real dependencies
	server := setupTestServer(testDB)

	// Test registration
	regReq := models.CreateUserRequest{
		Username: "testuser",
		Password: "testpass123",
	}
	regBody, _ := json.Marshal(regReq)

	// Use /api prefix consistently
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(regBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)
	t.Logf("Registration Response: %s", rr.Body.String())

	var regResp models.AuthResponse
	err := json.NewDecoder(rr.Body).Decode(&regResp)
	assert.NoError(t, err)
	assert.NotEmpty(t, regResp.Token)

	// Test protected endpoint with token
	protectedReq := httptest.NewRequest(http.MethodGet, "/api/messages/conversation?user_id=2", nil)
	protectedReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", regResp.Token))
	protectedReq.Header.Set("Content-Type", "application/json")
	protectedRR := httptest.NewRecorder()

	server.ServeHTTP(protectedRR, protectedReq)
	t.Logf("Protected Endpoint Response: %s", protectedRR.Body.String())
	assert.Equal(t, http.StatusOK, protectedRR.Code)

	// Test login
	loginReq := models.LoginRequest{
		Username: "testuser",
		Password: "testpass123",
	}
	loginBody, _ := json.Marshal(loginReq)

	loginHttpReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(loginBody))
	loginHttpReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()

	server.ServeHTTP(loginRR, loginHttpReq)
	t.Logf("Login Response: %s", loginRR.Body.String())

	assert.Equal(t, http.StatusOK, loginRR.Code)

	var loginResp models.AuthResponse
	err = json.NewDecoder(loginRR.Body).Decode(&loginResp)
	assert.NoError(t, err)
	assert.NotEmpty(t, loginResp.Token)
}
