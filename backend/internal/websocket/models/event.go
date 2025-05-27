package models

import (
	"encoding/json"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}



// Event types
const (
	EventSendMessage      = "send_message"
	EventReceiveMessage   = "receive_message" 
	EventBroadcastMessage = "broadcast_message"
	EventStatusChange     = "status_change"
	EventMessageRead      = "message_read"
	EventGetOnlineUsers   = "get_online_users"
	EventUserOnline   = "user_online"
    EventUserOffline  = "user_offline" 
    EventUserStatus   = "user_status"
)


// Message status
type MessageStatus string

const (
	StatusSent      MessageStatus = "sent"
	StatusDelivered MessageStatus = "delivered"
	StatusRead      MessageStatus = "read"
)

// Event payloads
type SendMessageEvent struct {
	Message  string `json:"message"`
	To       int    `json:"to"`
	MediaURL string `json:"media_url,omitempty"`
}

type BroadcastMessageEvent struct {
	Message     string `json:"message"`
	ReceiverIDs []int  `json:"receiver_ids"`
	MediaURL    string `json:"media_url,omitempty"`
}

type StatusChangeEvent struct {
	MessageID int           `json:"message_id"`
	Status    MessageStatus `json:"status"`
	UserID    int           `json:"user_id,omitempty"`
}

type UserStatusEvent struct {
	UserID int    `json:"user_id"`
	Status string `json:"status"`
}

type MessagePayload struct {
	ID         int           `json:"id"`
	SenderID   int           `json:"sender_id"`
	ReceiverID int           `json:"receiver_id"`
	Content    string        `json:"content"`
	MediaURL   string        `json:"media_url,omitempty"`
	Status     MessageStatus `json:"status"`
	CreatedAt  string        `json:"created_at"`
}