package repository

import (
	"database/sql"

	"github.com/Mousa96/chatting-service/internal/user/models"
)

// PostgresRepository implements Repository interface with PostgreSQL
type PostgresRepository struct {
    db *sql.DB
}

// NewPostgresRepository creates a new PostgresRepository
func NewPostgresRepository(db *sql.DB) Repository {
    return &PostgresRepository{db: db}
}

// GetAllUsers retrieves all users from the database
func (r *PostgresRepository) GetAllUsers() ([]models.User, error) {
    rows, err := r.db.Query("SELECT id, username, created_at FROM users")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []models.User
    for rows.Next() {
        var user models.User
        if err := rows.Scan(&user.ID, &user.Username, &user.CreatedAt); err != nil {
            return nil, err
        }
        // Default status, will be updated from WebSocket if needed
        user.Status = "offline"
        users = append(users, user)
    }

    return users, nil
}

// GetUserByID retrieves a user by ID
func (r *PostgresRepository) GetUserByID(id int) (*models.User, error) {
    var user models.User
    err := r.db.QueryRow("SELECT id, username, created_at FROM users WHERE id = $1", id).
        Scan(&user.ID, &user.Username, &user.CreatedAt)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// GetUserByUsername retrieves a user by username
func (r *PostgresRepository) GetUserByUsername(username string) (*models.User, error) {
    var user models.User
    err := r.db.QueryRow("SELECT id, username, created_at FROM users WHERE username = $1", username).
        Scan(&user.ID, &user.Username, &user.CreatedAt)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// UpdateUser updates an existing user
func (r *PostgresRepository) UpdateUser(user *models.User) error {
    _, err := r.db.Exec(
        "UPDATE users SET username = $1 WHERE id = $2",
        user.Username, user.ID,
    )
    return err
}

// UpdateUserStatus updates a user's status
// Note: In a real implementation, you might store status in a dedicated table
func (r *PostgresRepository) UpdateUserStatus(userID int, status string) error {
    // This is a placeholder since status is handled by WebSocket module
    return nil
} 