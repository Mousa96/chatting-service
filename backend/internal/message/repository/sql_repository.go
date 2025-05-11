// Package repository implements the message repository interface
package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

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
	const query = `
        INSERT INTO messages (sender_id, receiver_id, content, media_url, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id
    `

	return r.db.QueryRow(query, msg.SenderID, msg.ReceiverID, msg.Content, msg.MediaURL, models.StatusSent, msg.CreatedAt, msg.UpdatedAt).
		Scan(&msg.ID)
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

// GetMessageHistoryPaginated retrieves messages with pagination
func (r *SQLMessageRepository) GetMessageHistoryPaginated(userID, page, pageSize int) ([]models.Message, *models.Pagination, error) {
	// Validate and normalize pagination parameters
	if page < 1 {
		page = 1
	}
	
	if pageSize < 1 {
		pageSize = 10 // Default page size
	} else if pageSize > 100 {
		pageSize = 100 // Maximum page size
	}
	
	// Calculate offset
	offset := (page - 1) * pageSize
	
	// First, get the total count for pagination
	var totalItems int
	countQuery := `SELECT COUNT(*) FROM messages WHERE sender_id = $1 OR receiver_id = $1`
	err := r.db.QueryRow(countQuery, userID).Scan(&totalItems)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count messages: %w", err)
	}
	
	// Create the pagination object
	pagination := models.NewPagination(page, pageSize, totalItems)
	
	// If the page is beyond available data, return empty results
	if page > pagination.TotalPages && pagination.TotalPages > 0 {
		return []models.Message{}, pagination, nil
	}
	
	// Query with pagination
	query := `SELECT id, sender_id, receiver_id, content, media_url, status, created_at 
		FROM messages 
		WHERE sender_id = $1 OR receiver_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	
	rows, err := r.db.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch messages: %w", err)
	}
	defer rows.Close()
	
	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var createdAt time.Time
		
		err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.ReceiverID,
			&msg.Content,
			&msg.MediaURL,
			&msg.Status,
			&createdAt,
		)
		
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan message row: %w", err)
		}
		
		msg.CreatedAt = createdAt
		messages = append(messages, msg)
	}
	
	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error iterating message rows: %w", err)
	}
	
	return messages, pagination, nil
}

// GetConversationPaginated retrieves the conversation between two users with pagination
func (r *SQLMessageRepository) GetConversationPaginated(userID1, userID2, page, pageSize int) ([]models.Message, *models.Pagination, error) {
	// Validate and normalize pagination parameters
	if page < 1 {
		page = 1
	}
	
	if pageSize < 1 {
		pageSize = 10 // Default page size
	} else if pageSize > 100 {
		pageSize = 100 // Maximum page size
	}
	
	// Calculate offset
	offset := (page - 1) * pageSize
	
	// First, get the total count for pagination
	var totalItems int
	countQuery := `SELECT COUNT(*) FROM messages 
		WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1)`
	err := r.db.QueryRow(countQuery, userID1, userID2).Scan(&totalItems)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count messages: %w", err)
	}
	
	// Create the pagination object
	pagination := models.NewPagination(page, pageSize, totalItems)
	
	// If the page is beyond available data, return empty results
	if page > pagination.TotalPages && pagination.TotalPages > 0 {
		return []models.Message{}, pagination, nil
	}
	
	// Query with pagination
	query := `SELECT id, sender_id, receiver_id, content, media_url, status, created_at 
		FROM messages 
		WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`
	
	rows, err := r.db.Query(query, userID1, userID2, pageSize, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch conversation: %w", err)
	}
	defer rows.Close()
	
	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var createdAt time.Time
		
		err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.ReceiverID,
			&msg.Content,
			&msg.MediaURL,
			&msg.Status,
			&createdAt,
		)
		
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan message row: %w", err)
		}
		
		msg.CreatedAt = createdAt
		messages = append(messages, msg)
	}
	
	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error iterating message rows: %w", err)
	}
	
	return messages, pagination, nil
}

// GetMessagesByUser retrieves all messages involving a user
func (r *SQLMessageRepository) GetMessagesByUser(userID int) ([]models.Message, error) {
	query := `
		SELECT id, sender_id, receiver_id, content, media_url, status, created_at
		FROM messages
		WHERE sender_id = $1 OR receiver_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages for user: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var createdAt time.Time
		
		err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.ReceiverID,
			&msg.Content,
			&msg.MediaURL,
			&msg.Status,
			&createdAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan message row: %w", err)
		}
		
		msg.CreatedAt = createdAt
		messages = append(messages, msg)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating message rows: %w", err)
	}
	
	return messages, nil
}