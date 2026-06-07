package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"rich_chat/server_api/database"
)

// TestAuthServiceImpl_Login tests the Login method
func TestAuthServiceImpl_Login(t *testing.T) {
	mockUserRepo := new(database.MockUserRepository)
	mockRateLimitRepo := new(database.MockRateLimitRepository)
	mockTokenService := new(MockTokenService)

	config := AuthConfig{
		MaxUsernameLength: 50,
		VerifyTokenTTL:    time.Minute * 5,
	}

	authService := NewAuthService(mockUserRepo, mockRateLimitRepo, mockTokenService, config)

	tests := []struct {
		name          string
		request       *LoginRequest
		expectedResp  *LoginResponse
		expectedError error
		setupMocks    func()
	}{
		{
			name: "successful login",
			request: &LoginRequest{
				Username:    "testuser",
				Password:    "password123",
				VerifyToken: "valid-token",
			},
			expectedResp: &LoginResponse{
				UserID:    1,
				UserToken: "jwt-token-here",
			},
			expectedError: nil,
			setupMocks: func() {
				passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

				mockTokenService.On("ValidateAndConsumeToken", "valid-token").Return(nil)
				mockRateLimitRepo.On("CheckAccountLocked", "testuser").Return(false, nil)
				mockUserRepo.On("FindByUsername", "testuser").Return(&database.User{
					ID:           1,
					Username:     "testuser",
					PasswordHash: string(passwordHash),
				}, nil)
				mockRateLimitRepo.On("TrackLoginAttempt", "testuser", true).Return(nil)
				mockUserRepo.On("UpdateLastLogin", 1).Return(nil)
				mockTokenService.On("GenerateJWT", 1, 30*24*time.Hour).Return("jwt-token-here", nil)
			},
		},
		{
			name: "invalid input - empty username",
			request: &LoginRequest{
				Username:    "",
				Password:    "password123",
				VerifyToken: "valid-token",
			},
			expectedResp:  nil,
			expectedError: ErrInvalidInput,
			setupMocks:    func() {},
		},
		{
			name: "invalid verification token",
			request: &LoginRequest{
				Username:    "testuser",
				Password:    "password123",
				VerifyToken: "invalid-token",
			},
			expectedResp:  nil,
			expectedError: ErrInvalidToken,
			setupMocks: func() {
				mockTokenService.On("ValidateAndConsumeToken", "invalid-token").Return(ErrInvalidToken)
			},
		},
		{
			name: "account locked",
			request: &LoginRequest{
				Username:    "lockeduser",
				Password:    "password123",
				VerifyToken: "valid-token",
			},
			expectedResp:  nil,
			expectedError: ErrAccountLocked,
			setupMocks: func() {
				mockTokenService.On("ValidateAndConsumeToken", "valid-token").Return(nil)
				mockRateLimitRepo.On("CheckAccountLocked", "lockeduser").Return(true, nil)
			},
		},
		{
			name: "user not found",
			request: &LoginRequest{
				Username:    "nonexistent",
				Password:    "password123",
				VerifyToken: "valid-token",
			},
			expectedResp:  nil,
			expectedError: ErrInvalidPassword,
			setupMocks: func() {
				mockTokenService.On("ValidateAndConsumeToken", "valid-token").Return(nil)
				mockRateLimitRepo.On("CheckAccountLocked", "nonexistent").Return(false, nil)
				mockUserRepo.On("FindByUsername", "nonexistent").Return(nil, errors.New("user not found"))
				mockRateLimitRepo.On("TrackLoginAttempt", "nonexistent", false).Return(nil)
			},
		},
		{
			name: "incorrect password",
			request: &LoginRequest{
				Username:    "testuser",
				Password:    "wrongpassword",
				VerifyToken: "valid-token",
			},
			expectedResp:  nil,
			expectedError: ErrInvalidPassword,
			setupMocks: func() {
				passwordHash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

				mockTokenService.On("ValidateAndConsumeToken", "valid-token").Return(nil)
				mockRateLimitRepo.On("CheckAccountLocked", "testuser").Return(false, nil)
				mockUserRepo.On("FindByUsername", "testuser").Return(&database.User{
					ID:           1,
					Username:     "testuser",
					PasswordHash: string(passwordHash),
				}, nil)
				mockRateLimitRepo.On("TrackLoginAttempt", "testuser", false).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			resp, err := authService.Login(context.Background(), tt.request)

			assert.Equal(t, tt.expectedResp, resp)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
			mockRateLimitRepo.AssertExpectations(t)
			mockTokenService.AssertExpectations(t)
		})
	}
}

