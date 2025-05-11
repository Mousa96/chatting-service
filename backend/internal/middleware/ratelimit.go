package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter represents a simple in-memory rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.Mutex
	limit    int
	interval time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		interval: interval,
	}
}

// Allow checks if a request from the given IP should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	
	// Initialize the requests slice if it doesn't exist
	if _, exists := rl.requests[ip]; !exists {
		rl.requests[ip] = []time.Time{}
	}
	
	// Filter out old requests that are outside our time window
	cutoff := now.Add(-rl.interval)
	var validRequests []time.Time
	
	for _, t := range rl.requests[ip] {
		if t.After(cutoff) {
			validRequests = append(validRequests, t)
		}
	}
	
	// Update the requests for this IP
	rl.requests[ip] = validRequests
	
	// Check if the client has exceeded the limit
	if len(validRequests) >= rl.limit {
		return false
	}
	
	// Add the current request time
	rl.requests[ip] = append(rl.requests[ip], now)
	return true
}

// RateLimitMiddleware creates a middleware that limits requests based on client IP
func RateLimitMiddleware(next http.Handler, limit int, interval time.Duration) http.Handler {
	limiter := NewRateLimiter(limit, interval)
	
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the client's IP address
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			// If we can't parse the IP, use the full RemoteAddr
			ip = r.RemoteAddr
		}
		
		// Check if the request is allowed
		if !limiter.Allow(ip) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		
		// If allowed, proceed to the next handler
		next.ServeHTTP(w, r)
	})
} 