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

// GetAllUsers godoc
// @Summary      Get all users
// @Description  Retrieves all users except the current user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Success      200  {array}   interface{}  "List of users"
// @Failure      401  {string}  string       "Unauthorized"
// @Failure      500  {string}  string       "Internal server error"
// @Security     Bearer
// @Router       /users [get]
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

// GetUserByID godoc
// @Summary      Get user by ID
// @Description  Retrieves a specific user by their ID
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id  query     int  true  "User ID"
// @Success      200  {object}  interface{}  "User details"
// @Failure      400  {string}  string       "Bad request"
// @Failure      401  {string}  string       "Unauthorized"
// @Failure      404  {string}  string       "User not found"
// @Security     Bearer
// @Router       /users/id [get]
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

// UpdateUserStatus godoc
// @Summary      Update user status
// @Description  Updates the current user's online status
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Status update request"
// @Param        request.status  body      string  true  "User status (online/offline/away)"
// @Success      200  {object}  map[string]string  "Success message"
// @Failure      400  {string}  string            "Bad request"
// @Failure      401  {string}  string            "Unauthorized"
// @Failure      500  {string}  string            "Internal server error"
// @Security     Bearer
// @Router       /users/status [put]
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