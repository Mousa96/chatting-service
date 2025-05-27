// backend/internal/router/router.go
package router

import (
	"net/http"

	_ "github.com/Mousa96/chatting-service/docs" // Import swagger docs
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
	registerSwaggerRoutes(mux) // Add Swagger routes
	registerAuthRoutes(mux, config.AuthHandler)
	registerMessageRoutes(mux, config.MessageHandler, config.JWTKey)
	registerUserRoutes(mux, config.UserHandler, config.JWTKey)
	registerWebSocketRoutes(mux, config.WebSocketHandler, config.JWTKey)
	registerStaticRoutes(mux)
	handler := mux
	return handler
}
