package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/Mousa96/chatting-service/internal/message/models"
	"github.com/stretchr/testify/assert"
)

func TestMessageHistoryPagination(t *testing.T) {
	// Set up test server
	mux := setupTestServer(testDB)
	
	// Set up a test user and get token
	userID := setupTestUserAndGetID("pagination_user", "password123")
	token := getAuthToken("pagination_user", "password123")
	
	// Create some test messages for this user
	createTestMessagesForPagination(t, userID)
	
	// Create test cases
	tests := []struct {
		name           string
		url            string
		expectedStatus int
		validateJSON   func(t *testing.T, body []byte)
	}{
		{
			name:           "Get first page with default size",
			url:            "/api/messages/history",
			expectedStatus: http.StatusOK,
			validateJSON: func(t *testing.T, body []byte) {
				var response struct {
					Messages   []models.Message   `json:"messages"`
					Pagination models.Pagination  `json:"pagination"`
				}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				
				// Validate pagination info
				assert.Equal(t, 1, response.Pagination.CurrentPage)
				assert.True(t, response.Pagination.PageSize > 0)
			},
		},
		{
			name:           "Get second page with custom size",
			url:            "/api/messages/history?page=2&page_size=5",
			expectedStatus: http.StatusOK,
			validateJSON: func(t *testing.T, body []byte) {
				var response struct {
					Messages   []models.Message   `json:"messages"`
					Pagination models.Pagination  `json:"pagination"`
				}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				
				// Validate pagination info
				assert.Equal(t, 2, response.Pagination.CurrentPage)
				assert.Equal(t, 5, response.Pagination.PageSize)
			},
		},
		{
			name:           "Invalid page parameter",
			url:            "/api/messages/history?page=invalid",
			expectedStatus: http.StatusBadRequest,
			validateJSON: func(t *testing.T, body []byte) {
				var response struct {
					Error string `json:"error"`
				}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response.Error, "invalid page")
			},
		},
	}
	
	// Helper function to create a test messages for this user
	// You'll need to add this function
	
	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			// Add the authorization header with token
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)
			
			// Verify status code
			assert.Equal(t, tt.expectedStatus, rr.Code)
			
			// Validate response body
			if tt.validateJSON != nil {
				tt.validateJSON(t, rr.Body.Bytes())
			}
		})
	}
}

// Add this helper function to create test messages for the pagination user
func createTestMessagesForPagination(t *testing.T, userID string) {
	// Convert userID to int
	uid, err := strconv.Atoi(userID)
	if err != nil {
		t.Fatalf("Failed to convert userID: %v", err)
	}
	
	// Create another user to exchange messages with
	otherUserID := setupTestUserAndGetID("pagination_recipient", "password123")
	otherUID, err := strconv.Atoi(otherUserID)
	if err != nil {
		t.Fatalf("Failed to convert otherUserID: %v", err)
	}
	
	// Create enough messages to test pagination
	for i := 1; i <= 15; i++ {
		// Alternate between sent and received messages
		var senderID, receiverID int
		if i%2 == 0 {
			senderID = uid
			receiverID = otherUID
		} else {
			senderID = otherUID
			receiverID = uid
		}
		
		// Create message request
		msgReq := models.CreateMessageRequest{
			ReceiverID: receiverID,
			Content:    fmt.Sprintf("Test message %d", i),
		}
		
		// Convert to JSON
		body, err := json.Marshal(msgReq)
		if err != nil {
			t.Fatalf("Failed to marshal message: %v", err)
		}
		
		// Create request
		req := httptest.NewRequest(http.MethodPost, "/api/messages", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		// Get token for sender
		var token string
		if senderID == uid {
			token = getAuthToken("pagination_user", "password123")
		} else {
			token = getAuthToken("pagination_recipient", "password123") 
		}
		req.Header.Set("Authorization", "Bearer "+token)
		
		// Create response recorder
		rr := httptest.NewRecorder()
		
		// Serve the request
		mux := setupTestServer(testDB)
		mux.ServeHTTP(rr, req)
		
		// Ensure message was created successfully
		assert.Equal(t, http.StatusOK, rr.Code)
	}
}

func TestConversationPagination(t *testing.T) {
	// Set up test server
	mux := setupTestServer(testDB)
	
	// Set up test users
	user1ID := setupTestUserAndGetID("user1_pagination", "password123")
	user1Token := getAuthToken("user1_pagination", "password123")
	user2ID := setupTestUserAndGetID("user2_pagination", "password123")
	
	// Convert IDs to integers
	user1IDInt, _ := strconv.Atoi(user1ID)
	user2IDInt, _ := strconv.Atoi(user2ID)
	
	// Create messages directly in DB
	createDirectMessagesInDB(t, user1IDInt, user2IDInt)
	
	// Test basic conversation retrieval
	t.Run("Basic conversation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, 
			fmt.Sprintf("/api/messages/conversation?user_id=%s", user2ID), nil)
		req.Header.Set("Authorization", "Bearer "+user1Token)
		
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		
		// Just verify status code
		assert.Equal(t, http.StatusOK, rr.Code)
		
		// Verify we can parse the response
		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		// Verify messages field exists
		messages, ok := response["messages"]
		assert.True(t, ok, "Response missing 'messages' field")
		
		// Verify it's an array
		messagesArr, ok := messages.([]interface{})
		assert.True(t, ok, "Messages is not an array")
		
		// We should have messages
		assert.NotEmpty(t, messagesArr, "Messages array is empty")
	})
	
	// Test with pagination parameters
	t.Run("With pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet,
			fmt.Sprintf("/api/messages/conversation?user_id=%s&page=1&page_size=5", user2ID), nil)
		req.Header.Set("Authorization", "Bearer "+user1Token)
		
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		
		// Just verify status code
		assert.Equal(t, http.StatusOK, rr.Code)
		
		// Verify we can parse the response
		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		// Verify messages field exists
		messages, ok := response["messages"]
		assert.True(t, ok, "Response missing 'messages' field")
		
		// Verify it's an array
		messagesArr, ok := messages.([]interface{})
		assert.True(t, ok, "Messages is not an array")
		
		// We should have messages
		assert.NotEmpty(t, messagesArr, "Messages array is empty")
	})
}

