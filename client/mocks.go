package main

import (
	"errors"
)

// MockAPIClient implements APIClient interface for testing
type MockAPIClient struct {
	GetVerifyTokenFunc   func() (string, error)
	LoginFunc            func(username, password, verifyToken string) (*AuthResponse, error)
	RegisterFunc         func(username, password, verifyToken string) (*AuthResponse, error)
	DeleteUserFunc       func(userID, password, verifyToken string) error
	GetUserProfileFunc   func(userID, verifyToken string) (*UserInfoResponse, error)
	UpdateUserProfileFunc func(userID, key, value, verifyToken string) error
	ChangePasswordFunc    func(userID, oldPassword, newPassword, verifyToken string) error
	CheckServerHealthFunc func() (bool, error)
}

func (m *MockAPIClient) GetVerifyToken() (string, error) {
	if m.GetVerifyTokenFunc != nil {
		return m.GetVerifyTokenFunc()
	}
	return "mock-verify-token", nil
}

func (m *MockAPIClient) Login(username, password, verifyToken string) (*AuthResponse, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(username, password, verifyToken)
	}
	return &AuthResponse{UserID: 1, UserToken: "mock-token"}, nil
}

func (m *MockAPIClient) Register(username, password, verifyToken string) (*AuthResponse, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(username, password, verifyToken)
	}
	return &AuthResponse{UserID: 1, UserToken: "mock-token"}, nil
}

func (m *MockAPIClient) DeleteUser(userID, password, verifyToken string) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(userID, password, verifyToken)
	}
	return nil
}

func (m *MockAPIClient) GetUserProfile(userID, verifyToken string) (*UserInfoResponse, error) {
	if m.GetUserProfileFunc != nil {
		return m.GetUserProfileFunc(userID, verifyToken)
	}
	return &UserInfoResponse{
		Data: &UserData{
			Username: "testuser",
			Nickname: "Test User",
			Bio:      "Test bio",
		},
	}, nil
}

func (m *MockAPIClient) UpdateUserProfile(userID, key, value, verifyToken string) error {
	if m.UpdateUserProfileFunc != nil {
		return m.UpdateUserProfileFunc(userID, key, value, verifyToken)
	}
	return nil
}

func (m *MockAPIClient) ChangePassword(userID, oldPassword, newPassword, verifyToken string) error {
	if m.ChangePasswordFunc != nil {
		return m.ChangePasswordFunc(userID, oldPassword, newPassword, verifyToken)
	}
	return nil
}

func (m *MockAPIClient) CheckServerHealth() (bool, error) {
	if m.CheckServerHealthFunc != nil {
		return m.CheckServerHealthFunc()
	}
	return true, nil
}

// MockConfigManager implements ConfigManager interface for testing
type MockConfigManager struct {
	ReadConfigFunc   func() (map[string]interface{}, error)
	SaveConfigFunc   func(data map[string]interface{}) error
	GetTokenFunc     func() (string, bool)
	GetUserIDFunc    func() (string, bool)
	SetTokenFunc     func(token string)
	SetUserIDFunc    func(userID string)
	ClearCredentialsFunc func()
	
	data map[string]interface{}
}

func NewMockConfigManager() *MockConfigManager {
	return &MockConfigManager{
		data: make(map[string]interface{}),
	}
}

func (m *MockConfigManager) ReadConfig() (map[string]interface{}, error) {
	if m.ReadConfigFunc != nil {
		return m.ReadConfigFunc()
	}
	return m.data, nil
}

func (m *MockConfigManager) SaveConfig(data map[string]interface{}) error {
	if m.SaveConfigFunc != nil {
		return m.SaveConfigFunc(data)
	}
	m.data = data
	return nil
}

func (m *MockConfigManager) GetToken() (string, bool) {
	if m.GetTokenFunc != nil {
		return m.GetTokenFunc()
	}
	token, exists := m.data["token"]
	if !exists {
		return "", false
	}
	tokenStr, ok := token.(string)
	return tokenStr, ok
}

func (m *MockConfigManager) GetUserID() (string, bool) {
	if m.GetUserIDFunc != nil {
		return m.GetUserIDFunc()
	}
	userID, exists := m.data["user_id"]
	if !exists {
		return "", false
	}
	userIDStr, ok := userID.(string)
	return userIDStr, ok
}

func (m *MockConfigManager) SetToken(token string) {
	if m.SetTokenFunc != nil {
		m.SetTokenFunc(token)
		return
	}
	m.data["token"] = token
}

func (m *MockConfigManager) SetUserID(userID string) {
	if m.SetUserIDFunc != nil {
		m.SetUserIDFunc(userID)
		return
	}
	m.data["user_id"] = userID
}

func (m *MockConfigManager) ClearCredentials() {
	if m.ClearCredentialsFunc != nil {
		m.ClearCredentialsFunc()
		return
	}
	delete(m.data, "token")
	delete(m.data, "user_id")
}

// MockTokenExtractor implements TokenExtractor interface for testing
type MockTokenExtractor struct {
	ExtractUserIDFunc func(token string) (string, error)
}

func (m *MockTokenExtractor) ExtractUserID(token string) (string, error) {
	if m.ExtractUserIDFunc != nil {
		return m.ExtractUserIDFunc(token)
	}
	return "1", nil
}

// Helper functions for creating test responses
func CreateSuccessAuthResponse(userID int, token string) *AuthResponse {
	return &AuthResponse{
		UserID:    userID,
		UserToken: token,
	}
}

func CreateErrorAuthResponse(message string) (*AuthResponse, error) {
	return nil, errors.New(message)
}

func CreateSuccessUserInfoResponse() *UserInfoResponse {
	return &UserInfoResponse{
		Data: &UserData{
			Username: "testuser",
			Nickname: "Test User",
			Bio:      "Test bio",
		},
	}
}
