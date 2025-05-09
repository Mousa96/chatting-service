// Package handler provides HTTP handlers for authentication operations
package handler

import "net/http"

// Handler defines the authentication handling interface
type Handler interface {
	// Register handles user registration requests
	Register(w http.ResponseWriter, r *http.Request)
	// Login handles user login requests
	Login(w http.ResponseWriter, r *http.Request)
}
