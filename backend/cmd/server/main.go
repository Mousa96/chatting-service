// Package main is the entry point for the chat service application
package main

import (
	"log"
	"net/http"
	"time"

	authHandler "github.com/Mousa96/chatting-service/internal/auth/handler"
	authRepo "github.com/Mousa96/chatting-service/internal/auth/repository"
	authService "github.com/Mousa96/chatting-service/internal/auth/service"
	"github.com/Mousa96/chatting-service/internal/db"
	msgHandler "github.com/Mousa96/chatting-service/internal/message/handler"
	msgRepo "github.com/Mousa96/chatting-service/internal/message/repository"
	msgService "github.com/Mousa96/chatting-service/internal/message/service"
	"github.com/Mousa96/chatting-service/internal/middleware"
)

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

	// Initialize services
	jwtKey := []byte("your-secret-key") // In production, use environment variable
	authSvc := authService.NewAuthService(userRepo, jwtKey)
	messageSvc := msgService.NewMessageService(messageRepo)

	// Initialize handlers
	authHdlr := authHandler.NewAuthHandler(authSvc)
	messageHdlr := msgHandler.NewMessageHandler(messageSvc)

	// Initialize middleware
	authMiddleware := middleware.AuthMiddleware(jwtKey)

	// Create a new ServeMux for better route handling
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Error writing response: %v", err)
		}
	})
	mux.HandleFunc("/api/auth/register", authHdlr.Register)
	mux.HandleFunc("/api/auth/login", authHdlr.Login)

	// Protected routes
	messageMux := http.NewServeMux()
	messageMux.HandleFunc("/api/messages", messageHdlr.SendMessage)
	messageMux.HandleFunc("/api/messages/conversation", messageHdlr.GetConversation)

	// Apply middleware to protected routes
	mux.Handle("/api/messages", authMiddleware(messageMux))
	mux.Handle("/api/messages/", authMiddleware(messageMux)) // Note the trailing slash

	port := ":8080"

	// Create a server with timeouts
	srv := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server starting on %s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
