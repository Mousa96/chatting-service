// backend/internal/router/router.go
package router

import (
	"net/http"

	authHandler "github.com/Mousa96/chatting-service/internal/auth/handler"
	msgHandler "github.com/Mousa96/chatting-service/internal/message/handler"
	userHandler "github.com/Mousa96/chatting-service/internal/user/handler"
	wsHandler "github.com/Mousa96/chatting-service/internal/websocket/handler"
)

// Config contains all dependencies needed for the router
type Config struct {
	AuthHandler    authHandler.Handler
	MessageHandler msgHandler.Handler
	UserHandler    userHandler.Handler
	WebSocketHandler wsHandler.Handler
	JWTKey         []byte
}

// New creates and returns a configured HTTP router with all routes registered
func New(config Config) http.Handler {
	mux := http.NewServeMux()
	
	// Register routes by category
	registerHealthCheck(mux)
	registerAuthRoutes(mux, config.AuthHandler)
	registerMessageRoutes(mux, config.MessageHandler, config.JWTKey)
	registerUserRoutes(mux, config.UserHandler, config.JWTKey)
	registerWebSocketRoutes(mux, config.WebSocketHandler, config.JWTKey)
	
	// Apply global middleware using local corsMiddleware
	handler := corsMiddleware(mux)
	
	return handler
}

// Helper to register health check
func registerHealthCheck(mux *http.ServeMux) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("OK"))
	})
}