package main

import (
	"errors"
	"testing"
)

func TestUserService_DeleteAccount(t *testing.T) {
	tests := []struct {
		name          string
		password      string
		mockSetup     func(*MockAPIClient, *MockConfigManager)
		expectedError bool
	}{
		{
			name:     "successful account deletion",
			password: "password123",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				config.SetUserID("1")
				api.DeleteUserFunc = func(userID, password, verifyToken string) error {
					return nil
				}
			},
			expectedError: false,
		},
		{
			name:     "deletion fails - wrong password",
			password: "wrongpassword",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				config.SetUserID("1")
				api.DeleteUserFunc = func(userID, password, verifyToken string) error {
					return errors.New("invalid password")
				}
			},
			expectedError: true,
		},
		{
			name:     "deletion fails - missing user ID",
			password: "password123",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				// Don't set user ID
			},
			expectedError: true,
		},
		{
			name:     "deletion clears credentials on success",
			password: "password123",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				config.SetUserID("1")
				config.SetToken("test-token")
				api.DeleteUserFunc = func(userID, password, verifyToken string) error {
					return nil
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

			service := NewUserService(mockAPI, mockConfig, mockExtractor)
			err := service.DeleteAccount(tt.password)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				// Verify credentials were cleared
				if _, tokenExists := mockConfig.GetToken(); tokenExists {
					t.Errorf("Token should be cleared after account deletion")
				}
				if _, userIDExists := mockConfig.GetUserID(); userIDExists {
					t.Errorf("UserID should be cleared after account deletion")
				}
			}
		})
	}
}

func TestUserService_GetProfile(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*MockAPIClient, *MockConfigManager)
		expectedError bool
		checkData     func(*UserData, *testing.T)
	}{
		{
			name: "successful profile retrieval",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				config.SetUserID("1")
				api.GetUserProfileFunc = func(userID, verifyToken string) (*UserInfoResponse, error) {
					return CreateSuccessUserInfoResponse(), nil
				}
			},
			expectedError: false,
			checkData: func(data *UserData, t *testing.T) {
				if data.Username != "testuser" {
					t.Errorf("Expected username 'testuser' but got '%s'", data.Username)
				}
				if data.Nickname != "Test User" {
					t.Errorf("Expected nickname 'Test User' but got '%s'", data.Nickname)
				}
				if data.Bio != "Test bio" {
					t.Errorf("Expected bio 'Test bio' but got '%s'", data.Bio)
				}
			},
		},
		{
			name: "profile retrieval fails - missing user ID",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				// Don't set user ID
			},
			expectedError: true,
		},
		{
			name: "profile retrieval fails - API error",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				config.SetUserID("1")
				api.GetUserProfileFunc = func(userID, verifyToken string) (*UserInfoResponse, error) {
					return nil, errors.New("network error")
				}
			},
			expectedError: true,
		},
		{
			name: "profile retrieval fails - empty response",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				config.SetUserID("1")
				api.GetUserProfileFunc = func(userID, verifyToken string) (*UserInfoResponse, error) {
					return &UserInfoResponse{Data: nil}, nil
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

			service := NewUserService(mockAPI, mockConfig, mockExtractor)
			data, err := service.GetProfile()

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if data == nil {
					t.Errorf("Expected user data but got nil")
				} else if tt.checkData != nil {
					tt.checkData(data, t)
				}
			}
		})
	}
}

func TestUserService_UpdateProfile(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		value         string
		mockSetup     func(*MockAPIClient, *MockConfigManager)
		expectedError bool
	}{
		{
			name:  "successful nickname update",
			key:   "nickname",
			value: "New Nickname",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				config.SetUserID("1")
				api.UpdateUserProfileFunc = func(userID, key, value, verifyToken string) error {
					return nil
				}
			},
			expectedError: false,
		},
		{
			name:  "successful bio update",
			key:   "bio",
			value: "New bio text",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				config.SetUserID("1")
				api.UpdateUserProfileFunc = func(userID, key, value, verifyToken string) error {
					return nil
				}
			},
			expectedError: false,
		},
		{
			name:  "update fails - missing user ID",
			key:   "nickname",
			value: "New Nickname",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				// Don't set user ID
			},
			expectedError: true,
		},
		{
			name:  "update fails - API error",
			key:   "nickname",
			value: "New Nickname",
			mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
				config.SetUserID("1")
				api.UpdateUserProfileFunc = func(userID, key, value, verifyToken string) error {
					return errors.New("update failed")
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

			service := NewUserService(mockAPI, mockConfig, mockExtractor)
			err := service.UpdateProfile(tt.key, tt.value)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}