// Helper to create messages directly in the database
func createDirectMessagesInDB(t *testing.T, user1ID, user2ID int) {
	// Connect to database directly
	db := testDB
	
	// Log the SQL schema first
	rows, err := db.Query("SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = 'messages'")
	if err != nil {
		t.Logf("Schema query error: %v", err)
	} else {
		defer rows.Close()
		t.Log("Messages table schema:")
		for rows.Next() {
			var colName, dataType, isNullable string
			if err := rows.Scan(&colName, &dataType, &isNullable); err == nil {
				t.Logf("  Column: %s, Type: %s, Nullable: %s", colName, dataType, isNullable)
			}
		}
	}
	
	// Create multiple direct messages - WITH EMPTY STRING FOR MEDIA_URL
	for i := 0; i < 5; i++ {
		// Create a message from user1 to user2
		_, err := db.Exec(
			"INSERT INTO messages (sender_id, receiver_id, content, media_url, status, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
			user1ID, user2ID, fmt.Sprintf("Test message %d from user1", i), "", "sent", time.Now(),
		)
		if err != nil {
			t.Fatalf("Failed to insert message from user1: %v", err)
		}
		
		// Create a message from user2 to user1
		_, err = db.Exec(
			"INSERT INTO messages (sender_id, receiver_id, content, media_url, status, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
			user2ID, user1ID, fmt.Sprintf("Test message %d from user2", i), "", "sent", time.Now(),
		)
		if err != nil {
			t.Fatalf("Failed to insert message from user2: %v", err)
		}
	}
	
	// Verify messages were created and log sample message
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM messages WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1)", 
					user1ID, user2ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count messages: %v", err)
	}
	
	// Log a sample message to see structure
	var id, senderID, receiverID int
	var content, mediaURL, status string
	var createdAt time.Time
	err = db.QueryRow("SELECT id, sender_id, receiver_id, content, media_url, status, created_at FROM messages WHERE sender_id = $1 LIMIT 1", 
					user1ID).Scan(&id, &senderID, &receiverID, &content, &mediaURL, &status, &createdAt)
	if err != nil {
		t.Logf("Failed to fetch sample message: %v", err)
	} else {
		t.Logf("Sample message: ID=%d, Sender=%d, Receiver=%d, Content=%s, MediaURL='%s', Status=%s", 
			   id, senderID, receiverID, content, mediaURL, status)
	}
	
	t.Logf("Created %d test messages between users %d and %d", count, user1ID, user2ID)
} 
