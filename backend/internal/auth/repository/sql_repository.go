package repository

import (
	"database/sql"

	"github.com/Mousa96/chatting-service/internal/auth/models"
)

type SQLUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) Repository {
	return &SQLUserRepository{db: db}
}

func (r *SQLUserRepository) Create(user *models.User) error {
	query := `
        INSERT INTO users (username, password_hash)
        VALUES ($1, $2)
        RETURNING id, created_at, updated_at`

	return r.db.QueryRow(query, user.Username, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *SQLUserRepository) GetByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `
        SELECT id, username, password_hash, created_at, updated_at
        FROM users WHERE username = $1`

	err := r.db.QueryRow(query, username).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}
