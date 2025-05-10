// Package handler provides HTTP handlers for WebSocket operations
package handler

import (
	"context"
	"errors"

	"github.com/Mousa96/chatting-service/internal/middleware"
)

// GetUserIDFromContext extracts user ID from the request context
func GetUserIDFromContext(ctx context.Context) (int, error) {
	userID, ok := ctx.Value(middleware.UserIDKey).(int)
	if !ok {
		return 0, errors.New("no user ID in context")
	}
	return userID, nil
} 