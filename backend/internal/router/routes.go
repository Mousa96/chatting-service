package router

import (
	"net/http"
	"time"

	authHandler "github.com/Mousa96/chatting-service/internal/auth/handler"
	msgHandler "github.com/Mousa96/chatting-service/internal/message/handler"
	"github.com/Mousa96/chatting-service/internal/middleware"
	userHandler "github.com/Mousa96/chatting-service/internal/user/handler"
	wsHandler "github.com/Mousa96/chatting-service/internal/websocket/handler"
)

// Register authentication routes
func registerAuthRoutes(mux *http.ServeMux, handler authHandler.Handler) {
	mux.Handle("/api/auth/register", corsMiddleware(http.HandlerFunc(handler.Register)))
	mux.Handle("/api/auth/login", corsMiddleware(http.HandlerFunc(handler.Login)))
}

// Register message routes with appropriate middleware
func registerMessageRoutes(mux *http.ServeMux, handler msgHandler.Handler, jwtKey []byte) {
	authMiddleware := middleware.AuthMiddleware(jwtKey)
	
	// Create submux for message endpoints
	messageMux := http.NewServeMux()
	messageMux.HandleFunc("/", handler.SendMessage)
	messageMux.HandleFunc("/conversation/", handler.GetConversation)
	messageMux.HandleFunc("/upload", handler.UploadMedia)
	messageMux.HandleFunc("/history", handler.GetMessageHistory)
	messageMux.HandleFunc("/status", handler.UpdateMessageStatus)
	
	// Apply middleware to message routes
	rateLimitedMessages := middleware.RateLimitMiddleware(
		authMiddleware(http.StripPrefix("/api/messages", messageMux)),
		10,
		time.Minute,
	)
	mux.Handle("/api/messages/", corsMiddleware(rateLimitedMessages))
	
	// Broadcast messages have stricter rate limit
	broadcastLimiter := middleware.RateLimitMiddleware(
		authMiddleware(http.HandlerFunc(handler.BroadcastMessage)),
		3, 
		time.Minute,
	)
	mux.Handle("/api/messages/broadcast", corsMiddleware(broadcastLimiter))
}

// Register user routes
func registerUserRoutes(mux *http.ServeMux, handler userHandler.Handler, jwtKey []byte) {
	authMiddleware := middleware.AuthMiddleware(jwtKey)
	
	userMux := http.NewServeMux()
	userMux.HandleFunc("/", handler.GetAllUsers)
	userMux.HandleFunc("/profile", handler.GetUserByID)
	userMux.HandleFunc("/status", handler.UpdateUserStatus)
	
	mux.Handle("/api/users", corsMiddleware(authMiddleware(userMux)))
	mux.Handle("/api/users/", corsMiddleware(authMiddleware(http.StripPrefix("/api/users", userMux))))
}

// Register WebSocket routes
func registerWebSocketRoutes(mux *http.ServeMux, handler wsHandler.Handler, jwtKey []byte) {
	authMiddleware := middleware.AuthMiddleware(jwtKey)
	wsAuthMiddleware := middleware.WebSocketAuthMiddleware(jwtKey)
	
	mux.Handle("/ws", wsAuthMiddleware(http.HandlerFunc(handler.HandleConnection)))
	mux.Handle("/ws/status", authMiddleware(http.HandlerFunc(handler.GetUserStatus)))
	mux.Handle("/ws/users", authMiddleware(http.HandlerFunc(handler.GetConnectedUsers)))
}

// corsMiddleware implementation (moved from main.go)
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
