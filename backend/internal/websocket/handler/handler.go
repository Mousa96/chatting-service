// Package handler provides HTTP handlers for WebSocket operations
package handler

import (
	"encoding/json"
	"log"
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
	jwtKey    []byte
}

// NewWebSocketHandler creates a new WebSocketHandler instance
func NewWebSocketHandler(wsService service.Service, jwtKey []byte) Handler {
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
		jwtKey: jwtKey,
	}
}

// HandleConnection upgrades an HTTP connection to WebSocket and handles the connection
// HandleConnection godoc
// @Summary      WebSocket connection
// @Description  Establishes a WebSocket connection for real-time messaging
// @Tags         WebSocket
// @Accept       json
// @Produce      json
// @Success      101  {string}  string  "Switching protocols"
// @Failure      400  {string}  string  "Bad request"
// @Failure      401  {string}  string  "Unauthorized"
// @Security     Bearer
// @Router       /ws [get]
func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleConnection: Received WebSocket connection request from %s", r.RemoteAddr)
	
	// Extract token from query parameter
	token := r.URL.Query().Get("token")
	log.Printf("HandleConnection: Token from query: %s", maskToken(token))
	
	if token == "" {
		log.Printf("HandleConnection: No token found in query parameters")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Validate the token and extract user ID
	userID, err := middleware.ValidateTokenAndGetUserID(token, string(h.jwtKey))
	if err != nil {
		log.Printf("HandleConnection: Token validation failed: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	log.Printf("HandleConnection: Token valid for user ID: %d", userID)

	// Upgrade the HTTP server connection to the WebSocket protocol
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("HandleConnection: Failed to upgrade connection: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	log.Printf("HandleConnection: Successfully upgraded to WebSocket for user %d", userID)

	// Handle the WebSocket connection
	if err := h.wsService.HandleConnection(conn, userID); err != nil {
		log.Printf("HandleConnection: Error handling WebSocket: %v", err)
		conn.Close()
		return
	}
}

// Helper function to mask most of the token for logging
func maskToken(token string) string {
	if len(token) <= 10 {
		return "***"
	}
	return token[:5] + "..." + token[len(token)-5:]
}

// GetUserStatus returns the online status of a user
// GetUserStatus godoc
// @Summary      Get user status
// @Description  Retrieves the online status of a specific user
// @Tags         WebSocket
// @Accept       json
// @Produce      json
// @Param        user_id  query     int  true  "User ID to get status for"
// @Success      200      {object}  map[string]string  "User status response with status field"
// @Failure      400      {string}  string             "Bad request"
// @Failure      401      {string}  string             "Unauthorized"
// @Failure      404      {string}  string             "User not found"
// @Security     Bearer
// @Router       /ws/status [get]
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
// GetConnectedUsers godoc
// @Summary      Get all connected users
// @Description  Retrieves a list of all currently connected user IDs
// @Tags         WebSocket
// @Accept       json
// @Produce      json
// @Success      200  {array}   int     "List of connected user IDs"
// @Failure      401  {string}  string  "Unauthorized"
// @Failure      500  {string}  string  "Internal server error"
// @Security     Bearer
// @Router       /ws/users [get]
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
