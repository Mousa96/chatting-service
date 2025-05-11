package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimitMiddleware(t *testing.T) {
	// Create a simple handler to wrap with rate limiting
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	tests := []struct {
		name           string
		requests       int
		interval       time.Duration
		limit          int
		expectedStatus []int
	}{
		{
			name:           "Under limit",
			requests:       3,
			interval:       time.Second,
			limit:          5,
			expectedStatus: []int{200, 200, 200},
		},
		{
			name:           "At limit",
			requests:       5,
			interval:       time.Second,
			limit:          5,
			expectedStatus: []int{200, 200, 200, 200, 200},
		},
		{
			name:           "Exceed limit",
			requests:       7,
			interval:       time.Second,
			limit:          5,
			expectedStatus: []int{200, 200, 200, 200, 200, 429, 429},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create rate limiter with the test parameters
			rateLimitedHandler := RateLimitMiddleware(testHandler, tt.limit, tt.interval)
			
			// Make multiple requests and check status codes
			for i := 0; i < tt.requests; i++ {
				req := httptest.NewRequest("GET", "/test", nil)
				// Use the same IP for all requests
				req.RemoteAddr = "192.168.1.1:12345"
				
				rr := httptest.NewRecorder()
				rateLimitedHandler.ServeHTTP(rr, req)
				
				assert.Equal(t, tt.expectedStatus[i], rr.Code, "Request %d should return status %d", i+1, tt.expectedStatus[i])
			}
		})
	}
}

func TestRateLimitMiddlewareWithDifferentIPs(t *testing.T) {
	// Create a simple handler to wrap with rate limiting
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})
	
	// Create rate limiter with a limit of 2 requests per second
	rateLimitedHandler := RateLimitMiddleware(testHandler, 2, time.Second)
	
	// Make requests from different IPs
	ips := []string{"192.168.1.1:12345", "192.168.1.2:12345", "192.168.1.3:12345"}
	
	// Each IP should get their own limit
	for _, ip := range ips {
		// First two requests should succeed
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = ip
			
			rr := httptest.NewRecorder()
			rateLimitedHandler.ServeHTTP(rr, req)
			
			assert.Equal(t, http.StatusOK, rr.Code, "Request %d from IP %s should return 200", i+1, ip)
		}
		
		// Third request should be rate limited
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip
		
		rr := httptest.NewRecorder()
		rateLimitedHandler.ServeHTTP(rr, req)
		
		assert.Equal(t, http.StatusTooManyRequests, rr.Code, "Request 3 from IP %s should be rate limited", ip)
	}
} 