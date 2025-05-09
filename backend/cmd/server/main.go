package main

import (
	"log"
	"net/http"

	"github.com/Mousa96/chatting-service/internal/db"
	"github.com/Mousa96/chatting-service/internal/handler"
	"github.com/Mousa96/chatting-service/internal/middleware"
	"github.com/Mousa96/chatting-service/internal/repository"
	"github.com/Mousa96/chatting-service/internal/service"
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
	userRepo := repository.NewUserRepository(database)
	messageRepo := repository.NewMessageRepository(database)

	// Initialize services
	jwtKey := []byte("your-secret-key") // In production, use environment variable
	authService := service.NewAuthService(userRepo, jwtKey)
	messageService := service.NewMessageService(messageRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	messageHandler := handler.NewMessageHandler(messageService)

	// Initialize middleware
	authMiddleware := middleware.AuthMiddleware(jwtKey)

	// Create a new ServeMux for better route handling
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)

	// Protected routes
	messageMux := http.NewServeMux()
	messageMux.HandleFunc("/api/messages", messageHandler.SendMessage)
	messageMux.HandleFunc("/api/messages/conversation", messageHandler.GetConversation)

	// Apply middleware to protected routes
	mux.Handle("/api/messages", authMiddleware(messageMux))
	mux.Handle("/api/messages/", authMiddleware(messageMux)) // Note the trailing slash

	port := ":8080"
	log.Printf("Server starting on %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
} 