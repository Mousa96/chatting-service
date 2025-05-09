package models

// BroadcastMessageRequest represents a request to send a message to multiple users
type BroadcastMessageRequest struct {
    ReceiverIDs []int  `json:"receiver_ids"`
    Content     string `json:"content"`
    MediaURL    string `json:"media_url,omitempty"`
} 