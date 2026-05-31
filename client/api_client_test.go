package main

import (
	"testing"
)

func TestRestAPIClient_CheckServerHealth(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*MockAPIClient)
		expectedOK    bool
		expectedError bool
	}{
		{
			name: "server is healthy",
			mockSetup: func(api *MockAPIClient) {
				api.CheckServerHealthFunc = func() (bool, error) {
					return true, nil
				}
			},
			expectedOK:    true,
			expectedError: false,
		},
		{
			name: "server is unhealthy",
			mockSetup: func(api *MockAPIClient) {
				api.CheckServerHealthFunc = func() (bool, error) {
					return false, nil
				}
			},
			expectedOK:    false,
			expectedError: false,
		},
		{
			name: "health check fails with error",
			mockSetup: func(api *MockAPIClient) {
				api.CheckServerHealthFunc = func() (bool, error) {
					return false, nil
				}
			},
			expectedOK:    false,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &MockAPIClient{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockAPI)
			}

			ok, err := mockAPI.CheckServerHealth()

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			if ok != tt.expectedOK {
				t.Errorf("Expected health status %v but got %v", tt.expectedOK, ok)
			}
		})
	}
}

func TestRestAPIClient_GetVerifyToken(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*MockAPIClient)
		expectedError bool
	}{
		{
			name: "successfully get verify token",
			mockSetup: func(api *MockAPIClient) {
				api.GetVerifyTokenFunc = func() (string, error) {
					return "verify-token-123", nil
				}
			},
			expectedError: false,
		},
		{
			name: "fail to get verify token",
			mockSetup: func(api *MockAPIClient) {
				api.GetVerifyTokenFunc = func() (string, error) {
					return "", nil
				}
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &MockAPIClient{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockAPI)
			}

			token, err := mockAPI.GetVerifyToken()

			if tt.expectedError {
				if err == nil && token == "" {
					t.Errorf("Expected error or non-empty token")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}
