package ui_handler

import "github.com/stretchr/testify/mock"

// MockAPIClient is a mock implementation of APIClient for testing
type MockAPIClient struct {
	mock.Mock
}

func (m *MockAPIClient) CheckServerHealth() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockAPIClient) GetVerifyToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockAPIClient) SetAuthHeaders(token, userID string) {
	m.Called(token, userID)
}

func (m *MockAPIClient) ClearAuthHeaders() {
	m.Called()
}

// MockChatAPIClient is a mock implementation of ChatAPIClient for testing
type MockChatAPIClient struct {
	mock.Mock
}

func (m *MockChatAPIClient) GetUserBasicInfo(currentUserID int, token, verifyToken string, targetUserID int) (*UserBasicInfo, error) {
	args := m.Called(currentUserID, token, verifyToken, targetUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserBasicInfo), args.Error(1)
}

func (m *MockChatAPIClient) CreateChatSession(userID int, token, verifyToken string, recipientID int) (int, error) {
	args := m.Called(userID, token, verifyToken, recipientID)
	return args.Int(0), args.Error(1)
}

func (m *MockChatAPIClient) GetUserSessions(userID int, token, verifyToken string) ([]*SessionInfo, error) {
	args := m.Called(userID, token, verifyToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*SessionInfo), args.Error(1)
}

func (m *MockChatAPIClient) GetUserPublicKey(currentUserID int, token, verifyToken string, targetUserID int) (string, string, error) {
	args := m.Called(currentUserID, token, verifyToken, targetUserID)
	return args.String(0), args.String(1), args.Error(2)
}

// MockAuthService is a mock implementation of AuthService for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) IsAuthenticated() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAuthService) Login(username, password string) error {
	args := m.Called(username, password)
	return args.Error(0)
}

func (m *MockAuthService) Register(username, password string) error {
	args := m.Called(username, password)
	return args.Error(0)
}

func (m *MockAuthService) Logout() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAuthService) GetCredentials() (string, string, error) {
	args := m.Called()
	return args.String(0), args.String(1), args.Error(2)
}

// MockUserService is a mock implementation of UserService for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetProfile() (*UserData, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserData), args.Error(1)
}

func (m *MockUserService) UpdateProfile(key, value string) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockUserService) ChangePassword(oldPassword, newPassword string) error {
	args := m.Called(oldPassword, newPassword)
	return args.Error(0)
}

func (m *MockUserService) DeleteAccount(password string) error {
	args := m.Called(password)
	return args.Error(0)
}

// MockChatService is a mock implementation of ChatService for testing
type MockChatService struct {
	mock.Mock
}

func (m *MockChatService) Connect(token string, userID int) error {
	args := m.Called(token, userID)
	return args.Error(0)
}

func (m *MockChatService) Disconnect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockChatService) HasEncryptionKey(userID int) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockChatService) SaveEncryptionKey(userID int, key string) error {
	args := m.Called(userID, key)
	return args.Error(0)
}

func (m *MockChatService) LoadEncryptionKey(userID int) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockChatService) GenerateAndUploadKeys(userID int, token, verifyToken string) error {
	args := m.Called(userID, token, verifyToken)
	return args.Error(0)
}

func (m *MockChatService) UploadPublicKeyOnly(userID int, token, verifyToken string) error {
	args := m.Called(userID, token, verifyToken)
	return args.Error(0)
}

func (m *MockChatService) SendMessage(sessionID, recipientID int, message, recipientPubKeyPEM string) error {
	args := m.Called(sessionID, recipientID, message, recipientPubKeyPEM)
	return args.Error(0)
}

func (m *MockChatService) GetMessages() <-chan *ChatMessage {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(<-chan *ChatMessage)
}

func (m *MockChatService) DecryptMessage(msg *ChatMessage) (string, error) {
	args := m.Called(msg)
	return args.String(0), args.Error(1)
}

// MockConfigManager is a mock implementation of ConfigManager for testing
type MockConfigManager struct {
	mock.Mock
}

func (m *MockConfigManager) ReadConfig() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockConfigManager) GetUserID() (string, bool) {
	args := m.Called()
	return args.String(0), args.Bool(1)
}

func (m *MockConfigManager) GetToken() (string, bool) {
	args := m.Called()
	return args.String(0), args.Bool(1)
}

// MockLanguagePack is a mock implementation of LanguagePack for testing
type MockLanguagePack struct {
	mock.Mock
}

func (m *MockLanguagePack) Get(key string) string {
	args := m.Called(key)
	return args.String(0)
}
