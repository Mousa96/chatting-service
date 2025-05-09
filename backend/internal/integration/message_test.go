package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"mime/multipart"

	msgModels "github.com/Mousa96/chatting-service/internal/message/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageFlow(t *testing.T) {
	// Setup two users and get their tokens
	user1Token := setupTestUser("user1", "pass123")
	t.Logf("User1 Token: %s", user1Token)

	user2Token := setupTestUser("user2", "pass123")
	t.Logf("User2 Token: %s", user2Token)

	// Test sending message (user1 -> user2)
	msgReq := msgModels.CreateMessageRequest{
		ReceiverID: 3, // user2's ID (we can see from logs it's ID 3)
		Content:    "Hello from user1!",
	}

	resp := sendTestMessage(msgReq, user1Token)
	t.Logf("Send Message Response: %s", resp.Body.String())

	// Get conversation as user2 with user1 (ID 2)
	messages, err := getTestConversation(2, user2Token)
	if err != nil {
		t.Fatalf("Failed to get conversation: %v", err)
	}
	
	assert.Len(t, messages, 1)
	if len(messages) > 0 {
		assert.Equal(t, "Hello from user1!", messages[0].Content)
	}
}

func TestMediaUpload(t *testing.T) {
	// Setup user and get token
	userToken := setupTestUser("mediauser", "pass123")

	// Create test file content with actual JPEG header
	fileContents := []byte{
		0xFF, 0xD8, 0xFF, 0xE0, // JPEG SOI and APP0 marker
		0x00, 0x10, 0x4A, 0x46, // APP0 length and "JF"
		0x49, 0x46, 0x00, 0x01, // "IF" and version
		// Add some fake image data
		0x01, 0x02, 0x03, 0x04,
	}
	
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// Create form file
	part, err := writer.CreateFormFile("media", "test.jpg")
	assert.NoError(t, err)
	
	_, err = part.Write(fileContents)
	assert.NoError(t, err)
	err = writer.Close()
	assert.NoError(t, err)

	// Make request
	req := httptest.NewRequest(http.MethodPost, "/api/messages/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))

	rr := httptest.NewRecorder()
	testServer.ServeHTTP(rr, req)

	// Debug output
	t.Logf("Response Status: %d", rr.Code)
	t.Logf("Response Body: %s", rr.Body.String())
	t.Logf("Content-Type header: %s", req.Header.Get("Content-Type"))

	assert.Equal(t, http.StatusOK, rr.Code)

	var response struct {
		URL string `json:"url"`
	}
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Contains(t, response.URL, "uploads/")
}

func TestBroadcastMessage(t *testing.T) {
	// Setup sender and get token
	senderToken := setupTestUser("broadcaster", "pass123")

	// Create broadcast request
	req := msgModels.BroadcastMessageRequest{
		ReceiverIDs: []int{2, 3, 4},
		Content:     "Hello everyone!",
	}

	body, err := json.Marshal(req)
	require.NoError(t, err)

	// Make request
	httpReq := httptest.NewRequest(http.MethodPost, "/api/messages/broadcast", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", senderToken))

	rr := httptest.NewRecorder()
	testServer.ServeHTTP(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify response
	var response struct {
		Messages []*msgModels.Message `json:"messages"`
	}
	err = json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Len(t, response.Messages, len(req.ReceiverIDs))
}

func TestGetMessageHistory(t *testing.T) {
	// Setup user and get token
	userToken := setupTestUser("historyuser", "pass123")

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/api/messages/history", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))

	rr := httptest.NewRecorder()
	testServer.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response struct {
		Messages []msgModels.Message `json:"messages"`
	}
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
}

func TestMessageStatusUpdate(t *testing.T) {
	// Register sender
	senderToken := setupTestUser("sender", "password")
	
	// Register receiver
	receiverToken := setupTestUser("receiver", "password")
	
	// Create a message from sender to receiver
	msgContent := "Test message for status update"
	sendReq := map[string]interface{}{
		"receiver_id": 8, // ID of receiver
		"content":     msgContent,
	}
	
	body, err := json.Marshal(sendReq)
	require.NoError(t, err)
	
	req := httptest.NewRequest(http.MethodPost, "/api/messages", bytes.NewBuffer(body))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", senderToken))
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	testServer.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	
	// Get the message ID from response
	var sendResp map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &sendResp)
	require.NoError(t, err)
	
	messageID := int(sendResp["id"].(float64))
	require.NotEqual(t, 0, messageID, "Message ID should not be 0")
	
	// Now run your status update tests with this messageID
	tests := []struct {
		name         string
		token        string
		status       string
		expectedCode int
	}{
		{"Mark_as_delivered_by_receiver", receiverToken, "delivered", http.StatusOK},
		{"Mark_as_read_by_receiver", receiverToken, "read", http.StatusOK},
		{"Attempt_update_by_sender", senderToken, "delivered", http.StatusForbidden},
		{"Invalid_status", receiverToken, "invalid", http.StatusBadRequest},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateReq := struct {
				MessageID int    `json:"message_id"`
				Status    string `json:"status"`
			}{
				MessageID: messageID,
				Status:    tt.status,
			}

			body, err := json.Marshal(updateReq)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/api/messages/status", bytes.NewBuffer(body))
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tt.token))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			testServer.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
		})
	}
}
