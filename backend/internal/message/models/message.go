// Package models provides the data structures for messaging functionality
package models

import "time"

// MessageStatus represents the delivery status of a message
type MessageStatus string

const (
	StatusSent      MessageStatus = "sent"
	StatusDelivered MessageStatus = "delivered"
	StatusRead      MessageStatus = "read"
)

// Message represents a chat message in the system
type Message struct {
	ID         int           `json:"id"`
	SenderID   int           `json:"sender_id"`
	ReceiverID int          `json:"receiver_id"`
	Content    string       `json:"content"`
	MediaURL   string       `json:"media_url,omitempty"`
	Status     MessageStatus `json:"status"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at,omitempty"`
}

// CreateMessageRequest represents the request body for creating a new message
type CreateMessageRequest struct {
	ReceiverID int    `json:"receiver_id" validate:"required"`
	Content    string `json:"content" validate:"required"`
	MediaURL   string `json:"media_url,omitempty"`
}
