// Package handler provides HTTP handlers for authentication operations
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Mousa96/chatting-service/internal/auth/models"
	"github.com/Mousa96/chatting-service/internal/auth/service"
)

func NewAuthHandler(authService service.Service) Handler {
	return &AuthHandler{authService: authService}
}

// AuthHandler implements the authentication HTTP handlers
type AuthHandler struct {
	authService service.Service
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account and returns JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.CreateUserRequest  true  "User registration details"
// @Success      200      {object}  models.AuthResponse       "Registration successful"
// @Failure      400      {string}  string                    "Bad request"
// @Failure      500      {string}  string                    "Internal server error"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	resp, err := h.authService.Register(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Login godoc
// @Summary      User login
// @Description  Authenticates a user and returns JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.LoginRequest  true  "Login credentials"
// @Success      200      {object}  models.AuthResponse  "Login successful"
// @Failure      400      {string}  string               "Bad request"
// @Failure      401      {string}  string               "Unauthorized"
// @Failure      500      {string}  string               "Internal server error"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
