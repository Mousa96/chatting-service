// Package handler provides HTTP handlers for WebSocket operations
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Mousa96/chatting-service/internal/middleware"
	"github.com/Mousa96/chatting-service/internal/websocket/service"
	"github.com/gorilla/websocket"
)

// WebSocketHandler implements the Handler interface for WebSocket operations
type WebSocketHandler struct {
	wsService service.Service
	upgrader  websocket.Upgrader
}

// NewWebSocketHandler creates a new WebSocketHandler instance
func NewWebSocketHandler(wsService service.Service) Handler {
	return &WebSocketHandler{
		wsService: wsService,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// Allow connections from any origin for development
			// In production, you might want to restrict this
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// HandleConnection upgrades an HTTP connection to WebSocket and handles the connection
func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the authenticated request
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade the HTTP server connection to the WebSocket protocol
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Handle the WebSocket connection
	if err := h.wsService.HandleConnection(conn, userID); err != nil {
		conn.Close()
		// Since we've already upgraded to WebSocket, we can't use http.Error
		// Instead, log the error and close the connection
		return
	}
}

// GetUserStatus returns the online status of a user
func (h *WebSocketHandler) GetUserStatus(w http.ResponseWriter, r *http.Request) {
	// Parse user ID from query parameters
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get user status from service
	status, err := h.wsService.GetUserStatus(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the status as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": string(status)})
}

// GetConnectedUsers returns a list of all currently connected users
func (h *WebSocketHandler) GetConnectedUsers(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	_, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get connected users from service
	userIDs, err := h.wsService.GetConnectedUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return just the userIDs array without wrapping it
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userIDs)
}
