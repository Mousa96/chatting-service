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

func (h *MessageHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form with 10MB max memory
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("Failed to parse multipart form: %v", err)
		http.Error(w, "file too large", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("media")
	if err != nil {
		log.Printf("Failed to get form file: %v", err)
		http.Error(w, "invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read a few bytes to detect content type
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		log.Printf("Failed to read file buffer: %v", err)
		http.Error(w, "failed to read file", http.StatusBadRequest)
		return
	}
	
	// Reset file pointer
	_, err = file.Seek(0, 0)
	if err != nil {
		log.Printf("Failed to reset file pointer: %v", err)
		http.Error(w, "failed to process file", http.StatusInternalServerError)
		return
	}
	
	// Detect content type
	contentType := http.DetectContentType(buffer[:n])
	log.Printf("Detected content type: %s", contentType)
	
	if !isAllowedFileType(contentType) {
		log.Printf("File type not allowed: %s", contentType)
		http.Error(w, "file type not allowed", http.StatusBadRequest)
		return
	}

	// Upload file
	url, err := h.messageService.UploadMedia(userID, header)
	if err != nil {
		log.Printf("Failed to upload file: %v", err)
		http.Error(w, "failed to upload file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url": url,
	})
}

func isAllowedFileType(contentType string) bool {
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"video/mp4":  true,
	}
	return allowedTypes[contentType]
}

func (h *MessageHandler) BroadcastMessage(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req models.BroadcastMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Send broadcast message
	messages, err := h.messageService.BroadcastMessage(userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
	})
}
