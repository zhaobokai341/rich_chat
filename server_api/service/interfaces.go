package service

import (
	"context"
	"errors"
	"time"

	"rich_chat/server_api/database"
)

// Application errors
var (
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidPassword       = errors.New("invalid password")
	ErrAccountLocked         = errors.New("account is locked")
	ErrInvalidToken          = errors.New("invalid or expired token")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrInvalidInput          = errors.New("invalid input")
)

// LoginRequest represents login credentials
type LoginRequest struct {
	Username    string
	Password    string
	VerifyToken string
}

// LoginResponse represents successful login response
type LoginResponse struct {
	UserID    int
	UserToken string
}

// RegisterRequest represents registration data
type RegisterRequest struct {
	Username    string
	Password    string
	VerifyToken string
}

// RegisterResponse represents successful registration response
type RegisterResponse struct {
	UserID    int
	UserToken string
}

// UserProfileUpdateRequest represents profile update data
type UserProfileUpdateRequest struct {
	UserID int
	Key    string
	Value  string
}

// DeleteUserRequest represents user deletion request
type DeleteUserRequest struct {
	UserID      int
	Password    string
	VerifyToken string
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	UserID      int
	OldPassword string
	NewPassword string
	VerifyToken string
}

// AuthService defines authentication operations
type AuthService interface {
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	GenerateToken(userID int) (string, error)
	GenerateVerifyToken() (string, error)
	ValidateVerifyToken(token string) error
}

// UserService defines user management operations
type UserService interface {
	GetUserProfile(ctx context.Context, userID int) (*database.UserInfo, error)
	UpdateUserProfile(ctx context.Context, req *UserProfileUpdateRequest) error
	ChangeUserPassword(ctx context.Context, req *ChangePasswordRequest) error
	DeleteUser(ctx context.Context, req *DeleteUserRequest) error
	CheckAccountLocked(identifier string) bool
	CheckUserExists(userID int) (bool, error)
}

// TokenService defines token operations
type TokenService interface {
	GenerateJWT(userID int, expiration time.Duration) (string, error)
	GenerateVerificationToken() (string, error)
	StoreVerificationToken(token string, ttl time.Duration) error
	ValidateAndConsumeToken(token string) error
}
