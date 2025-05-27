package handler

import (
	"net/http"

	"github.com/Mousa96/chatting-service/internal/websocket/service"
)

// WebSocketHandler implements the Handler interface
type WebSocketHandler struct {
	wsService service.Service
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(wsService service.Service) Handler {
	return &WebSocketHandler{
		wsService: wsService,
	}
}

// ServeWS godoc
// @Summary Establish WebSocket connection
// @Description Upgrades HTTP connection to WebSocket for real-time messaging. Supports message types: chat, status, typing
// @Tags websocket
// @Accept json
// @Produce json
// @Security Bearer
// @Success 101 {string} string "Switching Protocols - Connection established"
// @Success 200 {string} string "Message received/sent successfully"
// @Failure 400 {string} string "Bad Request - Invalid message format"
// @Failure 401 {string} string "Unauthorized - Invalid or missing token"
// @Router /ws [get]
func (h *WebSocketHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	h.wsService.ServeWs(w, r)
}