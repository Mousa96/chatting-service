package handler

import "net/http"

type Handler interface {
    SendMessage(w http.ResponseWriter, r *http.Request)
    GetConversation(w http.ResponseWriter, r *http.Request)
} 