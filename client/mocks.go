package main

import (
	"errors"
)

// MockAPIClient implements APIClient interface for testing
type MockAPIClient struct {
	GetVerifyTokenFunc    func() (string, error)
	LoginFunc             func(username, password, verifyToken string) (*AuthResponse, error)
	RegisterFunc          func(username, password, verifyToken string) (*AuthResponse, error)
	DeleteUserFunc        func(userID, token, password, verifyToken string) error
	GetUserProfileFunc    func(userID, token, verifyToken string) (*UserInfoResponse, error)
	UpdateUserProfileFunc func(userID, token, key, value, verifyToken string) error
	ChangePasswordFunc    func(userID, token, oldPassword, newPassword, verifyToken string) error
	CheckServerHealthFunc func() (bool, error)
	SetAuthHeadersFunc    func(token, userID string)
	ClearAuthHeadersFunc  func()
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

func (m *MockAPIClient) DeleteUser(userID, token, password, verifyToken string) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(userID, token, password, verifyToken)
	}
	return nil
}

func (m *MockAPIClient) GetUserProfile(userID, token, verifyToken string) (*UserInfoResponse, error) {
	if m.GetUserProfileFunc != nil {
		return m.GetUserProfileFunc(userID, token, verifyToken)
	}
	return &UserInfoResponse{
		Data: &UserData{
			Username: "testuser",
			Nickname: "Test User",
			Bio:      "Test bio",
		},
	}, nil
}

func (m *MockAPIClient) UpdateUserProfile(userID, token, key, value, verifyToken string) error {
	if m.UpdateUserProfileFunc != nil {
		return m.UpdateUserProfileFunc(userID, token, key, value, verifyToken)
	}
	return nil
}

func (m *MockAPIClient) ChangePassword(userID, token, oldPassword, newPassword, verifyToken string) error {
	if m.ChangePasswordFunc != nil {
		return m.ChangePasswordFunc(userID, token, oldPassword, newPassword, verifyToken)
	}
	return nil
}

func (m *MockAPIClient) CheckServerHealth() (bool, error) {
	if m.CheckServerHealthFunc != nil {
		return m.CheckServerHealthFunc()
	}
	return true, nil
}

func (m *MockAPIClient) SetAuthHeaders(token, userID string) {
	if m.SetAuthHeadersFunc != nil {
		m.SetAuthHeadersFunc(token, userID)
	}
}

func (m *MockAPIClient) ClearAuthHeaders() {
	if m.ClearAuthHeadersFunc != nil {
		m.ClearAuthHeadersFunc()
	}
}

// MockConfigManager implements ConfigManager interface for testing
type MockConfigManager struct {
	ReadConfigFunc        func() (map[string]interface{}, error)
	SaveConfigFunc        func(data map[string]interface{}) error
	GetTokenFunc          func() (string, bool)
	GetUserIDFunc         func() (string, bool)
	SetTokenFunc          func(token string)
	SetUserIDFunc         func(userID string)
	ClearCredentialsFunc  func()
	HasEncryptionKeyFunc  func(userID int) (bool, error)
	SaveEncryptionKeyFunc func(userID int, privateKeyPEM string) error
	GetEncryptionKeyFunc  func(userID int) (string, error)

	data           map[string]interface{}
	encryptionKeys map[int]string
}

func NewMockConfigManager() *MockConfigManager {
	return &MockConfigManager{
		data:           make(map[string]interface{}),
		encryptionKeys: make(map[int]string),
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

func (m *MockConfigManager) HasEncryptionKey(userID int) (bool, error) {
	if m.HasEncryptionKeyFunc != nil {
		return m.HasEncryptionKeyFunc(userID)
	}
	_, exists := m.encryptionKeys[userID]
	return exists, nil
}

func (m *MockConfigManager) SaveEncryptionKey(userID int, privateKeyPEM string) error {
	if m.SaveEncryptionKeyFunc != nil {
		return m.SaveEncryptionKeyFunc(userID, privateKeyPEM)
	}
	m.encryptionKeys[userID] = privateKeyPEM
	return nil
}

func (m *MockConfigManager) GetEncryptionKey(userID int) (string, error) {
	if m.GetEncryptionKeyFunc != nil {
		return m.GetEncryptionKeyFunc(userID)
	}
	key, exists := m.encryptionKeys[userID]
	if !exists {
		return "", errors.New("key not found")
	}
	return key, nil
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

// MockChatAPIClient implements chat API operations for testing
type MockChatAPIClient struct {
	GetUserBasicInfoFunc  func(currentUserID int, token, verifyToken string, targetUserID int) (*UserBasicInfo, error)
	CreateChatSessionFunc func(userID int, token, verifyToken string, recipientID int) (int, error)
	GetUserSessionsFunc   func(userID int, token, verifyToken string) ([]*ChatSession, error)
	GetUserPublicKeyFunc  func(currentUserID int, token, verifyToken string, targetUserID int) (string, string, error)
	StoreUserKeyFunc      func(userID int, token, publicKey string, encryptedPrivateKey []byte, keyAlgorithm string, verifyToken string) error
}

func (m *MockChatAPIClient) GetUserBasicInfo(currentUserID int, token, verifyToken string, targetUserID int) (*UserBasicInfo, error) {
	if m.GetUserBasicInfoFunc != nil {
		return m.GetUserBasicInfoFunc(currentUserID, token, verifyToken, targetUserID)
	}
	return &UserBasicInfo{ID: targetUserID, Username: "testuser"}, nil
}

func (m *MockChatAPIClient) CreateChatSession(userID int, token, verifyToken string, recipientID int) (int, error) {
	if m.CreateChatSessionFunc != nil {
		return m.CreateChatSessionFunc(userID, token, verifyToken, recipientID)
	}
	return 1, nil
}

func (m *MockChatAPIClient) GetUserSessions(userID int, token, verifyToken string) ([]*ChatSession, error) {
	if m.GetUserSessionsFunc != nil {
		return m.GetUserSessionsFunc(userID, token, verifyToken)
	}
	return []*ChatSession{}, nil
}

func (m *MockChatAPIClient) GetUserPublicKey(currentUserID int, token, verifyToken string, targetUserID int) (string, string, error) {
	if m.GetUserPublicKeyFunc != nil {
		return m.GetUserPublicKeyFunc(currentUserID, token, verifyToken, targetUserID)
	}
	return "", "", errors.New("public key not found")
}

func (m *MockChatAPIClient) StoreUserKey(userID int, token, publicKey string, encryptedPrivateKey []byte, keyAlgorithm string, verifyToken string) error {
	if m.StoreUserKeyFunc != nil {
		return m.StoreUserKeyFunc(userID, token, publicKey, encryptedPrivateKey, keyAlgorithm, verifyToken)
	}
	return nil
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
