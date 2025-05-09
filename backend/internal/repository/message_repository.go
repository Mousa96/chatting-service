package repository

import (
	"database/sql"

	"github.com/Mousa96/chatting-service/internal/models"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(msg *models.Message) error {
	query := `
        INSERT INTO messages (sender_id, receiver_id, content, media_url)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at`
	
	return r.db.QueryRow(query, msg.SenderID, msg.ReceiverID, msg.Content, msg.MediaURL).
		Scan(&msg.ID, &msg.CreatedAt)
}

func (r *MessageRepository) GetConversation(userID1, userID2 int) ([]models.Message, error) {
	query := `
        SELECT id, sender_id, receiver_id, content, media_url, created_at
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
		err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Content, &msg.MediaURL, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	
	return messages, nil
} 