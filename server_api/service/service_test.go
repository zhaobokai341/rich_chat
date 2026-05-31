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
			name: "invalid password",
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

			if tt.expectedResp != nil {
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedResp.UserID, resp.UserID)
				assert.Equal(t, tt.expectedResp.UserToken, resp.UserToken)
			} else {
				assert.Nil(t, resp)
			}

			if tt.expectedError != nil {
				assert.Error(t, err)
				// Check if error is or contains the expected error
				if !errors.Is(err, tt.expectedError) && err != tt.expectedError {
					assert.Contains(t, err.Error(), tt.expectedError.Error())
				}
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
				UserID:    2,
				UserToken: "jwt-token-here",
			},
			expectedError: nil,
			setupMocks: func() {
				mockTokenService.On("ValidateAndConsumeToken", "valid-token").Return(nil)
				mockUserRepo.On("ExistsByUsername", "newuser").Return(false, nil)
				mockUserRepo.On("CreateUser", "newuser", mock.MatchedBy(func(hash string) bool {
					err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("password123"))
					return err == nil
				})).Return(2, nil)
				mockTokenService.On("GenerateJWT", 2, 30*24*time.Hour).Return("jwt-token-here", nil)
			},
		},
		{
			name: "invalid input - empty username",
			request: &RegisterRequest{
				Username:    "",
				Password:    "password123",
				VerifyToken: "valid-token",
			},
			expectedResp:  nil,
			expectedError: ErrInvalidInput,
			setupMocks:    func() {},
		},
		{
			name: "username too long",
			request: &RegisterRequest{
				Username:    string(make([]byte, 51)),
				Password:    "password123",
				VerifyToken: "valid-token",
			},
			expectedResp:  nil,
			expectedError: ErrInvalidInput,
			setupMocks: func() {
				mockTokenService.On("ValidateAndConsumeToken", "valid-token").Return(nil)
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
			name: "duplicate key error",
			request: &RegisterRequest{
				Username:    "duplicateuser",
				Password:    "password123",
				VerifyToken: "valid-token",
			},
			expectedResp:  nil,
			expectedError: ErrUsernameAlreadyExists,
			setupMocks: func() {
				mockTokenService.On("ValidateAndConsumeToken", "valid-token").Return(nil)
				mockUserRepo.On("ExistsByUsername", "duplicateuser").Return(false, nil)
				mockUserRepo.On("CreateUser", "duplicateuser", mock.Anything).Return(0, errors.New("duplicate key value violates unique constraint"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			resp, err := authService.Register(context.Background(), tt.request)

			if tt.expectedResp != nil {
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedResp.UserID, resp.UserID)
				assert.Equal(t, tt.expectedResp.UserToken, resp.UserToken)
			} else {
				assert.Nil(t, resp)
			}

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
			mockTokenService.AssertExpectations(t)
		})
	}
}

