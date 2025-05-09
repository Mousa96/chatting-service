// Package main is the entry point for the chat service application
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	authHandler "github.com/Mousa96/chatting-service/internal/auth/handler"
	authRepo "github.com/Mousa96/chatting-service/internal/auth/repository"
	authService "github.com/Mousa96/chatting-service/internal/auth/service"
	"github.com/Mousa96/chatting-service/internal/db"
	msgHandler "github.com/Mousa96/chatting-service/internal/message/handler"
	msgRepo "github.com/Mousa96/chatting-service/internal/message/repository"
	msgService "github.com/Mousa96/chatting-service/internal/message/service"
	"github.com/Mousa96/chatting-service/internal/middleware"
	"github.com/Mousa96/chatting-service/internal/storage"
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

func main() {
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

	// Initialize repositories
	userRepo := authRepo.NewUserRepository(database)
	messageRepo := msgRepo.NewMessageRepository(database)

	// Initialize storage
	fileStorage := storage.NewLocalStorage("uploads", "/uploads")
	
	// Create a new ServeMux for better route handling
	mux := http.NewServeMux()

	// Ensure uploads directory exists
	uploadsDir := "uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}

	// Serve static files
	fileServer := http.FileServer(http.Dir(uploadsDir))
	mux.Handle("/api/uploads/", http.StripPrefix("/api/uploads/", fileServer))

	// Initialize services
	jwtKey := []byte("your-secret-key") // In production, use environment variable
	authSvc := authService.NewAuthService(userRepo, jwtKey)
	messageSvc := msgService.NewMessageService(messageRepo, fileStorage)

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
	messageMux.HandleFunc("/conversation", messageHdlr.GetConversation)
	messageMux.HandleFunc("/upload", messageHdlr.UploadMedia)
	messageMux.HandleFunc("/broadcast", messageHdlr.BroadcastMessage)
	messageMux.HandleFunc("/history", messageHdlr.GetMessageHistory)
	messageMux.HandleFunc("/status", messageHdlr.UpdateMessageStatus)

	// Apply middleware to protected routes
	mux.Handle("/api/messages/", corsMiddleware(authMiddleware(http.StripPrefix("/api/messages", messageMux))))

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
