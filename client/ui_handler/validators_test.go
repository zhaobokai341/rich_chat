package ui_handler

import (
	"testing"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "valid password with mixed chars",
			password:    "abc123!@#",
			expectError: false,
		},
		{
			name:        "too short only digits",
			password:    "1234567",
			expectError: true,
		},
		{
			name:        "too short only letters",
			password:    "abcdefg",
			expectError: true,
		},
		{
			name:        "valid long password only digits",
			password:    "123456789",
			expectError: false,
		},
		{
			name:        "valid long password only letters",
			password:    "abcdefghi",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		maxLength   int
		expectError bool
	}{
		{
			name:        "valid username",
			username:    "testuser",
			maxLength:   50,
			expectError: false,
		},
		{
			name:        "empty username",
			username:    "",
			maxLength:   50,
			expectError: true,
		},
		{
			name:        "too long username",
			username:    "this_is_a_very_long_username_that_exceeds_limit",
			maxLength:   10,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUsername(tt.username, tt.maxLength)
			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		expectError bool
	}{
		{
			name:        "valid email",
			email:       "test@example.com",
			expectError: false,
		},
		{
			name:        "empty email",
			email:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmail(tt.email)
			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		fieldName   string
		expectError bool
	}{
		{
			name:        "valid value",
			value:       "test",
			fieldName:   "test_field",
			expectError: false,
		},
		{
			name:        "empty value",
			value:       "",
			fieldName:   "test_field",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequired(tt.value, tt.fieldName)
			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
