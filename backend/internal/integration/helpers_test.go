// Package integration provides test helpers for integration testing
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	authModels "github.com/Mousa96/chatting-service/internal/auth/models"
	msgModels "github.com/Mousa96/chatting-service/internal/message/models"
)

const baseURL = "http://localhost:8080/api"


// setupTestUser creates a new test user and returns their authentication token
func setupTestUser(username, password string) string {
	server := setupTestServer(testDB)

	regReq := authModels.CreateUserRequest{
		Username: username,
		Password: password,
	}
	regBody, err := json.Marshal(regReq)
	if err != nil {
		panic(err) // For test setup, panic is acceptable
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(regBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	var resp authModels.AuthResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		panic(err)
	}
	return resp.Token
}

func sendTestMessage(req msgModels.CreateMessageRequest, token string) *httptest.ResponseRecorder {
	server := setupTestServer(testDB)

	body, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}

	httpReq := httptest.NewRequest(http.MethodPost, "/api/messages", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, httpReq)
	return rr
}

func makeAuthenticatedRequest(method, path, token string, body io.Reader) (*httptest.ResponseRecorder, error) {
	server := setupTestServer(testDB)
	
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)
	return rr, nil
}

func getTestConversation(userID int, token string) ([]msgModels.Message, error) {
	rr, err := makeAuthenticatedRequest("GET", fmt.Sprintf("/api/messages/conversation?user_id=%d", userID), token, nil)
	if err != nil {
		return nil, err
	}

	if rr.Code != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", rr.Code, rr.Body.String())
	}

	var response struct {
		Messages []msgModels.Message `json:"messages"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response.Messages, nil
}
