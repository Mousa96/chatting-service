package main

import (
	"log"
	"net/http"

	"github.com/Mousa96/chatting-service/internal/db"
	"github.com/Mousa96/chatting-service/internal/handler"
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

	// Initialize services
	jwtKey := []byte("your-secret-key") // In production, use environment variable
	authService := service.NewAuthService(userRepo, jwtKey)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)

	// Register routes
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	http.HandleFunc("/api/auth/register", authHandler.Register)
	http.HandleFunc("/api/auth/login", authHandler.Login)

	port := ":8080"
	log.Printf("Server starting on %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
} 