// TestAuthServiceImpl_Register tests the Register method
func TestAuthServiceImpl_Register(t *testing.T) {
	mockUserRepo := new(database.MockUserRepository)
	mockRateLimitRepo := new(database.MockRateLimitRepository)
	mockTokenService := new(MockTokenService)

	config := AuthConfig{
		MaxUsernameLength: 50,
		VerifyTokenTTL:    time.Minute * 5,
	}

	authService := NewAuthService(mockUserRepo, mockRateLimitRepo, mockTokenService, config)

	tests := []struct {
		name          string
		request       *RegisterRequest
		expectedResp  *RegisterResponse
		expectedError error
		setupMocks    func()
	}{
		{
			name: "successful registration",
			request: &RegisterRequest{
				Username:    "newuser",
				Password:    "password123",
				VerifyToken: "valid-token",
			},
			expectedResp: &RegisterResponse{
				UserID:    1,
				UserToken: "jwt-token-here",
			},
			expectedError: nil,
			setupMocks: func() {
				mockTokenService.On("ValidateAndConsumeToken", "valid-token").Return(nil)
				mockUserRepo.On("ExistsByUsername", "newuser").Return(false, nil)
				mockUserRepo.On("CreateUser", "newuser", mock.AnythingOfType("string")).Return(1, nil)
				mockTokenService.On("GenerateJWT", 1, 30*24*time.Hour).Return("jwt-token-here", nil)
			},
		},
		{
			name: "valid short username registration",
			request: &RegisterRequest{
				Username:    "ab",
				Password:    "password123",
				VerifyToken: "valid-token",
			},
			expectedResp: &RegisterResponse{
				UserID:    2,
				UserToken: "jwt-token-ab",
			},
			expectedError: nil,
			setupMocks: func() {
				mockTokenService.On("ValidateAndConsumeToken", "valid-token").Return(nil)
				mockUserRepo.On("ExistsByUsername", "ab").Return(false, nil)
				mockUserRepo.On("CreateUser", "ab", mock.AnythingOfType("string")).Return(2, nil)
				mockTokenService.On("GenerateJWT", 2, 30*24*time.Hour).Return("jwt-token-ab", nil)
			},
		},
		{
			name: "valid password registration",
			request: &RegisterRequest{
				Username:    "validuser",
				Password:    "weak",
				VerifyToken: "valid-token",
			},
			expectedResp: &RegisterResponse{
				UserID:    3,
				UserToken: "jwt-token-validuser",
			},
			expectedError: nil,
			setupMocks: func() {
				mockTokenService.On("ValidateAndConsumeToken", "valid-token").Return(nil)
				mockUserRepo.On("ExistsByUsername", "validuser").Return(false, nil)
				mockUserRepo.On("CreateUser", "validuser", mock.AnythingOfType("string")).Return(3, nil)
				mockTokenService.On("GenerateJWT", 3, 30*24*time.Hour).Return("jwt-token-validuser", nil)
			},
		},
		{
			name: "invalid verification token",
			request: &RegisterRequest{
				Username:    "newuser",
				Password:    "password123",
				VerifyToken: "invalid-token",
			},
			expectedResp:  nil,
			expectedError: ErrInvalidToken,
			setupMocks: func() {
				mockTokenService.On("ValidateAndConsumeToken", "invalid-token").Return(ErrInvalidToken)
			},
		},
		{
			name: "username already exists",
			request: &RegisterRequest{
				Username:    "existinguser",
				Password:    "password123",
				VerifyToken: "valid-token",
			},
			expectedResp:  nil,
			expectedError: ErrUsernameAlreadyExists,
			setupMocks: func() {
				mockTokenService.On("ValidateAndConsumeToken", "valid-token").Return(nil)
				mockUserRepo.On("ExistsByUsername", "existinguser").Return(true, nil)
			},
		},
		{
			name: "registration fails",
			request: &RegisterRequest{
				Username:    "failuser",
				Password:    "password123",
				VerifyToken: "valid-token",
			},
			expectedResp:  nil,
			expectedError: errors.New("registration failed"),
			setupMocks: func() {
				mockTokenService.On("ValidateAndConsumeToken", "valid-token").Return(nil)
				mockUserRepo.On("ExistsByUsername", "failuser").Return(false, nil)
				mockUserRepo.On("CreateUser", "failuser", mock.AnythingOfType("string")).Return(0, errors.New("registration failed"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			resp, err := authService.Register(context.Background(), tt.request)

			assert.Equal(t, tt.expectedResp, resp)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
			mockRateLimitRepo.AssertExpectations(t)
			mockTokenService.AssertExpectations(t)
		})
	}
}
