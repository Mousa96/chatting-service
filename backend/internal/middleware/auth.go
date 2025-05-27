// Package middleware provides HTTP middleware functions
package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Define context key type and constant
type contextKey string
const UserIDKey = contextKey("user_id")

// Claims represents the JWT token claims structure
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// ErrorResponse is a standardized JSON error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// sendJSONError sends a standardized JSON error response
func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// AuthMiddleware creates a new authentication middleware
func AuthMiddleware(jwtKey []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			//log.Printf("Auth header: %s", authHeader) // Debug

			// Check if this is an API request to determine response format
			isAPIRequest := strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/ws/")

			if authHeader == "" {
				log.Printf("No auth header") // Debug
				if isAPIRequest {
					sendJSONError(w, "Authorization header required", http.StatusUnauthorized)
				} else {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
				}
				return
			}

			// Extract the token from the Authorization header
			// Format should be: "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Printf("Invalid auth header format: %v", parts) // Debug
				if isAPIRequest {
					sendJSONError(w, "Invalid authorization header format", http.StatusUnauthorized)
				} else {
					http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				}
				return
			}

			tokenString := parts[1]
			claims := jwt.MapClaims{}

			// Parse and validate the token
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					log.Printf("Unexpected signing method: %v", token.Header["alg"]) // Debug
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return jwtKey, nil
			})

			if err != nil {
				//log.Printf("Token validation error: %v", err) // Debug
				if isAPIRequest {
					sendJSONError(w, "Invalid or expired token", http.StatusUnauthorized)
				} else {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
				}
				return
			}

			if !token.Valid {
				log.Printf("Token invalid") // Debug
				if isAPIRequest {
					sendJSONError(w, "Invalid token", http.StatusUnauthorized)
				} else {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
				}
				return
			}

			//log.Printf("Token validated successfully, claims: %+v", claims) // Debug
			// Add user ID to request context
			userID := int(claims["user_id"].(float64))
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (int, error) {
	userID, ok := ctx.Value(UserIDKey).(int)
	if !ok {
		return 0, errors.New("user ID not found in context")
	}
	return userID, nil
}

// ValidateTokenAndGetUserID validates a JWT token and returns the user ID
func ValidateTokenAndGetUserID(tokenString string, jwtKey string) (int, error) {
	// Parse the JWT token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtKey), nil
	})

	if err != nil {
		return 0, err
	}

	// Extract claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, fmt.Errorf("invalid token")
}