package service

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"rich_chat/server_api/database"
)

// MockTokenService is a mock implementation of TokenService for testing
type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GenerateJWT(userID int, expiration time.Duration) (string, error) {
	args := m.Called(userID, expiration)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) GenerateVerificationToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) StoreVerificationToken(token string, ttl time.Duration) error {
	args := m.Called(token, ttl)
	return args.Error(0)
}

func (m *MockTokenService) ValidateAndConsumeToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

// MockAuthService is a mock implementation of AuthService for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*LoginResponse), args.Error(1)
}

func (m *MockAuthService) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*RegisterResponse), args.Error(1)
}

func (m *MockAuthService) GenerateToken(userID int) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GenerateVerifyToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ValidateVerifyToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

// MockUserService is a mock implementation of UserService for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetUserProfile(ctx context.Context, userID int) (*database.UserInfo, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*database.UserInfo), args.Error(1)
}

func (m *MockUserService) UpdateUserProfile(ctx context.Context, req *UserProfileUpdateRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockUserService) ChangeUserPassword(ctx context.Context, req *ChangePasswordRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(ctx context.Context, req *DeleteUserRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockUserService) CheckAccountLocked(identifier string) bool {
	args := m.Called(identifier)
	return args.Bool(0)
}

func (m *MockUserService) CheckUserExists(userID int) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserService) GetUserBasicInfo(ctx context.Context, userID int) (*database.UserBasicInfo, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*database.UserBasicInfo), args.Error(1)
}

// MockChatService is a mock implementation of ChatService for testing
type MockChatService struct {
	mock.Mock
}

func (m *MockChatService) CreateSession(ctx context.Context, req *CreateSessionRequest) (*CreateSessionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CreateSessionResponse), args.Error(1)
}

func (m *MockChatService) JoinSession(ctx context.Context, req *JoinSessionRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockChatService) LeaveSession(ctx context.Context, req *LeaveSessionRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockChatService) GetUserSessions(ctx context.Context, userID int) ([]*SessionInfo, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*SessionInfo), args.Error(1)
}

func (m *MockChatService) StoreUserKey(ctx context.Context, req *StoreUserKeyRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockChatService) GetUserPublicKey(ctx context.Context, req *GetUserKeyRequest) (*GetUserKeyResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*GetUserKeyResponse), args.Error(1)
}

func (m *MockChatService) SendEncryptedMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SendMessageResponse), args.Error(1)
}

func (m *MockChatService) GetOfflineMessages(ctx context.Context, req *GetOfflineMessagesRequest) (*GetOfflineMessagesResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*GetOfflineMessagesResponse), args.Error(1)
}

func (m *MockChatService) MarkMessageDelivered(ctx context.Context, userID, messageID int) error {
	args := m.Called(ctx, userID, messageID)
	return args.Error(0)
}

func (m *MockChatService) GetUndeliveredMessageCount(ctx context.Context, userID int) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockChatService) StoreOfflineEncryptedMessage(ctx context.Context, offlineMsg *database.OfflineEncryptedMessage) (int, error) {
	args := m.Called(ctx, offlineMsg)
	return args.Int(0), args.Error(1)
}
