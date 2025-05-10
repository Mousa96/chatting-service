package middleware

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
)

// WebSocketAuthMiddleware is similar to AuthMiddleware but also checks query parameters for tokens
func WebSocketAuthMiddleware(jwtKey []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First try to get token from Authorization header
			tokenString := r.Header.Get("Authorization")
			
			// If not in header, try query parameter
			if tokenString == "" {
				tokenString = r.URL.Query().Get("token")
			}
			
			// Strip 'Bearer ' prefix if present
			if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
				tokenString = tokenString[7:]
			}
			
			if tokenString == "" {
				http.Error(w, "Missing auth token", http.StatusUnauthorized)
				return
			}
			
			// Parse and validate the token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return jwtKey, nil
			})
			
			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			
			// Extract claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}
			
			userID := int(claims["user_id"].(float64))
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
