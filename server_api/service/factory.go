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
	MaxPasswordLength int
	MaxBioLength      int
	MaxEmailLength    int
	MaxNicknameLength int
}

// Services holds all service instances
type Services struct {
	AuthService          AuthService
	UserService          UserService
	TokenService         TokenService
	E2EEncryptionService E2EEncryptionService
	ChatService          ChatService
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
	userService := NewUserService(
		userRepo, rateLimitRepo,
		config.MaxPasswordLength, config.MaxBioLength, config.MaxEmailLength, config.MaxNicknameLength,
	)

	// Create E2EE encryption service
	e2eEncryptionService := NewE2EEncryptionService()

	// Get chat repository from database service
	chatRepo := dbService.GetChatRepository()

	// Create chat service
	chatService := NewChatService(chatRepo, userRepo, e2eEncryptionService)

	return &Services{
		AuthService:          authService,
		UserService:          userService,
		TokenService:         tokenService,
		E2EEncryptionService: e2eEncryptionService,
		ChatService:          chatService,
	}
}
