// Package handler provides HTTP handlers for message operations
package handler

import "net/http"

// Handler defines the message handling interface
type Handler interface {
	// SendMessage handles the message sending request
	SendMessage(w http.ResponseWriter, r *http.Request)
	// GetConversation retrieves the conversation history between users
	GetConversation(w http.ResponseWriter, r *http.Request)
}
