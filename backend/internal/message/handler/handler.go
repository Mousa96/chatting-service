// Package handler implements the HTTP handlers for message operations
package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

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
	// Extract current user ID from context
	currentUserID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Extract other user ID from path
	// URL format: /conversation/{id}
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 1 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	
	otherUserIDStr := parts[len(parts)-1]
	otherUserID, err := strconv.Atoi(otherUserIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	
	messages, err := h.messageService.GetConversation(currentUserID, otherUserID)
	
	// Return 200 with empty array instead of error if no messages found
	if err != nil && (strings.Contains(err.Error(), "no conversation") || 
					  strings.Contains(err.Error(), "not found")) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"messages": []interface{}{},
		})
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return messages
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
	})
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

func (h *MessageHandler) GetMessageHistory(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get message history
	messages, err := h.messageService.GetMessageHistory(userID)
	if err != nil {
		http.Error(w, "failed to get message history", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
	})
}

func (h *MessageHandler) UpdateMessageStatus(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		MessageID int    `json:"message_id"`
		Status    string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate message ID
	if req.MessageID <= 0 {
		log.Printf("Invalid message ID: %d", req.MessageID)
		http.Error(w, "invalid message_id", http.StatusBadRequest)
		return
	}

	// Convert string to MessageStatus type and log
	log.Printf("Status received: '%s'", req.Status)
	status := models.MessageStatus(req.Status)
	
	// Validate status
	if !status.IsValid() {
		log.Printf("Invalid status: '%s'", status)
		http.Error(w, fmt.Sprintf("invalid status: %s", status), http.StatusBadRequest)
		return
	}

	// Update message status
	if err := h.messageService.UpdateMessageStatus(req.MessageID, status, userID); err != nil {
		log.Printf("Error updating message status: %v, userID: %d, messageID: %d", err, userID, req.MessageID)
		
		if strings.Contains(err.Error(), "not authorized") {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to update message status: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}
