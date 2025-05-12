// backend/internal/router/router.go
package router

import (
	"net/http"

	_ "github.com/Mousa96/chatting-service/docs" // Import swagger docs
	authHandler "github.com/Mousa96/chatting-service/internal/auth/handler"
	msgHandler "github.com/Mousa96/chatting-service/internal/message/handler"
	userHandler "github.com/Mousa96/chatting-service/internal/user/handler"
	wsHandler "github.com/Mousa96/chatting-service/internal/websocket/handler"
	httpSwagger "github.com/swaggo/http-swagger"
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

// registerSwaggerRoutes registers Swagger documentation routes
func registerSwaggerRoutes(mux *http.ServeMux) {
	// Serve Swagger UI
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // The URL pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))
}