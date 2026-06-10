package models

import "time"

type User struct {
    ID        string    `json:"id" bson:"_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
    Name      string    `json:"name" bson:"name" example:"John Doe"`
    Email     string    `json:"email" bson:"email" example:"john@example.com"`
    Password  string    `json:"-" bson:"password"`
    Role      string    `json:"role" bson:"role" example:"user"`
    CreatedAt time.Time `json:"created_at" bson:"created_at" example:"2025-01-01T12:00:00Z"`
    UpdatedAt time.Time `json:"updated_at" bson:"updated_at" example:"2025-01-01T12:00:00Z"`
}

// CreateUserRequest represents the payload for creating a user.
type CreateUserRequest struct {
    Name     string `json:"name" binding:"required" example:"Jane Smith"`
    Email    string `json:"email" binding:"required,email" example:"jane@example.com"`
    Password string `json:"password" binding:"required,min=6" example:"secret123"`
    Role     string `json:"role" binding:"omitempty,oneof=user admin" example:"user"`
}

// LoginRequest represents the payload for login.
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email" example:"john@example.com"`
    Password string `json:"password" binding:"required" example:"password123"`
}