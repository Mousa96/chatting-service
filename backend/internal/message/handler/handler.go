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

// SendMessage godoc
// @Summary      Send a message
// @Description  Sends a message to another user
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Param        request  body      models.CreateMessageRequest  true  "Message details"
// @Success      200      {object}  models.Message               "Message sent"
// @Failure      400      {string}  string                       "Bad request"
// @Failure      401      {string}  string                       "Unauthorized"
// @Failure      500      {string}  string                       "Internal server error"
// @Security     Bearer
// @Router       /messages [post]
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

// GetConversation godoc
// @Summary      Get conversation
// @Description  Retrieves message history between two users with pagination
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Param        user_id    query    int  true   "User ID to get conversation with"
// @Param        page       query    int  false  "Page number"
// @Param        page_size  query    int  false  "Items per page"
// @Success      200        {object}  map[string][]models.Message  "Conversation messages"
// @Failure      400        {string}  string                    "Bad request"
// @Failure      401        {string}  string                    "Unauthorized"
// @Failure      500        {string}  string                    "Internal server error"
// @Security     Bearer
// @Router       /messages/conversation [get]
func (h *MessageHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	// Extract current user ID from context
	currentUserID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Extract other user ID from query parameter
	otherUserIDStr := r.URL.Query().Get("user_id")
	if otherUserIDStr == "" {
		http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
		return
	}
	
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
		// Return object with messages field - not direct array
		json.NewEncoder(w).Encode(map[string]interface{}{
			"messages": []models.Message{},
		})
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return messages in object with messages field - not direct array
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
	})
}

// UploadMedia godoc
// @Summary      Upload media
// @Description  Uploads a media file
// @Tags         Messages
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file    true  "Media file to upload"
// @Success      200   {object}  map[string]string  "File URL"
// @Failure      400   {string}  string            "Bad request"
// @Failure      401   {string}  string            "Unauthorized"
// @Failure      500   {string}  string            "Internal server error"
// @Security     Bearer
// @Router       /messages/upload [post]
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

// BroadcastMessage godoc
// @Summary      Broadcast a message
// @Description  Sends a message to multiple users
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Param        request  body      models.BroadcastMessageRequest  true  "Broadcast message details"
// @Success      200      {array}   models.Message                 "Messages broadcasted"
// @Failure      400      {string}  string                         "Bad request"
// @Failure      401      {string}  string                         "Unauthorized"
// @Failure      429      {string}  string                         "Rate limit exceeded"
// @Failure      500      {string}  string                         "Internal server error"
// @Security     Bearer
// @Router       /messages/broadcast [post]
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

// GetMessageHistory godoc
// @Summary      Get user message history
// @Description  Retrieves all messages for the current user with pagination
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Param        page       query     int  false  "Page number (default: 1)"
// @Param        page_size  query     int  false  "Items per page (default: 10)"
// @Success      200        {object}  map[string]interface{} "Messages with pagination"
// @Failure      400        {object}  map[string]string       "Bad request"
// @Failure      401        {string}  string                 "Unauthorized"
// @Failure      500        {string}  string                 "Internal server error"
// @Security     Bearer
// @Router       /messages/history [get]
func (h *MessageHandler) GetMessageHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Initialize default values
	page := 1
	pageSize := 10
	
	// Get page parameter if provided
	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			// Return JSON error response instead of plain text
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "invalid page parameter",
			})
			return
		}
	}
	
	// Get page_size parameter if provided
	pageSizeStr := r.URL.Query().Get("page_size")
	if pageSizeStr != "" {
		var err error
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "invalid page_size parameter",
			})
			return
		}
	}
	
	messages, pagination, err := h.messageService.GetMessageHistoryPaginated(userID, page, pageSize)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get message history: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
		"pagination": pagination,
	})
}

// UpdateMessageStatus godoc
// @Summary      Update message status
// @Description  Updates a message's status (read, delivered, etc.)
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Status update request"
// @Param        request.message_id  body   int  true  "Message ID to update"
// @Param        request.status      body   string  true  "New status (read, delivered, etc.)"
// @Success      200  {object}  map[string]string  "Success response"
// @Failure      400  {string}  string            "Bad request"
// @Failure      401  {string}  string            "Unauthorized"
// @Failure      403  {string}  string            "Forbidden"
// @Failure      404  {string}  string            "Not found"
// @Failure      500  {string}  string            "Internal server error"
// @Security     Bearer
// @Router       /messages/status [put]
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

// GetConversationPaginated godoc
// @Summary      Get paginated conversation
// @Description  Retrieves message history between current user and another user with pagination
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Param        user_id    query     int  true   "User ID to get conversation with"
// @Param        page       query     int  false  "Page number (default: 1)"
// @Param        page_size  query     int  false  "Items per page (default: 10, max: 100)"
// @Success      200        {object}  object  "Response with messages array and pagination object"
// @Failure      400        {string}  string  "Bad request"
// @Failure      401        {string}  string  "Unauthorized" 
// @Failure      500        {string}  string  "Internal server error"
// @Security     Bearer
// @Router       /messages/conversation/paginated [get]
func (h *MessageHandler) GetConversationPaginated(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context using the proper key
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get other user ID from query parameters
	otherUserIDStr := r.URL.Query().Get("user_id")
	if otherUserIDStr == "" {
		http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
		return
	}

	otherUserID, err := strconv.Atoi(otherUserIDStr)
	if err != nil {
		http.Error(w, "Invalid user_id parameter", http.StatusBadRequest)
		return
	}

	// Check if pagination is requested
	var page, pageSize int = 1, 10 // Default values
	
	// Parse pagination parameters if provided
	pageParam := r.URL.Query().Get("page")
	if pageParam != "" {
		page, err = strconv.Atoi(pageParam)
		if err != nil || page < 1 {
			page = 1
		}
	}
	
	pageSizeParam := r.URL.Query().Get("page_size")
	if pageSizeParam != "" {
		pageSize, err = strconv.Atoi(pageSizeParam)
		if err != nil || pageSize < 1 {
			pageSize = 10
		} else if pageSize > 100 {
			pageSize = 100 // Limit maximum page size
		}
	}

	// Get conversation with pagination
	messages, pagination, err := h.messageService.GetConversationPaginated(userID, otherUserID, page, pageSize)
	if err != nil {
		log.Printf("Error getting conversation: %v", err)
		http.Error(w, "Failed to retrieve conversation", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := struct {
		Messages   []models.Message          `json:"messages"`
		Pagination *models.Pagination  `json:"pagination"`
	}{
		Messages:   messages,
		Pagination: pagination,
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
