package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/Mousa96/chatting-service/internal/auth/models"
	msgModels "github.com/Mousa96/chatting-service/internal/message/models"
)

func setupTestUser(username, password string) string {
    server := setupTestServer(testDB)
    
    regReq := models.CreateUserRequest{
        Username: username,
        Password: password,
    }
    regBody, _ := json.Marshal(regReq)
    
    req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(regBody))
    req.Header.Set("Content-Type", "application/json")
    rr := httptest.NewRecorder()
    
    server.ServeHTTP(rr, req)
    
    var resp models.AuthResponse
    json.NewDecoder(rr.Body).Decode(&resp)
    return resp.Token
}

func sendTestMessage(req msgModels.CreateMessageRequest, token string) *httptest.ResponseRecorder {
    server := setupTestServer(testDB)
    
    body, _ := json.Marshal(req)
    httpReq := httptest.NewRequest(http.MethodPost, "/api/messages", bytes.NewBuffer(body))
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+token)
    rr := httptest.NewRecorder()
    
    server.ServeHTTP(rr, httpReq)
    return rr
}

func getTestConversation(otherUserID int, token string) []msgModels.Message {
    server := setupTestServer(testDB)
    
    req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/messages/conversation?user_id=%d", otherUserID), nil)
    req.Header.Set("Authorization", "Bearer "+token)
    rr := httptest.NewRecorder()
    
    server.ServeHTTP(rr, req)
    
    var messages []msgModels.Message
    json.NewDecoder(rr.Body).Decode(&messages)
    return messages
} 