// TestUserServiceImpl_GetUserProfile tests the GetUserProfile method
func TestUserServiceImpl_GetUserProfile(t *testing.T) {
	mockUserRepo := new(database.MockUserRepository)
	mockRateLimitRepo := new(database.MockRateLimitRepository)

	userService := NewUserService(mockUserRepo, mockRateLimitRepo)

	tests := []struct {
		name          string
		userID        int
		expectedInfo  *database.UserInfo
		expectedError error
		setupMocks    func()
	}{
		{
			name:   "successful profile retrieval",
			userID: 1,
			expectedInfo: &database.UserInfo{
				Username: "testuser",
				Nickname: "Test User",
				Bio:      "Test bio",
			},
			expectedError: nil,
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 1).Return(true, nil)
				mockUserRepo.On("GetUserProfile", 1).Return(&database.UserInfo{
					Username: "testuser",
					Nickname: "Test User",
					Bio:      "Test bio",
				}, nil)
			},
		},
		{
			name:          "user not found",
			userID:        999,
			expectedInfo:  nil,
			expectedError: ErrUserNotFound,
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 999).Return(false, nil)
			},
		},
		{
			name:          "database error on existence check",
			userID:        2,
			expectedInfo:  nil,
			expectedError: errors.New("failed to check user existence"),
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 2).Return(false, errors.New("database error"))
			},
		},
		{
			name:          "database error on profile retrieval",
			userID:        3,
			expectedInfo:  nil,
			expectedError: errors.New("failed to get user profile"),
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 3).Return(true, nil)
				mockUserRepo.On("GetUserProfile", 3).Return(nil, errors.New("database error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			info, err := userService.GetUserProfile(context.Background(), tt.userID)

			if tt.expectedInfo != nil {
				assert.NotNil(t, info)
				assert.Equal(t, tt.expectedInfo.Username, info.Username)
				assert.Equal(t, tt.expectedInfo.Nickname, info.Nickname)
				assert.Equal(t, tt.expectedInfo.Bio, info.Bio)
			} else {
				assert.Nil(t, info)
			}

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

// TestUserServiceImpl_UpdateUserProfile tests the UpdateUserProfile method
func TestUserServiceImpl_UpdateUserProfile(t *testing.T) {
	mockUserRepo := new(database.MockUserRepository)
	mockRateLimitRepo := new(database.MockRateLimitRepository)

	userService := NewUserService(mockUserRepo, mockRateLimitRepo)

	tests := []struct {
		name          string
		request       *UserProfileUpdateRequest
		expectedError error
		setupMocks    func()
	}{
		{
			name: "successful profile update",
			request: &UserProfileUpdateRequest{
				UserID: 1,
				Key:    "nickname",
				Value:  "New Nickname",
			},
			expectedError: nil,
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 1).Return(true, nil)
				mockUserRepo.On("UpdateProfile", 1, "nickname", "New Nickname").Return(nil)
			},
		},
		{
			name: "invalid input - empty key",
			request: &UserProfileUpdateRequest{
				UserID: 1,
				Key:    "",
				Value:  "New Nickname",
			},
			expectedError: ErrInvalidInput,
			setupMocks:    func() {},
		},
		{
			name: "invalid input - empty value",
			request: &UserProfileUpdateRequest{
				UserID: 1,
				Key:    "nickname",
				Value:  "",
			},
			expectedError: ErrInvalidInput,
			setupMocks:    func() {},
		},
		{
			name: "user not found",
			request: &UserProfileUpdateRequest{
				UserID: 999,
				Key:    "nickname",
				Value:  "New Nickname",
			},
			expectedError: ErrUserNotFound,
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 999).Return(false, nil)
			},
		},
		{
			name: "database error",
			request: &UserProfileUpdateRequest{
				UserID: 2,
				Key:    "bio",
				Value:  "New bio",
			},
			expectedError: errors.New("failed to update user profile"),
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 2).Return(true, nil)
				mockUserRepo.On("UpdateProfile", 2, "bio", "New bio").Return(errors.New("database error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := userService.UpdateUserProfile(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

// TestUserServiceImpl_DeleteUser tests the DeleteUser method
func TestUserServiceImpl_DeleteUser(t *testing.T) {
	tests := []struct {
		name          string
		request       *DeleteUserRequest
		expectedError error
		setupMocks    func(*database.MockUserRepository, *database.MockRateLimitRepository)
	}{
		{
			name: "successful deletion",
			request: &DeleteUserRequest{
				UserID:      1,
				Password:    "password123",
				VerifyToken: "valid-token",
			},
			expectedError: nil,
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

				mockRateLimitRepo.On("CheckAccountLocked", "1").Return(false, nil)
				mockUserRepo.On("FindByID", 1).Return(&database.User{
					ID:           1,
					Username:     "testuser",
					PasswordHash: string(passwordHash),
				}, nil)
				mockRateLimitRepo.On("TrackLoginAttempt", "1", true).Return(nil)
				mockUserRepo.On("DeleteUser", 1).Return(nil)
			},
		},
		{
			name: "invalid input - empty password",
			request: &DeleteUserRequest{
				UserID:      1,
				Password:    "",
				VerifyToken: "valid-token",
			},
			expectedError: ErrInvalidInput,
			setupMocks:    func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {},
		},
		{
			name: "account locked",
			request: &DeleteUserRequest{
				UserID:      1,
				Password:    "password123",
				VerifyToken: "valid-token",
			},
			expectedError: ErrAccountLocked,
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockRateLimitRepo.On("CheckAccountLocked", "1").Return(true, nil)
			},
		},
		{
			name: "user not found",
			request: &DeleteUserRequest{
				UserID:      999,
				Password:    "password123",
				VerifyToken: "valid-token",
			},
			expectedError: ErrUserNotFound,
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockRateLimitRepo.On("CheckAccountLocked", "999").Return(false, nil)
				mockUserRepo.On("FindByID", 999).Return(nil, errors.New("user not found"))
				mockRateLimitRepo.On("TrackLoginAttempt", "999", false).Return(nil)
			},
		},
		{
			name: "invalid password",
			request: &DeleteUserRequest{
				UserID:      1,
				Password:    "wrongpassword",
				VerifyToken: "valid-token",
			},
			expectedError: ErrInvalidPassword,
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				passwordHash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

				mockRateLimitRepo.On("CheckAccountLocked", "1").Return(false, nil)
				mockUserRepo.On("FindByID", 1).Return(&database.User{
					ID:           1,
					Username:     "testuser",
					PasswordHash: string(passwordHash),
				}, nil)
				mockRateLimitRepo.On("TrackLoginAttempt", "1", false).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(database.MockUserRepository)
			mockRateLimitRepo := new(database.MockRateLimitRepository)
			userService := NewUserService(mockUserRepo, mockRateLimitRepo)

			tt.setupMocks(mockUserRepo, mockRateLimitRepo)

			err := userService.DeleteUser(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
			mockRateLimitRepo.AssertExpectations(t)
		})
	}
}

