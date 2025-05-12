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
	"github.com/Mousa96/chatting-service/internal/router"
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
	wsHandler := wsHandler.NewWebSocketHandler(wsSvc, jwtKey)
	
	// Register routes with WebSocket auth middleware
	wsAuthMiddleware := middleware.WebSocketAuthMiddleware(jwtKey)
	mux.Handle("/ws", wsAuthMiddleware(http.HandlerFunc(wsHandler.HandleConnection)))
	mux.Handle("/ws/status", authMiddleware(http.HandlerFunc(wsHandler.GetUserStatus)))
	mux.Handle("/ws/users", authMiddleware(http.HandlerFunc(wsHandler.GetConnectedUsers)))

	// Return the WebSocket service
	return wsSvc
}

func setupWebSocketRoutesWithThrottling(
	mux *http.ServeMux,
	messageSvc msgService.Service,
	authMiddleware func(http.Handler) http.Handler,
	jwtKey []byte,
	throttleLimit int,
	throttleWindow time.Duration,
) wsService.Service {
	wsRepo := wsRepository.NewMemoryRepository()
	wsSvc := wsService.NewWebSocketServiceWithThrottling(wsRepo, messageSvc, throttleLimit, throttleWindow)
	wsHandler := wsHandler.NewWebSocketHandler(wsSvc, jwtKey)

	// WebSocket endpoint
	mux.Handle("/ws", middleware.WebSocketAuthMiddleware(jwtKey)(http.HandlerFunc(wsHandler.HandleConnection)))
	mux.Handle("/ws/status", authMiddleware(http.HandlerFunc(wsHandler.GetUserStatus)))
	mux.Handle("/ws/users", authMiddleware(http.HandlerFunc(wsHandler.GetConnectedUsers)))
	
	// Other WebSocket routes...
	
	return wsSvc
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
	authRepo := authRepository.NewUserRepository(database)
	userRepo := userRepository.NewPostgresRepository(database)
	
	// Initialize storage
	fileStorage := storage.NewLocalStorage("uploads", "/uploads")
	
	// Initialize JWT key
	jwtKey := []byte("your-secret-key") // In production, use environment variable
	
	// Initialize services
	authSvc := authService.NewAuthService(authRepo, jwtKey)
	messageSvc := msgService.NewMessageService(msgRepo.NewMessageRepository(database), fileStorage)
	
	// Initialize handlers
	authHdlr := authHandler.NewAuthHandler(authSvc)
	messageHdlr := msgHandler.NewMessageHandler(messageSvc)
	
	// Set up WebSocket components
	wsRepo := wsRepository.NewMemoryRepository()
	wsSvc := wsService.NewWebSocketServiceWithThrottling(
		wsRepo, 
		messageSvc, 
		5,    // 5 messages per second max
		time.Second,
	)
	wsHdlr := wsHandler.NewWebSocketHandler(wsSvc, jwtKey)
	
	// Initialize user components
	userSvc := userService.NewUserService(userRepo, wsSvc)
	userHdlr := userHandler.NewUserHandler(userSvc)
	
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
