package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Mousa96/chatting-service/internal/middleware"
	"github.com/Mousa96/chatting-service/internal/user/service"
)

// UserHandler implements the Handler interface
type UserHandler struct {
    userService service.Service
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService service.Service) Handler {
    return &UserHandler{userService: userService}
}

// GetAllUsers returns all users except the current user
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
    // Get current user ID from context
    currentUserID, _ := middleware.GetUserIDFromContext(r.Context())
    
    // Get all users
    users, err := h.userService.GetAllUsers()
    if err != nil {
        http.Error(w, "Failed to get users", http.StatusInternalServerError)
        return
    }
    
    // Filter out current user
    filteredUsers := make([]interface{}, 0)
    for _, user := range users {
        if user.ID != currentUserID {
            filteredUsers = append(filteredUsers, user)
        }
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(filteredUsers)
}

// GetUserByID returns a specific user by ID
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
    // Extract user ID from query parameters
    userIDStr := r.URL.Query().Get("id")
    if userIDStr == "" {
        http.Error(w, "Missing user ID", http.StatusBadRequest)
        return
    }
    
    userID, err := strconv.Atoi(userIDStr)
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }
    
    // Get user
    user, err := h.userService.GetUserByID(userID)
    if err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

// UpdateUserStatus updates a user's online status
func (h *UserHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
    // Get current user ID from context
    userID, _ := middleware.GetUserIDFromContext(r.Context())
    
    // Parse request body
    var req struct {
        Status string `json:"status"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Validate status
    if req.Status != "online" && req.Status != "offline" && req.Status != "away" {
        http.Error(w, "Invalid status", http.StatusBadRequest)
        return
    }
    
    // Update status
    if err := h.userService.UpdateUserStatus(userID, req.Status); err != nil {
        http.Error(w, "Failed to update status", http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
} 