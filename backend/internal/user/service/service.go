package service

import (
	"github.com/Mousa96/chatting-service/internal/user/models"
	"github.com/Mousa96/chatting-service/internal/user/repository"
	wsModels "github.com/Mousa96/chatting-service/internal/websocket/models"
	websocketService "github.com/Mousa96/chatting-service/internal/websocket/service"
)

// UserService implements Service interface
type UserService struct {
	repo       repository.Repository
	wsService  websocketService.Service
}

// NewUserService creates a new UserService
func NewUserService(repo repository.Repository, wsService websocketService.Service) Service {
	return &UserService{repo: repo, wsService: wsService}
}

// GetAllUsers retrieves all users
func (s *UserService) GetAllUsers() ([]models.User, error) {
	users, err := s.repo.GetAllUsers()
	if err != nil {
		return nil, err
	}
	
	// Enhance with online status from WebSocket service if available
	if s.wsService != nil {
		connectedUsers, err := s.wsService.GetConnectedUsers()
		if err == nil {
			// Create a map for faster lookup
			onlineMap := make(map[int]bool)
			for _, id := range connectedUsers {
				onlineMap[id] = true
			}
			
			// Update user status
			for i := range users {
				if onlineMap[users[i].ID] {
					users[i].Status = "online"
				}
			}
		}
	}
	
	return users, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id int) (*models.User, error) {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	
	// Default status to offline
	user.Status = "offline"
	
	// Check if user is online
	if s.wsService != nil {
		status, err := s.wsService.GetUserStatus(id)
		// Only update if we get a valid online status
		if err == nil && status == wsModels.StatusOnline {
			user.Status = "online"
		}
	}
	
	return user, nil
}

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	return s.repo.GetUserByUsername(username)
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(user *models.User) error {
	return s.repo.UpdateUser(user)
}

// UpdateUserStatus updates a user's status
func (s *UserService) UpdateUserStatus(userID int, status string) error {
	// Update in repository
	if err := s.repo.UpdateUserStatus(userID, status); err != nil {
		return err
	}
	
	// If WebSocket service is available, update status there too
	if s.wsService != nil {
		// Convert string to appropriate WebSocket status type
		var wsStatus wsModels.UserStatus
		switch status {
		case "online":
			wsStatus = wsModels.StatusOnline
		case "offline":
			wsStatus = wsModels.StatusOffline
		default:
			wsStatus = wsModels.StatusAway
		}
		
		if err := s.wsService.UpdateUserStatus(userID, wsStatus); err != nil {
			return err
		}
	}
	
	return nil
} 