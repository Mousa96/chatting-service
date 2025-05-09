package integration

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"testing"

	authHandler "github.com/Mousa96/chatting-service/internal/auth/handler"
	authRepo "github.com/Mousa96/chatting-service/internal/auth/repository"
	authService "github.com/Mousa96/chatting-service/internal/auth/service"
	"github.com/Mousa96/chatting-service/internal/db"
	msgHandler "github.com/Mousa96/chatting-service/internal/message/handler"
	msgRepo "github.com/Mousa96/chatting-service/internal/message/repository"
	msgService "github.com/Mousa96/chatting-service/internal/message/service"
	"github.com/Mousa96/chatting-service/internal/middleware"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	// Setup test database
	dbConfig := &db.Config{
		Host:     "test-db", // match docker-compose service name
		Port:     "5432",    // internal docker port
		User:     "postgres",
		Password: "postgres",
		DBName:   "chat_service_test",
	}

	// Use absolute path for migrations in container
	migrationsPath := "/app/internal/db/migrations"
	if err := db.RunMigrationsWithPath(dbConfig, migrationsPath); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize database connection
	var err error
	testDB, err = db.NewConnection(dbConfig)
	if err != nil {
		log.Fatal("Could not initialize test database connection:", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup: Close connection before dropping tables
	testDB.Close()

	// Clean up database by truncating all tables
	cleanupDB := func() error {
		db, err := db.NewConnection(dbConfig)
		if err != nil {
			return err
		}
		defer db.Close()

		// Truncate all tables in reverse order of dependencies
		_, err = db.Exec(`
			TRUNCATE TABLE messages, users CASCADE;
		`)
		return err
	}

	if err := cleanupDB(); err != nil {
		log.Printf("Failed to cleanup test database: %v", err)
	}

	os.Exit(code)
}

func setupTestServer(db *sql.DB) *http.ServeMux {
	// Initialize repositories
	userRepo := authRepo.NewUserRepository(db)
	messageRepo := msgRepo.NewMessageRepository(db)

	// Initialize services
	jwtKey := []byte("test-key")
	authSvc := authService.NewAuthService(userRepo, jwtKey)
	messageSvc := msgService.NewMessageService(messageRepo)

	// Initialize handlers
	authHdlr := authHandler.NewAuthHandler(authSvc)
	messageHdlr := msgHandler.NewMessageHandler(messageSvc)

	// Initialize middleware
	authMiddleware := middleware.AuthMiddleware(jwtKey)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/register", authHdlr.Register)
	mux.HandleFunc("/api/auth/login", authHdlr.Login)
	mux.Handle("/api/messages", authMiddleware(http.HandlerFunc(messageHdlr.SendMessage)))
	mux.Handle("/api/messages/conversation", authMiddleware(http.HandlerFunc(messageHdlr.GetConversation)))

	return mux
} 