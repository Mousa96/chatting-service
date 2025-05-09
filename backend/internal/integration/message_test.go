package integration

import (
	"testing"

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
	messages := getTestConversation(2, user2Token)
	assert.Len(t, messages, 1)
	if len(messages) > 0 {
		assert.Equal(t, "Hello from user1!", messages[0].Content)
	}
}
