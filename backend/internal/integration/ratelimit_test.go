package integration

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Mousa96/chatting-service/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestAPIRateLimiting(t *testing.T) {
	// Set up a test server with rate limiting applied
	mux := setupTestServerWithRateLimit(testDB, 3, time.Second)
	
	// Get a valid token for authentication
	token := setupTestUser("rate_limit_user", "password123")
	
	// Make requests to a rate-limited endpoint
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/messages/history", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.RemoteAddr = "192.168.1.1:12345" // Ensure same IP for all requests
		
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		
		// First 3 requests should succeed, then we should see 429 responses
		if i < 3 {
			assert.Equal(t, http.StatusOK, rr.Code, "Request %d should return OK", i+1)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, rr.Code, "Request %d should be rate limited", i+1)
		}
	}
}

// Helper function to set up a test server with rate limiting
func setupTestServerWithRateLimit(db *sql.DB, limit int, window time.Duration) *http.ServeMux {
	mux := setupTestServer(db)
	
	// Apply rate limiting to the entire mux
	rateLimitedMux := http.NewServeMux()
	rateLimitedMux.Handle("/", middleware.RateLimitMiddleware(mux, limit, window))
	
	return rateLimitedMux
} 