package models

import "time"

type Message struct {
    ID         int       `json:"id"`
    SenderID   int       `json:"sender_id"`
    ReceiverID int       `json:"receiver_id"`
    Content    string    `json:"content"`
    MediaURL   string    `json:"media_url,omitempty"`
    CreatedAt  time.Time `json:"created_at"`
}

type CreateMessageRequest struct {
    ReceiverID int    `json:"receiver_id" validate:"required"`
    Content    string `json:"content" validate:"required"`
    MediaURL   string `json:"media_url,omitempty"`
}