package service

import "github.com/Mousa96/chatting-service/internal/user/models"

// Service defines business logic for user operations
type Service interface {
    GetAllUsers() ([]models.User, error)
    GetUserByID(id int) (*models.User, error)
    GetUserByUsername(username string) (*models.User, error)
    UpdateUser(user *models.User) error
    UpdateUserStatus(userID int, status string) error
}
