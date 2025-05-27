package router

import (
	"net/http"
	"time"

	authHandler "github.com/Mousa96/chatting-service/internal/auth/handler"
	msgHandler "github.com/Mousa96/chatting-service/internal/message/handler"
	"github.com/Mousa96/chatting-service/internal/middleware"
	userHandler "github.com/Mousa96/chatting-service/internal/user/handler"

	"os"

	wsHandler "github.com/Mousa96/chatting-service/internal/websocket/handler"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Register authentication routes
func registerAuthRoutes(mux *http.ServeMux, handler authHandler.Handler) {
	mux.Handle("/api/auth/register", corsMiddleware(http.HandlerFunc(handler.Register)))
	mux.Handle("/api/auth/login", corsMiddleware(http.HandlerFunc(handler.Login)))
}
// Register message routes with appropriate middleware
func registerMessageRoutes(mux *http.ServeMux, handler msgHandler.Handler, jwtKey []byte) {
	authMiddleware := middleware.AuthMiddleware(jwtKey)
	mux.Handle("/api/messages", corsMiddleware(
		middleware.RateLimitMiddleware(
			authMiddleware(http.HandlerFunc(handler.SendMessage)),
			10,
			time.Minute,
		),
	))
	
	// Get conversation endpoint - register both with and without trailing slash
	conversationHandler := corsMiddleware(
		middleware.RateLimitMiddleware(
			authMiddleware(http.HandlerFunc(handler.GetConversation)),
			10,
			time.Minute,
		),
	)
	mux.Handle("/api/messages/conversation", conversationHandler)
	
	// Upload media endpoint
	mux.Handle("/api/messages/upload", corsMiddleware(
		middleware.RateLimitMiddleware(
			authMiddleware(http.HandlerFunc(handler.UploadMedia)),
			10,
			time.Minute,
		),
	))
	
	// Message history endpoint
	mux.Handle("/api/messages/history", corsMiddleware(
		middleware.RateLimitMiddleware(
			authMiddleware(http.HandlerFunc(handler.GetMessageHistory)),
			10,
			time.Minute,
		),
	))
	
	// Update message status endpoint
	mux.Handle("/api/messages/status", corsMiddleware(
		middleware.RateLimitMiddleware(
			authMiddleware(http.HandlerFunc(handler.UpdateMessageStatus)),
			10,
			time.Minute,
		),
	))
	
	// Broadcast messages have stricter rate limit
	mux.Handle("/api/messages/broadcast", corsMiddleware(
		middleware.RateLimitMiddleware(
			authMiddleware(http.HandlerFunc(handler.BroadcastMessage)),
			3,
			time.Minute,
		),
	))
}
// Register user routes
func registerUserRoutes(mux *http.ServeMux, handler userHandler.Handler, jwtKey []byte) {
	authMiddleware := middleware.AuthMiddleware(jwtKey)
	
	// Register user endpoints directly instead of using submux
	// Get all users
	mux.Handle("/api/users", corsMiddleware(authMiddleware(http.HandlerFunc(handler.GetAllUsers))))
	
	// Get user by ID
	mux.Handle("/api/users/profile", corsMiddleware(authMiddleware(http.HandlerFunc(handler.GetUserByID))))
	
	// Update user status
	mux.Handle("/api/users/status", corsMiddleware(authMiddleware(http.HandlerFunc(handler.UpdateUserStatus))))
}
// Register WebSocket routes
func registerWebSocketRoutes(mux *http.ServeMux, handler wsHandler.Handler, jwtKey []byte) {
	mux.Handle("/ws", http.HandlerFunc(handler.ServeWS))
}
// Register static routes
func registerStaticRoutes(mux *http.ServeMux) {
    // Serve uploaded files from /uploads/ directory
    fileServer := http.FileServer(http.Dir("/app/uploads"))
    mux.Handle("/uploads/", http.StripPrefix("/uploads/", fileServer))
    
    // Serve frontend static assets (CSS, JS, images)
    mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("/app/frontend/assets/"))))
    
    // Serve specific frontend pages
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/", "/index.html":
            http.ServeFile(w, r, "/app/frontend/index.html")
        case "/chat.html":
            http.ServeFile(w, r, "/app/frontend/chat.html")
        default:
            // Try to serve as static file from frontend directory
            filePath := "/app/frontend" + r.URL.Path
            if _, err := os.Stat(filePath); err == nil {
                http.ServeFile(w, r, filePath)
            } else {
                // If file doesn't exist, serve index.html (for SPA routing)
                http.ServeFile(w, r, "/app/frontend/index.html")
            }
        }
    })
}
// Register health check
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
// corsMiddleware implementation
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}