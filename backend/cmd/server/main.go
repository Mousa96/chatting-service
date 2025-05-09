package main

import (
	"log"
	"net/http"

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
	defer database.Close()

	log.Println("Successfully connected to database")

	// Initialize repositories
	userRepo := authRepo.NewUserRepository(database)
	messageRepo := msgRepo.NewMessageRepository(database)

	// Initialize services
	jwtKey := []byte("your-secret-key") // In production, use environment variable
	var authSvc authService.Service = authService.NewAuthService(userRepo, jwtKey)
	var messageSvc msgService.Service = msgService.NewMessageService(messageRepo)

	// Initialize handlers
	var authHdlr authHandler.Handler = authHandler.NewAuthHandler(authSvc)
	var messageHdlr msgHandler.Handler = msgHandler.NewMessageHandler(messageSvc)

	// Initialize middleware
	authMiddleware := middleware.AuthMiddleware(jwtKey)

	// Create a new ServeMux for better route handling
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
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
	log.Printf("Server starting on %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
} 