// TestTokenServiceImpl_GenerateJWT tests JWT token generation
func TestTokenServiceImpl_GenerateJWT(t *testing.T) {
	mockTokenRepo := new(database.MockTokenRepository)

	jwtConfig := JWTConfig{
		Secret:     "test-secret-key-for-jwt-signing",
		Expiration: time.Hour * 24,
		Issuer:     "rich_chat",
	}

	tokenService := NewTokenService(jwtConfig, mockTokenRepo)

	token, err := tokenService.GenerateJWT(1, time.Hour*24)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.IsType(t, "", token)
}

// TestTokenServiceImpl_GenerateVerificationToken tests verification token generation
func TestTokenServiceImpl_GenerateVerificationToken(t *testing.T) {
	mockTokenRepo := new(database.MockTokenRepository)

	jwtConfig := JWTConfig{
		Secret:     "test-secret",
		Expiration: time.Hour,
		Issuer:     "rich_chat",
	}

	tokenService := NewTokenService(jwtConfig, mockTokenRepo)

	token, err := tokenService.GenerateVerificationToken()

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Len(t, token, 64) // 32 bytes = 64 hex characters
}

// TestTokenServiceImpl_StoreVerificationToken tests verification token storage
func TestTokenServiceImpl_StoreVerificationToken(t *testing.T) {
	mockTokenRepo := new(database.MockTokenRepository)

	jwtConfig := JWTConfig{
		Secret:     "test-secret",
		Expiration: time.Hour,
		Issuer:     "rich_chat",
	}

	tokenService := NewTokenService(jwtConfig, mockTokenRepo)

	token := "test-verify-token"
	ttl := time.Minute * 5

	mockTokenRepo.On("StoreVerifyToken", token, ttl).Return(nil)

	err := tokenService.StoreVerificationToken(token, ttl)

	assert.NoError(t, err)
	mockTokenRepo.AssertExpectations(t)
}

