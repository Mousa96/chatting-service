package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Mousa96/chatting-service/internal/models"
	"github.com/Mousa96/chatting-service/internal/service"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("user_id").(int)

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
	json.NewEncoder(w).Encode(msg)
}

func (h *MessageHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
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
	json.NewEncoder(w).Encode(messages)
} 