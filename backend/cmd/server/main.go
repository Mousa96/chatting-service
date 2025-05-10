// Package main is the entry point for the chat service application
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	authHandler "github.com/Mousa96/chatting-service/internal/auth/handler"
	authRepository "github.com/Mousa96/chatting-service/internal/auth/repository"
	authService "github.com/Mousa96/chatting-service/internal/auth/service"
	"github.com/Mousa96/chatting-service/internal/db"
	msgHandler "github.com/Mousa96/chatting-service/internal/message/handler"
	msgRepo "github.com/Mousa96/chatting-service/internal/message/repository"
	msgService "github.com/Mousa96/chatting-service/internal/message/service"
	"github.com/Mousa96/chatting-service/internal/middleware"
	"github.com/Mousa96/chatting-service/internal/storage"
	userHandler "github.com/Mousa96/chatting-service/internal/user/handler"
	userRepository "github.com/Mousa96/chatting-service/internal/user/repository"
	userService "github.com/Mousa96/chatting-service/internal/user/service"
	wsHandler "github.com/Mousa96/chatting-service/internal/websocket/handler"
	wsRepository "github.com/Mousa96/chatting-service/internal/websocket/repository"
	wsService "github.com/Mousa96/chatting-service/internal/websocket/service"
)

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

func setupWebSocketRoutes(mux *http.ServeMux, messageSvc msgService.Service, authMiddleware func(http.Handler) http.Handler, jwtKey []byte) wsService.Service {
	// Create repository
	wsRepo := wsRepository.NewMemoryRepository()
	
	// Create service with correct name
	wsSvc := wsService.NewWebSocketService(wsRepo, messageSvc)
	
	// Create handler
	wsHandler := wsHandler.NewWebSocketHandler(wsSvc)
	
	// Register routes with WebSocket auth middleware
	wsAuthMiddleware := middleware.WebSocketAuthMiddleware(jwtKey)
	mux.Handle("/ws", wsAuthMiddleware(http.HandlerFunc(wsHandler.HandleConnection)))
	mux.Handle("/ws/status", authMiddleware(http.HandlerFunc(wsHandler.GetUserStatus)))
	mux.Handle("/ws/users", authMiddleware(http.HandlerFunc(wsHandler.GetConnectedUsers)))

	// Return the WebSocket service
	return wsSvc
}

func main() {
	// Add this debugging code at the beginning

	// Create a basic file check
	staticFilePath := "/app/static/index.html"
	if _, err := os.Stat(staticFilePath); os.IsNotExist(err) {
		log.Printf("WARNING: Static file %s does not exist", staticFilePath)
	} else {
		log.Printf("Static file %s exists and is accessible", staticFilePath)
	}

	// Keep only ONE static file handler in your code
	// If you're using a custom mux (recommended):
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("/app/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Or if using default mux (less recommended):
	// fs := http.FileServer(http.Dir("/app/static"))
	// http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Test simple route to verify basic HTTP functionality
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	// Database configuration
	dbConfig := &db.Config{
		Host:     "db",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		DBName:   "chat_service",
	}

	// Run migrations
	if err := db.RunMigrations(dbConfig); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
	log.Println("Migrations completed successfully")

	// Initialize database connection
	database, err := db.NewConnection(dbConfig)
	if err != nil {
		log.Fatal("Could not initialize database connection:", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	log.Println("Successfully connected to database")

	// Initialize separate repositories for auth and user
	authRepo := authRepository.NewUserRepository(database)
	userRepo := userRepository.NewPostgresRepository(database)

	// Initialize storage
	fileStorage := storage.NewLocalStorage("uploads", "/uploads")
	
	// Initialize services
	jwtKey := []byte("your-secret-key") // In production, use environment variable
	authSvc := authService.NewAuthService(authRepo, jwtKey)
	messageSvc := msgService.NewMessageService(msgRepo.NewMessageRepository(database), fileStorage)

	// Initialize handlers
	authHdlr := authHandler.NewAuthHandler(authSvc)
	messageHdlr := msgHandler.NewMessageHandler(messageSvc)

	// Initialize middleware
	authMiddleware := middleware.AuthMiddleware(jwtKey)

	// Public routes
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Error writing response: %v", err)
		}
	})
	mux.Handle("/api/auth/register", corsMiddleware(http.HandlerFunc(authHdlr.Register)))
	mux.Handle("/api/auth/login", corsMiddleware(http.HandlerFunc(authHdlr.Login)))

	// Protected routes
	messageMux := http.NewServeMux()
	messageMux.HandleFunc("/", messageHdlr.SendMessage)
	messageMux.HandleFunc("/conversation/", messageHdlr.GetConversation)
	messageMux.HandleFunc("/upload", messageHdlr.UploadMedia)
	messageMux.HandleFunc("/broadcast", messageHdlr.BroadcastMessage)
	messageMux.HandleFunc("/history", messageHdlr.GetMessageHistory)
	messageMux.HandleFunc("/status", messageHdlr.UpdateMessageStatus)

	// Apply middleware to protected routes
	mux.Handle("/api/messages/", corsMiddleware(authMiddleware(http.StripPrefix("/api/messages", messageMux))))

	// Get websocket service
	wsSvc := setupWebSocketRoutes(mux, messageSvc, authMiddleware, jwtKey)

	// Initialize user components with the websocket service
	userSvc := userService.NewUserService(userRepo, wsSvc)
	userHandler := userHandler.NewUserHandler(userSvc)

	// Set up user routes
	userMux := http.NewServeMux()
	userMux.HandleFunc("/", userHandler.GetAllUsers)
	userMux.HandleFunc("/profile", userHandler.GetUserByID)
	userMux.HandleFunc("/status", userHandler.UpdateUserStatus)

	// Apply middleware to user routes
	mux.Handle("/api/users", corsMiddleware(authMiddleware(userMux)))
	mux.Handle("/api/users/", corsMiddleware(authMiddleware(http.StripPrefix("/api/users", userMux))))

	// Print out the current working directory when the server starts
	fmt.Println("Current working directory:", currentWorkingDir())

	port := ":8080"

	// Create a server with timeouts
	srv := &http.Server{
		Addr:         port,
		Handler:      corsMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server starting on %s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func currentWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "Error getting working directory"
	}
	return dir
}
