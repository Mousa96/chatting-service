// Package main is the entry point for the chat service application
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/Mousa96/chatting-service/docs" // Import swagger docs
	authHandler "github.com/Mousa96/chatting-service/internal/auth/handler"
	authRepository "github.com/Mousa96/chatting-service/internal/auth/repository"
	authService "github.com/Mousa96/chatting-service/internal/auth/service"
	"github.com/Mousa96/chatting-service/internal/db"
	msgHandler "github.com/Mousa96/chatting-service/internal/message/handler"
	msgRepo "github.com/Mousa96/chatting-service/internal/message/repository"
	msgService "github.com/Mousa96/chatting-service/internal/message/service"
	"github.com/Mousa96/chatting-service/internal/router"
	"github.com/Mousa96/chatting-service/internal/storage"
	userHandler "github.com/Mousa96/chatting-service/internal/user/handler"
	userRepository "github.com/Mousa96/chatting-service/internal/user/repository"
	userService "github.com/Mousa96/chatting-service/internal/user/service"

	wsHandler "github.com/Mousa96/chatting-service/internal/websocket/handler"
	wsService "github.com/Mousa96/chatting-service/internal/websocket/service"
	//chatHandler "github.com/Mousa96/chatting-service/internal/chat/models"
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
	authRepo := authRepository.NewUserRepository(database)
	userRepo := userRepository.NewPostgresRepository(database)
	
	// Initialize storage
	fileStorage := storage.NewLocalStorage("/app/uploads", "/uploads")
	
	// Initialize JWT key
	jwtKey := []byte("your-secret-key") // In production, use environment variable
	
	// Initialize services
	authSvc := authService.NewAuthService(authRepo, jwtKey)
	userSvc := userService.NewUserService(userRepo)
	messageSvc := msgService.NewMessageService(msgRepo.NewMessageRepository(database), fileStorage)
	wsSvc := wsService.NewWebSocketService(messageSvc, jwtKey)
	
	// Initialize handlers
	authHdlr := authHandler.NewAuthHandler(authSvc)
	userHdlr := userHandler.NewUserHandler(userSvc)
	messageHdlr := msgHandler.NewMessageHandler(messageSvc)
	wsHdlr := wsHandler.NewWebSocketHandler(wsSvc)

	
	// Configure router
	routerConfig := router.Config{
		AuthHandler:      authHdlr,
		MessageHandler:   messageHdlr,
		UserHandler:      userHdlr,
		WebSocketHandler: wsHdlr,
		JWTKey:           jwtKey,
	}
	
	// Create server with timeouts
	port := ":8080"
	srv := &http.Server{
		Addr:         port,
		Handler:      router.New(routerConfig),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	// Print out the current working directory when the server starts
	fmt.Println("Current working directory:", currentWorkingDir())
	
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
