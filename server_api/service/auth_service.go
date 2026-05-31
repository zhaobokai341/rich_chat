package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"rich_chat/server_api/database"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	MaxUsernameLength int
	VerifyTokenTTL    time.Duration
}

// AuthServiceImpl implements AuthService
type AuthServiceImpl struct {
	userRepo      database.UserRepository
	rateLimitRepo database.RateLimitRepository
	tokenService  TokenService
	config        AuthConfig
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo database.UserRepository,
	rateLimitRepo database.RateLimitRepository,
	tokenService TokenService,
	config AuthConfig,
) *AuthServiceImpl {
	return &AuthServiceImpl{
		userRepo:      userRepo,
		rateLimitRepo: rateLimitRepo,
		tokenService:  tokenService,
		config:        config,
	}
}

// Login authenticates a user and returns JWT token
func (s *AuthServiceImpl) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Validate input
	if req.Username == "" || req.Password == "" {
		return nil, ErrInvalidInput
	}

	// Validate verification token
	if err := s.tokenService.ValidateAndConsumeToken(req.VerifyToken); err != nil {
		return nil, err
	}

	// Check if account is locked
	if s.rateLimitRepo != nil {
		locked, _ := s.rateLimitRepo.CheckAccountLocked(req.Username)
		if locked {
			log.WithFields(log.Fields{
				"username": req.Username,
			}).Warning("Login attempt on locked account")
			return nil, ErrAccountLocked
		}
	}

	// Find user by username
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		log.WithFields(log.Fields{
			"username": req.Username,
			"error":    err.Error(),
		}).Warning("User not found during login")
		
		// Track failed login attempt
		if s.rateLimitRepo != nil {
			_ = s.rateLimitRepo.TrackLoginAttempt(req.Username, false)
		}
		
		return nil, ErrInvalidPassword
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		log.WithFields(log.Fields{
			"username": req.Username,
			"user_id":  user.ID,
		}).Warning("Invalid password during login")
		
		// Track failed login attempt
		if s.rateLimitRepo != nil {
			_ = s.rateLimitRepo.TrackLoginAttempt(req.Username, false)
		}
		
		return nil, ErrInvalidPassword
	}

	// Track successful login
	if s.rateLimitRepo != nil {
		_ = s.rateLimitRepo.TrackLoginAttempt(req.Username, true)
	}

	// Update last login
	_ = s.userRepo.UpdateLastLogin(user.ID)

	// Generate JWT token
	token, err := s.tokenService.GenerateJWT(user.ID, 30*24*time.Hour) // 30 days
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": user.ID,
			"error":   err.Error(),
		}).Error("Failed to generate token after successful login")
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	log.WithFields(log.Fields{
		"user_id":  user.ID,
		"username": req.Username,
	}).Info("User logged in successfully")

	return &LoginResponse{
		UserID:    user.ID,
		UserToken: token,
	}, nil
}

// Register creates a new user and returns JWT token
func (s *AuthServiceImpl) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// Validate input
	if req.Username == "" || req.Password == "" {
		return nil, ErrInvalidInput
	}

	// Validate verification token
	if err := s.tokenService.ValidateAndConsumeToken(req.VerifyToken); err != nil {
		return nil, err
	}

	// Validate username length
	if len(req.Username) > s.config.MaxUsernameLength {
		return nil, fmt.Errorf("%w: username must be less than %d characters", 
			ErrInvalidInput, s.config.MaxUsernameLength)
	}

	// Check if username already exists
	exists, err := s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		log.WithFields(log.Fields{
			"username": req.Username,
			"error":    err.Error(),
		}).Error("Failed to check username existence")
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if exists {
		return nil, ErrUsernameAlreadyExists
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.WithFields(log.Fields{
			"username": req.Username,
			"error":    err.Error(),
		}).Error("Failed to hash password")
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	userID, err := s.userRepo.CreateUser(req.Username, string(passwordHash))
	if err != nil {
		// Check for duplicate key error
		if strings.Contains(err.Error(), "duplicate key") || 
		   strings.Contains(err.Error(), "unique constraint") {
			return nil, ErrUsernameAlreadyExists
		}
		
		log.WithFields(log.Fields{
			"username": req.Username,
			"error":    err.Error(),
		}).Error("Failed to create user")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := s.tokenService.GenerateJWT(userID, 30*24*time.Hour) // 30 days
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to generate token after registration")
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	log.WithFields(log.Fields{
		"user_id":  userID,
		"username": req.Username,
	}).Info("User registered successfully")

	return &RegisterResponse{
		UserID:    userID,
		UserToken: token,
	}, nil
}

// GenerateToken generates a JWT token for a user
func (s *AuthServiceImpl) GenerateToken(userID int) (string, error) {
	return s.tokenService.GenerateJWT(userID, 30*24*time.Hour)
}

// GenerateVerifyToken generates a verification token
func (s *AuthServiceImpl) GenerateVerifyToken() (string, error) {
	return s.tokenService.GenerateVerificationToken()
}

// ValidateVerifyToken validates a verification token
func (s *AuthServiceImpl) ValidateVerifyToken(token string) error {
	return s.tokenService.ValidateAndConsumeToken(token)
}
