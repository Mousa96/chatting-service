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
