// Package models provides data structures for WebSocket functionality
package models

import (
	"time"

	messageModels "github.com/Mousa96/chatting-service/internal/message/models"
	"github.com/gorilla/websocket"
)

// EventType represents the type of WebSocket event
type EventType string

const (
	// Event types
	EventMessage      EventType = "message"
	EventTyping       EventType = "typing"
	EventStatusChange EventType = "status_change"
	EventUserStatus   EventType = "user_status"
	EventError        EventType = "error"
	EventBroadcast    EventType = "broadcast"
)

// UserStatus represents a user's online status
type UserStatus string

const (
	StatusOnline  UserStatus = "online"
	StatusOffline UserStatus = "offline"
	StatusAway    UserStatus = "away"
)

// Client represents a connected WebSocket client
type Client struct {
	UserID     int
	Connection *websocket.Conn
	Send       chan []byte
}

// Event represents a WebSocket event to be sent to clients
type Event struct {
	Type      EventType               `json:"type"`
	Payload   interface{}             `json:"payload"`
	SenderID  int                     `json:"sender_id"`
	Timestamp time.Time               `json:"timestamp"`
	Message   *messageModels.Message  `json:"message,omitempty"`
}

// StatusUpdateEvent represents a message status update event
type StatusUpdateEvent struct {
	MessageID int                        `json:"message_id"`
	Status    messageModels.MessageStatus `json:"status"`
}

// TypingEvent represents a user typing notification
type TypingEvent struct {
	UserID     int  `json:"user_id"`
	IsTyping   bool `json:"is_typing"`
	ReceiverID int  `json:"receiver_id"`
}

// UserStatusEvent represents a user's online status change
type UserStatusEvent struct {
	UserID int        `json:"user_id"`
	Status UserStatus `json:"status"`
}
