package main

import (
	"testing"
)

func TestJWTTokenExtractor_ExtractUserID(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		expectedID    string
		expectedError bool
	}{
		{
			name:          "extract from valid JWT with numeric user_id",
			token:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjN9.abc",
			expectedID:    "123",
			expectedError: false,
		},
		{
			name:          "extract from valid JWT with string user_id",
			token:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNDU2In0.abc",
			expectedID:    "456",
			expectedError: false,
		},
		{
			name:          "invalid token format - wrong number of parts",
			token:         "invalid.token",
			expectedID:    "",
			expectedError: true,
		},
		{
			name:          "invalid token format - empty string",
			token:         "",
			expectedID:    "",
			expectedError: true,
		},
		{
			name:          "token without user_id claim",
			token:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.abc",
			expectedID:    "",
			expectedError: true,
		},
	}

	extractor := NewJWTTokenExtractor()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := extractor.ExtractUserID(tt.token)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if userID != tt.expectedID {
					t.Errorf("Expected userID '%s' but got '%s'", tt.expectedID, userID)
				}
			}
		})
	}
}
