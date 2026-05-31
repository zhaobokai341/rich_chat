package service

import (
	"time"

	"rich_chat/server_api/database"
)

// ServiceConfig holds all service configuration
type ServiceConfig struct {
	JWTSecret         string
	JWTExpiration     time.Duration
	MaxUsernameLength int
	VerifyTokenTTL    time.Duration
}

// Services holds all service instances
type Services struct {
	AuthService AuthService
	UserService UserService
	TokenService TokenService
}

// NewServices creates and initializes all services
func NewServices(
	dbService *database.DatabaseService,
	config ServiceConfig,
) *Services {
	// Get repositories from database service
	userRepo := dbService.GetUserRepository()
	rateLimitRepo := dbService.GetRateLimitRepository()
	tokenRepo := dbService.GetTokenRepository()

	// Create JWT config
	jwtConfig := JWTConfig{
		Secret:     config.JWTSecret,
		Expiration: config.JWTExpiration,
		Issuer:     "rich_chat",
	}

	// Create token service
	tokenService := NewTokenService(jwtConfig, tokenRepo)

	// Create auth config
	authConfig := AuthConfig{
		MaxUsernameLength: config.MaxUsernameLength,
		VerifyTokenTTL:    config.VerifyTokenTTL,
	}

	// Create auth service
	authService := NewAuthService(userRepo, rateLimitRepo, tokenService, authConfig)

	// Create user service
	userService := NewUserService(userRepo, rateLimitRepo)

	return &Services{
		AuthService:  authService,
		UserService:  userService,
		TokenService: tokenService,
	}
}
