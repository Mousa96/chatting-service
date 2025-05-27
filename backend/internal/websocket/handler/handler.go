package handler

import (
	"net/http"

	"github.com/Mousa96/chatting-service/internal/websocket/service"
)

// WebSocketHandler implements the Handler interface
type WebSocketHandler struct {
	wsService service.Service
}

// NewWebSocketHandler creates a new WebSocketHandler instance
func NewWebSocketHandler(wsService service.Service) Handler {
	return &WebSocketHandler{
		wsService: wsService,
	}
}

// ServeWS upgrades an HTTP connection to WebSocket and handles the connection
func (h *WebSocketHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	h.wsService.ServeWs(w, r)
}
