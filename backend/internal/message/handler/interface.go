// Package handler provides HTTP handlers for message operations
package handler

import "net/http"

// Handler defines the message handling interface
type Handler interface {
	// SendMessage handles the message sending request
	SendMessage(w http.ResponseWriter, r *http.Request)
	// GetConversation retrieves the conversation history between users
	GetConversation(w http.ResponseWriter, r *http.Request)
	// UploadMedia handles the media uploading request
	UploadMedia(w http.ResponseWriter, r *http.Request)
	// BroadcastMessage handles the broadcast message request
	BroadcastMessage(w http.ResponseWriter, r *http.Request)
}
