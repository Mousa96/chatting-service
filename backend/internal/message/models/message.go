// Package models provides the data structures for messaging functionality
package models

import "time"

// Message represents a chat message in the system
type Message struct {
	ID         int       `json:"id"`
	SenderID   int       `json:"sender_id"`
	ReceiverID int       `json:"receiver_id"`
	Content    string    `json:"content"`
	MediaURL   string    `json:"media_url,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateMessageRequest represents the request body for creating a new message
type CreateMessageRequest struct {
	ReceiverID int    `json:"receiver_id" validate:"required"`
	Content    string `json:"content" validate:"required"`
	MediaURL   string `json:"media_url,omitempty"`
}
