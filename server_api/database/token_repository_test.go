package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRedisTokenRepository_StoreVerifyToken tests the StoreVerifyToken method
func TestRedisTokenRepository_StoreVerifyToken(t *testing.T) {
	mockCache := new(MockCacheService)
	tokenRepo := NewRedisTokenRepository(mockCache, time.Minute*5)

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
			ttl:         time.Minute * 5,
			expectedErr: nil,
			setupMocks: func() {
				mockCache.On("SetWithTTL", "verify_token:valid-token", "valid", 300).Return() // 300 seconds = 5 minutes
			},
		},
		{
			name:        "token storage with different TTL",
			token:       "different-ttl-token",
			ttl:         time.Minute * 10,
			expectedErr: nil,
			setupMocks: func() {
				mockCache.On("SetWithTTL", "verify_token:different-ttl-token", "valid", 600).Return() // 600 seconds = 10 minutes
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := tokenRepo.StoreVerifyToken(tt.token, tt.ttl)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockCache.AssertExpectations(t)
		})
	}
}

// TestRedisTokenRepository_VerifyAndConsumeToken tests the VerifyAndConsumeToken method
func TestRedisTokenRepository_VerifyAndConsumeToken(t *testing.T) {
	mockCache := new(MockCacheService)
	tokenRepo := NewRedisTokenRepository(mockCache, time.Minute*5)

	tests := []struct {
		name            string
		token           string
		expectedIsValid bool
		expectedError   error
		setupMocks      func()
	}{
		{
			name:            "valid token",
			token:           "valid-token",
			expectedIsValid: true,
			expectedError:   nil,
			setupMocks: func() {
				mockCache.On("Get", "verify_token:valid-token").Return("valid", true)
				mockCache.On("Delete", "verify_token:valid-token").Return()
			},
		},
		{
			name:            "expired token",
			token:           "expired-token",
			expectedIsValid: false,
			expectedError:   nil,
			setupMocks: func() {
				mockCache.On("Get", "verify_token:expired-token").Return("", false)
			},
		},
		{
			name:            "token consumed",
			token:           "consumed-token",
			expectedIsValid: false,
			expectedError:   nil,
			setupMocks: func() {
				mockCache.On("Get", "verify_token:consumed-token").Return("", false)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			isValid, err := tokenRepo.VerifyAndConsumeToken(tt.token)

			assert.Equal(t, tt.expectedIsValid, isValid)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockCache.AssertExpectations(t)
		})
	}
}
