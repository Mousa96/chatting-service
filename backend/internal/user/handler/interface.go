package handler

import "net/http"

// Handler defines HTTP handlers for user operations
type Handler interface {
    GetAllUsers(w http.ResponseWriter, r *http.Request)
    GetUserByID(w http.ResponseWriter, r *http.Request)
    UpdateUserStatus(w http.ResponseWriter, r *http.Request)
}
