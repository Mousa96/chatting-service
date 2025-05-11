package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

// GetPaginationParams extracts and validates pagination parameters from request
func GetPaginationParams(r *http.Request) (page, pageSize int, err error) {
	page = 1
	pageSize = 10
	
	if p := r.URL.Query().Get("page"); p != "" {
		if page, err = strconv.Atoi(p); err != nil || page < 1 {
			return 0, 0, errors.New("invalid page parameter")
		}
	}
	
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if pageSize, err = strconv.Atoi(ps); err != nil || pageSize < 1 {
			return 0, 0, errors.New("invalid page_size parameter")
		}
		if pageSize > 100 {
			pageSize = 100 // Limit maximum page size
		}
	}
	
	return page, pageSize, nil
}

// WriteJSON sends a JSON response with given status code
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, log and send error response
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
