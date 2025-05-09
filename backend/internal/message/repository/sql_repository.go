// Package repository implements the message repository interface
package repository

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Mousa96/chatting-service/internal/message/models"
)

// SQLMessageRepository provides a PostgreSQL implementation of Repository
type SQLMessageRepository struct {
	db *sql.DB
}

// NewMessageRepository creates a new SQLMessageRepository instance
func NewMessageRepository(db *sql.DB) Repository {
	return &SQLMessageRepository{db: db}
}

func (r *SQLMessageRepository) Create(msg *models.Message) error {
	query := `
        INSERT INTO messages (sender_id, receiver_id, content, media_url, status)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at`

	return r.db.QueryRow(query, msg.SenderID, msg.ReceiverID, msg.Content, msg.MediaURL, models.StatusSent).
		Scan(&msg.ID, &msg.CreatedAt)
}

func (r *SQLMessageRepository) GetConversation(userID1, userID2 int) ([]models.Message, error) {
	query := `
        SELECT id, sender_id, receiver_id, content, media_url, status, created_at
        FROM messages
        WHERE (sender_id = $1 AND receiver_id = $2)
           OR (sender_id = $2 AND receiver_id = $1)
        ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID1, userID2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.ReceiverID,
			&msg.Content,
			&msg.MediaURL,
			&msg.Status,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *SQLMessageRepository) GetMessageHistory(userID int) ([]models.Message, error) {
    query := `
        SELECT id, sender_id, receiver_id, content, media_url, status, created_at 
        FROM messages 
        WHERE sender_id = $1 OR receiver_id = $1 
        ORDER BY created_at DESC`

    rows, err := r.db.Query(query, userID)
    if err != nil {
        log.Printf("Database error in GetMessageHistory: %v", err)
        return nil, fmt.Errorf("failed to get message history: %w", err)
    }
    defer rows.Close()

    var messages []models.Message
    for rows.Next() {
        var msg models.Message
        err := rows.Scan(
            &msg.ID,
            &msg.SenderID,
            &msg.ReceiverID,
            &msg.Content,
            &msg.MediaURL,
            &msg.Status,
            &msg.CreatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan message: %w", err)
        }
        messages = append(messages, msg)
    }

    if messages == nil {
        messages = []models.Message{} // Return empty slice instead of nil
    }

    return messages, nil
}

func (r *SQLMessageRepository) GetMessageByID(messageID int) (*models.Message, error) {
	query := `
		SELECT id, sender_id, receiver_id, content, media_url, status, created_at, updated_at
		FROM messages
		WHERE id = $1`

	msg := &models.Message{}
	err := r.db.QueryRow(query, messageID).Scan(
		&msg.ID,
		&msg.SenderID,
		&msg.ReceiverID,
		&msg.Content,
		&msg.MediaURL,
		&msg.Status,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("message not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	return msg, nil
}

func (r *SQLMessageRepository) UpdateMessageStatus(messageID int, status models.MessageStatus) error {
    query := `
        UPDATE messages 
        SET status = $1, updated_at = CURRENT_TIMESTAMP 
        WHERE id = $2`

    result, err := r.db.Exec(query, status, messageID)
    if err != nil {
        log.Printf("Database error in UpdateMessageStatus: %v", err)
        return fmt.Errorf("failed to update message status: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking rows affected: %w", err)
    }
    if rows == 0 {
        return fmt.Errorf("message not found")
    }

    return nil
}