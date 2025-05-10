// Package handler provides HTTP handlers for WebSocket operations
package handler

import "net/http"

// Handler defines the WebSocket HTTP handler interface
type Handler interface {
	// HandleConnection upgrades an HTTP connection to WebSocket
	HandleConnection(w http.ResponseWriter, r *http.Request)
	
	// GetUserStatus returns the online status of a user
	GetUserStatus(w http.ResponseWriter, r *http.Request)
	
	// GetConnectedUsers returns a list of all connected users
	GetConnectedUsers(w http.ResponseWriter, r *http.Request)
}
