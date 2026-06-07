package service

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"rich_chat/server_api/database"
)

// TestTokenServiceImpl_GenerateJWT tests the GenerateJWT method
func TestTokenServiceImpl_GenerateJWT(t *testing.T) {
	mockTokenRepo := new(database.MockTokenRepository)
	jwtConfig := JWTConfig{
		Secret: "test-secret",
	}

	tokenService := NewTokenService(jwtConfig, mockTokenRepo)

	tests := []struct {
		name           string
		userID         int
		expirationTime time.Duration
		expectedError  error
		setupMocks     func()
	}{
		{
			name:           "successful JWT generation",
			userID:         1,
			expirationTime: time.Hour,
			expectedError:  nil,
			setupMocks:     func() {},
		},
		{
			name:           "negative user ID generates token",
			userID:         -1,
			expirationTime: time.Hour,
			expectedError:  nil,
			setupMocks:     func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			tokenString, err := tokenService.GenerateJWT(tt.userID, tt.expirationTime)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Empty(t, tokenString)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tokenString)

				// Parse the token to verify it's valid
				parsedToken, parseErr := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					return []byte("test-secret"), nil
				})

				assert.NoError(t, parseErr)
				assert.True(t, parsedToken.Valid)

				claims, ok := parsedToken.Claims.(jwt.MapClaims)
				assert.True(t, ok)
				assert.Equal(t, float64(tt.userID), claims["user_id"])
			}
		})
	}
}

// TestTokenServiceImpl_GenerateVerificationToken tests the GenerateVerificationToken method
func TestTokenServiceImpl_GenerateVerificationToken(t *testing.T) {
	mockTokenRepo := new(database.MockTokenRepository)
	jwtConfig := JWTConfig{
		Secret: "test-secret",
	}

	tokenService := NewTokenService(jwtConfig, mockTokenRepo)

	tests := []struct {
		name          string
		expectedError error
		setupMocks    func()
	}{
		{
			name:          "successful verification token generation",
			expectedError: nil,
			setupMocks:    func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			token, err := tokenService.GenerateVerificationToken()

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				// Verification tokens should be 64 characters long (hex encoded 32 bytes)
				assert.Equal(t, 64, len(token))
			}
		})
	}
}

// TestTokenServiceImpl_StoreVerificationToken tests the StoreVerificationToken method
func TestTokenServiceImpl_StoreVerificationToken(t *testing.T) {
	mockTokenRepo := new(database.MockTokenRepository)
	jwtConfig := JWTConfig{
		Secret: "test-secret",
	}

	tokenService := NewTokenService(jwtConfig, mockTokenRepo)

	tests := []struct {
		name        string
		token       string
		ttl         time.Duration
		expectedErr error
		setupMocks  func()
	}{
		{
			name:        "successful token storage",
			token:       "valid-token",
			ttl:         time.Hour,
			expectedErr: nil,
			setupMocks: func() {
				mockTokenRepo.On("StoreVerifyToken", "valid-token", time.Hour).Return(nil)
			},
		},
		{
			name:        "token storage fails",
			token:       "fail-token",
			ttl:         time.Hour,
			expectedErr: errors.New("storage failed"),
			setupMocks: func() {
				mockTokenRepo.On("StoreVerifyToken", "fail-token", time.Hour).Return(errors.New("storage failed"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := tokenService.StoreVerificationToken(tt.token, tt.ttl)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockTokenRepo.AssertExpectations(t)
		})
	}
}

// TestTokenServiceImpl_ValidateAndConsumeToken tests the ValidateAndConsumeToken method
func TestTokenServiceImpl_ValidateAndConsumeToken(t *testing.T) {
	mockTokenRepo := new(database.MockTokenRepository)
	jwtConfig := JWTConfig{
		Secret: "test-secret",
	}

	tokenService := NewTokenService(jwtConfig, mockTokenRepo)

	tests := []struct {
		name          string
		token         string
		expectedError error
		setupMocks    func()
	}{
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
			expectedError: errors.New("token validation failed"),
			setupMocks: func() {
				mockTokenRepo.On("VerifyAndConsumeToken", "invalid-token").Return(false, errors.New("token validation failed"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := tokenService.ValidateAndConsumeToken(tt.token)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockTokenRepo.AssertExpectations(t)
		})
	}
}
