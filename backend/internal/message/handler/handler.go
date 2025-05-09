// Package handler implements the HTTP handlers for message operations
package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/Mousa96/chatting-service/internal/message/models"
	"github.com/Mousa96/chatting-service/internal/message/service"
	"github.com/Mousa96/chatting-service/internal/middleware"
)

// Define custom type for context key
type contextKey string
const userIDContextKey = contextKey("user_id")

// MessageHandler provides the implementation of the Handler interface
type MessageHandler struct {
	messageService service.Service
}

// NewMessageHandler creates a new MessageHandler instance
func NewMessageHandler(messageService service.Service) Handler {
	return &MessageHandler{messageService: messageService}
}

func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		log.Printf("Failed to get userID from context") // Add debug log
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	msg, err := h.messageService.SendMessage(userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *MessageHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		log.Printf("Failed to get userID from context") // Add debug log
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	otherUserID := r.URL.Query().Get("user_id")

	otherID, err := strconv.Atoi(otherUserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	messages, err := h.messageService.GetConversation(userID, otherID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(messages); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
