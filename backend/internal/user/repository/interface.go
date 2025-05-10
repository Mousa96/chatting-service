package repository

import "github.com/Mousa96/chatting-service/internal/user/models"

// Repository defines data access operations for users
type Repository interface {
    GetAllUsers() ([]models.User, error)
    GetUserByID(id int) (*models.User, error)
    GetUserByUsername(username string) (*models.User, error)
    UpdateUser(user *models.User) error
    UpdateUserStatus(userID int, status string) error
}
