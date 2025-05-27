package handler

import "net/http"

// Handler defines the WebSocket HTTP handler interface
type Handler interface {
	// ServeWS upgrades an HTTP connection to WebSocket and handles the connection
	ServeWS(w http.ResponseWriter, r *http.Request)
}
