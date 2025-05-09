package models

import "time"

type User struct {
    ID           int       `json:"id"`
    Username     string    `json:"username"`
    PasswordHash string    `json:"-"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50"`
    Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
    Username string `json:"username" validate:"required"`
    Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
    Token string `json:"token"`
    User  User   `json:"user"`
} 