package service

import "net/http"

type Service interface {
	ServeWs(w http.ResponseWriter, r *http.Request)
}
