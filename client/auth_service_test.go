package main

import (
	"errors"
	"testing"
)

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		password      string
		mockSetup     func(*MockAPIClient, *MockConfigManager)
		expectedError bool
		errorMessage  string
	}{
		{
			name:     "successful login",
			username: "testuser",
			password: "password123",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				api.LoginFunc = func(username, password, verifyToken string) (*AuthResponse, error) {
					return CreateSuccessAuthResponse(1, "mock-jwt-token"), nil
				}
			},
			expectedError: false,
		},
		{
			name:     "login with invalid credentials",
			username: "wronguser",
			password: "wrongpass",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				api.LoginFunc = func(username, password, verifyToken string) (*AuthResponse, error) {
					return CreateErrorAuthResponse("invalid credentials")
				}
			},
			expectedError: true,
			errorMessage:  "invalid credentials",
		},
		{
			name:     "login when account is locked",
			username: "lockeduser",
			password: "password123",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				api.LoginFunc = func(username, password, verifyToken string) (*AuthResponse, error) {
					return CreateErrorAuthResponse("account_locked_try_later")
				}
			},
			expectedError: true,
			errorMessage:  "account_locked_try_later",
		},
		{
			name:     "login fails to get verify token",
			username: "testuser",
			password: "password123",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				api.GetVerifyTokenFunc = func() (string, error) {
					return "", errors.New("network error")
				}
			},
			expectedError: true,
		},
		{
			name:     "login saves credentials on success",
			username: "testuser",
			password: "password123",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				api.LoginFunc = func(username, password, verifyToken string) (*AuthResponse, error) {
					return CreateSuccessAuthResponse(42, "test-jwt-token"), nil
				}
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &MockAPIClient{}
			mockConfig := NewMockConfigManager()
			mockExtractor := &MockTokenExtractor{}

			if tt.mockSetup != nil {
				tt.mockSetup(mockAPI, mockConfig)
			}

			service := NewAuthService(mockAPI, mockConfig, mockExtractor)
			err := service.Login(tt.username, tt.password)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("Expected error message '%s' but got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				// Verify credentials were saved
				token, tokenExists := mockConfig.GetToken()
				if !tokenExists {
					t.Errorf("Credentials not saved after successful login")
				}
				if tokenExists && token != "mock-jwt-token" && tt.name != "login saves credentials on success" {
					t.Errorf("Expected token 'mock-jwt-token' but got '%s'", token)
				}
			}
		})
	}
}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		password      string
		mockSetup     func(*MockAPIClient, *MockConfigManager)
		expectedError bool
	}{
		{
			name:     "successful registration",
			username: "newuser",
			password: "password123",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				api.RegisterFunc = func(username, password, verifyToken string) (*AuthResponse, error) {
					return CreateSuccessAuthResponse(2, "new-user-token"), nil
				}
			},
			expectedError: false,
		},
		{
			name:     "registration with existing username",
			username: "existinguser",
			password: "password123",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				api.RegisterFunc = func(username, password, verifyToken string) (*AuthResponse, error) {
					return CreateErrorAuthResponse("username_already_exists")
				}
			},
			expectedError: true,
		},
		{
			name:     "registration fails to get verify token",
			username: "newuser",
			password: "password123",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				api.GetVerifyTokenFunc = func() (string, error) {
					return "", errors.New("network error")
				}
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &MockAPIClient{}
			mockConfig := NewMockConfigManager()
			mockExtractor := &MockTokenExtractor{}

			if tt.mockSetup != nil {
				tt.mockSetup(mockAPI, mockConfig)
			}

			service := NewAuthService(mockAPI, mockConfig, mockExtractor)
			err := service.Register(tt.username, tt.password)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				// Verify credentials were saved
				if _, tokenExists := mockConfig.GetToken(); !tokenExists {
					t.Errorf("Credentials not saved after successful registration")
				}
			}
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	mockAPI := &MockAPIClient{}
	mockConfig := NewMockConfigManager()
	mockExtractor := &MockTokenExtractor{}

	// Set initial credentials
	mockConfig.SetToken("test-token")
	mockConfig.SetUserID("1")

	service := NewAuthService(mockAPI, mockConfig, mockExtractor)
	err := service.Logout()

	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	// Verify credentials were cleared
	if _, tokenExists := mockConfig.GetToken(); tokenExists {
		t.Errorf("Token should be cleared after logout")
	}
	if _, userIDExists := mockConfig.GetUserID(); userIDExists {
		t.Errorf("UserID should be cleared after logout")
	}
}

func TestAuthService_IsAuthenticated(t *testing.T) {
	tests := []struct {
		name             string
		setupConfig      func(*MockConfigManager)
		expectedAuth     bool
	}{
		{
			name: "authenticated with valid credentials",
			setupConfig: func(config *MockConfigManager) {
				config.SetToken("valid-token")
				config.SetUserID("1")
			},
			expectedAuth: true,
		},
		{
			name: "not authenticated - missing token",
			setupConfig: func(config *MockConfigManager) {
				config.SetUserID("1")
			},
			expectedAuth: false,
		},
		{
			name: "not authenticated - missing user ID",
			setupConfig: func(config *MockConfigManager) {
				config.SetToken("valid-token")
			},
			expectedAuth: false,
		},
		{
			name: "not authenticated - no credentials",
			setupConfig: func(config *MockConfigManager) {
				// Empty config
			},
			expectedAuth: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &MockAPIClient{}
			mockConfig := NewMockConfigManager()
			mockExtractor := &MockTokenExtractor{}

			tt.setupConfig(mockConfig)

			service := NewAuthService(mockAPI, mockConfig, mockExtractor)
			result := service.IsAuthenticated()

			if result != tt.expectedAuth {
				t.Errorf("Expected authentication status %v but got %v", tt.expectedAuth, result)
			}
		})
	}
}

func TestAuthService_GetCredentials(t *testing.T) {
	tests := []struct {
		name          string
		setupConfig   func(*MockConfigManager)
		expectedError bool
		expectedToken string
		expectedID    string
	}{
		{
			name: "get valid credentials",
			setupConfig: func(config *MockConfigManager) {
				config.SetToken("test-token-123")
				config.SetUserID("42")
			},
			expectedError: false,
			expectedToken: "test-token-123",
			expectedID:    "42",
		},
		{
			name: "get credentials fails - missing token",
			setupConfig: func(config *MockConfigManager) {
				config.SetUserID("42")
			},
			expectedError: true,
		},
		{
			name: "get credentials fails - missing user ID",
			setupConfig: func(config *MockConfigManager) {
				config.SetToken("test-token-123")
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &MockAPIClient{}
			mockConfig := NewMockConfigManager()
			mockExtractor := &MockTokenExtractor{}

			tt.setupConfig(mockConfig)

			service := NewAuthService(mockAPI, mockConfig, mockExtractor)
			token, userID, err := service.GetCredentials()

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if token != tt.expectedToken {
					t.Errorf("Expected token '%s' but got '%s'", tt.expectedToken, token)
				}
				if userID != tt.expectedID {
					t.Errorf("Expected userID '%s' but got '%s'", tt.expectedID, userID)
				}
			}
		})
	}
}
