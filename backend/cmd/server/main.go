package main

import (
	"log"
	"net/http"

	"github.com/Mousa96/chatting-service/internal/db"
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

	// Initialize database connection
	database, err := db.NewConnection(dbConfig)
	if err != nil {
		log.Fatal("Could not initialize database connection:", err)
	}
	defer database.Close()

	log.Println("Successfully connected to database")

	// Basic health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	port := ":8080"
	log.Printf("Server starting on %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
} 