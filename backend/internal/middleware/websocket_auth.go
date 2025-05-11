package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
)

// WebSocketAuthMiddleware is similar to AuthMiddleware but also checks query parameters for tokens
func WebSocketAuthMiddleware(jwtKey []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("WebSocketAuthMiddleware: Request URL: %s", r.URL.String())
			
			// First try to get token from Authorization header
			tokenString := r.Header.Get("Authorization")
			log.Printf("WebSocketAuthMiddleware: Token from header: %s", maskToken(tokenString))
			
			// If not in header, try query parameter
			if tokenString == "" {
				tokenString = r.URL.Query().Get("token")
				log.Printf("WebSocketAuthMiddleware: Token from query: %s", maskToken(tokenString))
			}
			
			// Strip 'Bearer ' prefix if present
			if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
				tokenString = tokenString[7:]
				log.Printf("WebSocketAuthMiddleware: Stripped Bearer prefix")
			}
			
			if tokenString == "" {
				log.Printf("WebSocketAuthMiddleware: No token found")
				http.Error(w, "Missing auth token", http.StatusUnauthorized)
				return
			}
			
			// Parse and validate the token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return jwtKey, nil
			})
			
			if err != nil || !token.Valid {
				log.Printf("WebSocketAuthMiddleware: Token validation failed: %v", err)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			
			// Extract claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				log.Printf("WebSocketAuthMiddleware: Invalid token claims format")
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}
			
			userID := int(claims["user_id"].(float64))
			log.Printf("WebSocketAuthMiddleware: Token valid for user ID: %d", userID)
			
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helper function to mask most of the token for logging
func maskToken(token string) string {
	if token == "" {
		return "[empty]"
	}
	if len(token) <= 10 {
		return "***"
	}
	return token[:5] + "..." + token[len(token)-5:]
}