// TestTokenServiceImpl_ValidateAndConsumeToken tests token validation
func TestTokenServiceImpl_ValidateAndConsumeToken(t *testing.T) {
	mockTokenRepo := new(database.MockTokenRepository)

	jwtConfig := JWTConfig{
		Secret:     "test-secret",
		Expiration: time.Hour,
		Issuer:     "rich_chat",
	}

	tokenService := NewTokenService(jwtConfig, mockTokenRepo)

	tests := []struct {
		name          string
		token         string
		expectedError error
		setupMocks    func()
	}{
		{
			name:          "empty token",
			token:         "",
			expectedError: ErrInvalidToken,
			setupMocks:    func() {},
		},
		{
			name:          "valid token",
			token:         "valid-token",
			expectedError: nil,
			setupMocks: func() {
				mockTokenRepo.On("VerifyAndConsumeToken", "valid-token").Return(true, nil)
			},
		},
		{
			name:          "invalid token",
			token:         "invalid-token",
			expectedError: ErrInvalidToken,
			setupMocks: func() {
				mockTokenRepo.On("VerifyAndConsumeToken", "invalid-token").Return(false, nil)
			},
		},
		{
			name:          "repository error",
			token:         "error-token",
			expectedError: errors.New("failed to verify token"),
			setupMocks: func() {
				mockTokenRepo.On("VerifyAndConsumeToken", "error-token").Return(false, errors.New("redis error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := tokenService.ValidateAndConsumeToken(tt.token)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if tt.name == "repository error" {
					assert.Contains(t, err.Error(), tt.expectedError.Error())
				} else {
					assert.Equal(t, tt.expectedError, err)
				}
			} else {
				assert.NoError(t, err)
			}

			mockTokenRepo.AssertExpectations(t)
		})
	}
}

// TestUserServiceImpl_CheckAccountLocked tests account lock checking
func TestUserServiceImpl_CheckAccountLocked(t *testing.T) {
	mockUserRepo := new(database.MockUserRepository)
	mockRateLimitRepo := new(database.MockRateLimitRepository)

	userService := NewUserService(mockUserRepo, mockRateLimitRepo)

	tests := []struct {
		name         string
		identifier   string
		expectedLock bool
		setupMocks   func()
	}{
		{
			name:         "account locked",
			identifier:   "lockeduser",
			expectedLock: true,
			setupMocks: func() {
				mockRateLimitRepo.On("CheckAccountLocked", "lockeduser").Return(true, nil)
			},
		},
		{
			name:         "account not locked",
			identifier:   "normaluser",
			expectedLock: false,
			setupMocks: func() {
				mockRateLimitRepo.On("CheckAccountLocked", "normaluser").Return(false, nil)
			},
		},
		{
			name:         "no rate limit repo",
			identifier:   "anyuser",
			expectedLock: false,
			setupMocks:   func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "no rate limit repo" {
				userServiceNoRL := NewUserService(mockUserRepo, nil)
				locked := userServiceNoRL.CheckAccountLocked(tt.identifier)
				assert.Equal(t, tt.expectedLock, locked)
			} else {
				tt.setupMocks()
				locked := userService.CheckAccountLocked(tt.identifier)
				assert.Equal(t, tt.expectedLock, locked)
				mockRateLimitRepo.AssertExpectations(t)
			}
		})
	}
}

// TestUserServiceImpl_CheckUserExists tests user existence checking
func TestUserServiceImpl_CheckUserExists(t *testing.T) {
	mockUserRepo := new(database.MockUserRepository)
	mockRateLimitRepo := new(database.MockRateLimitRepository)

	userService := NewUserService(mockUserRepo, mockRateLimitRepo)

	tests := []struct {
		name          string
		userID        int
		expectedExist bool
		expectedError error
		setupMocks    func()
	}{
		{
			name:          "user exists",
			userID:        1,
			expectedExist: true,
			expectedError: nil,
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 1).Return(true, nil)
			},
		},
		{
			name:          "user does not exist",
			userID:        999,
			expectedExist: false,
			expectedError: nil,
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 999).Return(false, nil)
			},
		},
		{
			name:          "database error",
			userID:        2,
			expectedExist: false,
			expectedError: errors.New("database error"),
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 2).Return(false, errors.New("database error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			exists, err := userService.CheckUserExists(tt.userID)

			assert.Equal(t, tt.expectedExist, exists)
			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}
