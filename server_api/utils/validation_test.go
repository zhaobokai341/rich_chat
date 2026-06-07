package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name          string
		password      string
		maxLength     int
		expectedError bool
		errorContains string
	}{
		{
			name:          "valid password",
			password:      "Password123",
			maxLength:     128,
			expectedError: false,
		},
		{
			name:          "password exceeds max length",
			password:      string(make([]byte, 200)),
			maxLength:     128,
			expectedError: true,
			errorContains: "password exceeds maximum length",
		},
		{
			name:          "short password with only digits - no error",
			password:      "123456",
			maxLength:     128,
			expectedError: false,
		},
		{
			name:          "short password with only letters - no error",
			password:      "abcdef",
			maxLength:     128,
			expectedError: false,
		},
		{
			name:          "short password with mixed chars - no error",
			password:      "abc123",
			maxLength:     128,
			expectedError: false,
		},
		{
			name:          "empty password - no error",
			password:      "",
			maxLength:     128,
			expectedError: false,
		},
		{
			name:          "valid complex password",
			password:      "MyP@ssw0rd!",
			maxLength:     128,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password, tt.maxLength)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateBio(t *testing.T) {
	tests := []struct {
		name          string
		bio           string
		maxLength     int
		expectedError bool
		errorContains string
	}{
		{
			name:          "valid bio",
			bio:           "Hello, this is my bio",
			maxLength:     500,
			expectedError: false,
		},
		{
			name:          "bio exceeds max length",
			bio:           string(make([]byte, 600)),
			maxLength:     500,
			expectedError: true,
			errorContains: "bio exceeds maximum length",
		},
		{
			name:          "empty bio",
			bio:           "",
			maxLength:     500,
			expectedError: false,
		},
		{
			name:          "bio at max length",
			bio:           string(make([]byte, 500)),
			maxLength:     500,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBio(tt.bio, tt.maxLength)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		maxLength     int
		expectedError bool
		errorContains string
	}{
		{
			name:          "valid email",
			email:         "test@example.com",
			maxLength:     255,
			expectedError: false,
		},
		{
			name:          "invalid email format",
			email:         "invalid-email",
			maxLength:     255,
			expectedError: true,
			errorContains: "invalid email format",
		},
		{
			name:          "email exceeds max length",
			email:         string(make([]byte, 300)) + "@example.com",
			maxLength:     255,
			expectedError: true,
			errorContains: "email exceeds maximum length",
		},
		{
			name:          "email with special characters",
			email:         "test.user+tag@example-domain.com",
			maxLength:     255,
			expectedError: false,
		},
		{
			name:          "empty email",
			email:         "",
			maxLength:     255,
			expectedError: true,
			errorContains: "invalid email format",
		},
		{
			name:          "email without domain",
			email:         "test@",
			maxLength:     255,
			expectedError: true,
			errorContains: "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email, tt.maxLength)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
