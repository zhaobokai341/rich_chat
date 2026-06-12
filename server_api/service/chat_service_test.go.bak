package service

import (
	"context"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"rich_chat/server_api/database"
)

// MockChatRepository is a mock implementation of ChatRepository for testing
type MockChatRepository struct {
	CreateChatSessionFunc             func(ctx context.Context, sessionType string, name *string, createdBy *int) (int, error)
	GetChatSessionFunc                func(ctx context.Context, sessionID int) (*database.ChatSession, error)
	AddUserToChatSessionFunc          func(ctx context.Context, sessionID, userID int) error
	RemoveUserFromChatSessionFunc     func(ctx context.Context, sessionID, userID int) error
	GetUsersInChatSessionFunc         func(ctx context.Context, sessionID int) ([]int, error)
	GetUserChatSessionsFunc           func(ctx context.Context, userID int) ([]*database.ChatSession, error)
	CreateMessageIndexFunc            func(ctx context.Context, sessionID, senderID int, messageType string, replyToMessageID *int) (int, error)
	GetMessageIndexFunc               func(ctx context.Context, messageID int) (*database.MessageIndex, error)
	UpdateMessageReadStatusFunc       func(ctx context.Context, messageID int, isRead bool) error
	DeleteMessageFunc                 func(ctx context.Context, messageID int) error
	StoreEncryptedMessageFunc         func(ctx context.Context, encryptedMsg *database.EncryptedMessage) error
	GetEncryptedMessageFunc           func(ctx context.Context, messageID int) (*database.EncryptedMessage, error)
	GetMessagesForSessionFunc         func(ctx context.Context, sessionID int, limit, offset int) ([]*database.MessageIndex, error)
	GetMessagesForUserFunc            func(ctx context.Context, userID int, limit, offset int) ([]*database.MessageIndex, error)
	MarkMessageAsReadByUserFunc       func(ctx context.Context, messageID, userID int) error
	GetUnreadMessagesForUserFunc      func(ctx context.Context, userID int) ([]int, error)
	GetReadReceiptsForMessageFunc     func(ctx context.Context, messageID int) ([]*database.MessageReadReceipt, error)
	CreateGroupChatFunc               func(ctx context.Context, sessionID int, description, avatarURL *string, maxMembers int, privacyLevel string) error
	UpdateGroupChatFunc               func(ctx context.Context, sessionID int, description, avatarURL *string, maxMembers int, privacyLevel string) error
	GetGroupChatFunc                  func(ctx context.Context, sessionID int) (*database.GroupChat, error)
	StoreUserKeyFunc                  func(ctx context.Context, userKey *database.UserKey) error
	GetUserKeyFunc                    func(ctx context.Context, userID int) (*database.UserKey, error)
	GetUserPublicKeyFunc              func(ctx context.Context, userID int) (string, error)
	UpdateUserKeyFunc                 func(ctx context.Context, userKey *database.UserKey) error
	DeactivateUserKeyFunc             func(ctx context.Context, userID int) error
	StoreOfflineEncryptedMessageFunc  func(ctx context.Context, offlineMsg *database.OfflineEncryptedMessage) (int, error)
	GetUndeliveredOfflineMessagesFunc func(ctx context.Context, recipientID int) ([]*database.OfflineEncryptedMessage, error)
	MarkOfflineMessageDeliveredFunc   func(ctx context.Context, messageID int) error
	DeleteOfflineMessageFunc          func(ctx context.Context, messageID int) error
	GetUndeliveredMessageCountFunc    func(ctx context.Context, recipientID int) (int, error)
}

func (m *MockChatRepository) CreateChatSession(ctx context.Context, sessionType string, name *string, createdBy *int) (int, error) {
	return m.CreateChatSessionFunc(ctx, sessionType, name, createdBy)
}

func (m *MockChatRepository) GetChatSession(ctx context.Context, sessionID int) (*database.ChatSession, error) {
	return m.GetChatSessionFunc(ctx, sessionID)
}

func (m *MockChatRepository) AddUserToChatSession(ctx context.Context, sessionID, userID int) error {
	return m.AddUserToChatSessionFunc(ctx, sessionID, userID)
}

func (m *MockChatRepository) RemoveUserFromChatSession(ctx context.Context, sessionID, userID int) error {
	return m.RemoveUserFromChatSessionFunc(ctx, sessionID, userID)
}

func (m *MockChatRepository) GetUsersInChatSession(ctx context.Context, sessionID int) ([]int, error) {
	return m.GetUsersInChatSessionFunc(ctx, sessionID)
}

func (m *MockChatRepository) GetUserChatSessions(ctx context.Context, userID int) ([]*database.ChatSession, error) {
	return m.GetUserChatSessionsFunc(ctx, userID)
}

func (m *MockChatRepository) CreateMessageIndex(ctx context.Context, sessionID, senderID int, messageType string, replyToMessageID *int) (int, error) {
	return m.CreateMessageIndexFunc(ctx, sessionID, senderID, messageType, replyToMessageID)
}

func (m *MockChatRepository) GetMessageIndex(ctx context.Context, messageID int) (*database.MessageIndex, error) {
	return m.GetMessageIndexFunc(ctx, messageID)
}

func (m *MockChatRepository) UpdateMessageReadStatus(ctx context.Context, messageID int, isRead bool) error {
	return m.UpdateMessageReadStatusFunc(ctx, messageID, isRead)
}

func (m *MockChatRepository) DeleteMessage(ctx context.Context, messageID int) error {
	return m.DeleteMessageFunc(ctx, messageID)
}

func (m *MockChatRepository) StoreEncryptedMessage(ctx context.Context, encryptedMsg *database.EncryptedMessage) error {
	return m.StoreEncryptedMessageFunc(ctx, encryptedMsg)
}

func (m *MockChatRepository) GetEncryptedMessage(ctx context.Context, messageID int) (*database.EncryptedMessage, error) {
	return m.GetEncryptedMessageFunc(ctx, messageID)
}

func (m *MockChatRepository) GetMessagesForSession(ctx context.Context, sessionID int, limit, offset int) ([]*database.MessageIndex, error) {
	return m.GetMessagesForSessionFunc(ctx, sessionID, limit, offset)
}

func (m *MockChatRepository) GetMessagesForUser(ctx context.Context, userID int, limit, offset int) ([]*database.MessageIndex, error) {
	return m.GetMessagesForUserFunc(ctx, userID, limit, offset)
}

func (m *MockChatRepository) MarkMessageAsReadByUser(ctx context.Context, messageID, userID int) error {
	return m.MarkMessageAsReadByUserFunc(ctx, messageID, userID)
}

func (m *MockChatRepository) GetUnreadMessagesForUser(ctx context.Context, userID int) ([]int, error) {
	return m.GetUnreadMessagesForUserFunc(ctx, userID)
}

func (m *MockChatRepository) GetReadReceiptsForMessage(ctx context.Context, messageID int) ([]*database.MessageReadReceipt, error) {
	return m.GetReadReceiptsForMessageFunc(ctx, messageID)
}

func (m *MockChatRepository) CreateGroupChat(ctx context.Context, sessionID int, description, avatarURL *string, maxMembers int, privacyLevel string) error {
	return m.CreateGroupChatFunc(ctx, sessionID, description, avatarURL, maxMembers, privacyLevel)
}

func (m *MockChatRepository) UpdateGroupChat(ctx context.Context, sessionID int, description, avatarURL *string, maxMembers int, privacyLevel string) error {
	return m.UpdateGroupChatFunc(ctx, sessionID, description, avatarURL, maxMembers, privacyLevel)
}

func (m *MockChatRepository) GetGroupChat(ctx context.Context, sessionID int) (*database.GroupChat, error) {
	return m.GetGroupChatFunc(ctx, sessionID)
}

func (m *MockChatRepository) StoreUserKey(ctx context.Context, userKey *database.UserKey) error {
	return m.StoreUserKeyFunc(ctx, userKey)
}

func (m *MockChatRepository) GetUserKey(ctx context.Context, userID int) (*database.UserKey, error) {
	return m.GetUserKeyFunc(ctx, userID)
}

func (m *MockChatRepository) GetUserPublicKey(ctx context.Context, userID int) (string, error) {
	return m.GetUserPublicKeyFunc(ctx, userID)
}

func (m *MockChatRepository) UpdateUserKey(ctx context.Context, userKey *database.UserKey) error {
	return m.UpdateUserKeyFunc(ctx, userKey)
}

func (m *MockChatRepository) DeactivateUserKey(ctx context.Context, userID int) error {
	return m.DeactivateUserKeyFunc(ctx, userID)
}

func (m *MockChatRepository) StoreOfflineEncryptedMessage(ctx context.Context, offlineMsg *database.OfflineEncryptedMessage) (int, error) {
	return m.StoreOfflineEncryptedMessageFunc(ctx, offlineMsg)
}

func (m *MockChatRepository) GetUndeliveredOfflineMessages(ctx context.Context, recipientID int) ([]*database.OfflineEncryptedMessage, error) {
	return m.GetUndeliveredOfflineMessagesFunc(ctx, recipientID)
}

func (m *MockChatRepository) MarkOfflineMessageDelivered(ctx context.Context, messageID int) error {
	return m.MarkOfflineMessageDeliveredFunc(ctx, messageID)
}

func (m *MockChatRepository) DeleteOfflineMessage(ctx context.Context, messageID int) error {
	return m.DeleteOfflineMessageFunc(ctx, messageID)
}

func (m *MockChatRepository) GetUndeliveredMessageCount(ctx context.Context, recipientID int) (int, error) {
	return m.GetUndeliveredMessageCountFunc(ctx, recipientID)
}

// MockUserRepository is a mock implementation of UserRepository for testing
type MockUserRepository struct {
	CreateUserFunc       func(username, passwordHash string) (int, error)
	FindByIDFunc         func(id int) (*database.User, error)
	FindByUsernameFunc   func(username string) (*database.User, error)
	ExistsByIDFunc       func(id int) (bool, error)
	ExistsByUsernameFunc func(username string) (bool, error)
	GetUserProfileFunc   func(userID int) (*database.UserInfo, error)
	GetUserBasicInfoFunc func(userID int) (*database.UserBasicInfo, error)
	UpdateProfileFunc    func(userID int, key, value string) error
	UpdateLastLoginFunc  func(userID int) error
	UpdateLockStatusFunc func(identifier string, lockUntil *time.Time) error
	UpdatePasswordFunc   func(userID int, newPasswordHash string) error
	DeleteUserFunc       func(userID int) error
	GetLockStatusFunc    func(identifier string) (*time.Time, error)
	ClearExpiredLockFunc func(identifier string) error
}

func (m *MockUserRepository) CreateUser(username, passwordHash string) (int, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(username, passwordHash)
	}
	return 1, nil
}

func (m *MockUserRepository) FindByID(id int) (*database.User, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(id)
	}
	return &database.User{ID: id, Username: "testuser"}, nil
}

func (m *MockUserRepository) FindByUsername(username string) (*database.User, error) {
	if m.FindByUsernameFunc != nil {
		return m.FindByUsernameFunc(username)
	}
	return &database.User{ID: 1, Username: username}, nil
}

func (m *MockUserRepository) ExistsByID(id int) (bool, error) {
	if m.ExistsByIDFunc != nil {
		return m.ExistsByIDFunc(id)
	}
	return true, nil
}

func (m *MockUserRepository) ExistsByUsername(username string) (bool, error) {
	if m.ExistsByUsernameFunc != nil {
		return m.ExistsByUsernameFunc(username)
	}
	return true, nil
}

func (m *MockUserRepository) GetUserProfile(userID int) (*database.UserInfo, error) {
	if m.GetUserProfileFunc != nil {
		return m.GetUserProfileFunc(userID)
	}
	return &database.UserInfo{Username: "testuser"}, nil
}

func (m *MockUserRepository) GetUserBasicInfo(userID int) (*database.UserBasicInfo, error) {
	if m.GetUserBasicInfoFunc != nil {
		return m.GetUserBasicInfoFunc(userID)
	}
	return &database.UserBasicInfo{
		ID:       userID,
		Username: "testuser",
		Nickname: "Test User",
	}, nil
}

func (m *MockUserRepository) UpdateProfile(userID int, key, value string) error {
	if m.UpdateProfileFunc != nil {
		return m.UpdateProfileFunc(userID, key, value)
	}
	return nil
}

func (m *MockUserRepository) UpdateLastLogin(userID int) error {
	if m.UpdateLastLoginFunc != nil {
		return m.UpdateLastLoginFunc(userID)
	}
	return nil
}

func (m *MockUserRepository) UpdateLockStatus(identifier string, lockUntil *time.Time) error {
	if m.UpdateLockStatusFunc != nil {
		return m.UpdateLockStatusFunc(identifier, lockUntil)
	}
	return nil
}

func (m *MockUserRepository) UpdatePassword(userID int, newPasswordHash string) error {
	if m.UpdatePasswordFunc != nil {
		return m.UpdatePasswordFunc(userID, newPasswordHash)
	}
	return nil
}

func (m *MockUserRepository) DeleteUser(userID int) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(userID)
	}
	return nil
}

func (m *MockUserRepository) GetLockStatus(identifier string) (*time.Time, error) {
	if m.GetLockStatusFunc != nil {
		return m.GetLockStatusFunc(identifier)
	}
	return nil, nil
}

func (m *MockUserRepository) ClearExpiredLock(identifier string) error {
	if m.ClearExpiredLockFunc != nil {
		return m.ClearExpiredLockFunc(identifier)
	}
	return nil
}

// MockE2EEncryptionService is a mock implementation of E2EEncryptionService for testing
type MockE2EEncryptionService struct {
	GenerateRSAKeyPairFunc        func() (string, string, error)
	EncryptMessageFunc            func(plaintext []byte, recipientPublicKeyPEM string) ([]byte, []byte, []byte, []byte, error)
	DecryptMessageFunc            func(encryptedSessionKey []byte, encryptedContent []byte, iv []byte, authTag []byte, privateKeyPEM string) ([]byte, error)
	EncryptWithAESFunc            func(plaintext []byte, key []byte) ([]byte, []byte, []byte, error)
	DecryptWithAESFunc            func(encrypted []byte, iv []byte, authTag []byte, key []byte) ([]byte, error)
	EncryptSessionKeyFunc         func(sessionKey []byte, publicKeyPEM string) ([]byte, error)
	DecryptSessionKeyFunc         func(encryptedSessionKey []byte, privateKeyPEM string) ([]byte, error)
	ParseRSAPublicKeyFromPEMFunc  func(pemStr string) (*rsa.PublicKey, error)
	ParseRSAPrivateKeyFromPEMFunc func(pemStr string) (*rsa.PrivateKey, error)
}

func (m *MockE2EEncryptionService) GenerateRSAKeyPair() (string, string, error) {
	return m.GenerateRSAKeyPairFunc()
}

func (m *MockE2EEncryptionService) EncryptMessage(plaintext []byte, recipientPublicKeyPEM string) ([]byte, []byte, []byte, []byte, error) {
	return m.EncryptMessageFunc(plaintext, recipientPublicKeyPEM)
}

func (m *MockE2EEncryptionService) DecryptMessage(encryptedSessionKey []byte, encryptedContent []byte, iv []byte, authTag []byte, privateKeyPEM string) ([]byte, error) {
	return m.DecryptMessageFunc(encryptedSessionKey, encryptedContent, iv, authTag, privateKeyPEM)
}

func (m *MockE2EEncryptionService) EncryptWithAES(plaintext []byte, key []byte) ([]byte, []byte, []byte, error) {
	return m.EncryptWithAESFunc(plaintext, key)
}

func (m *MockE2EEncryptionService) DecryptWithAES(encrypted []byte, iv []byte, authTag []byte, key []byte) ([]byte, error) {
	return m.DecryptWithAESFunc(encrypted, iv, authTag, key)
}

func (m *MockE2EEncryptionService) EncryptSessionKey(sessionKey []byte, publicKeyPEM string) ([]byte, error) {
	return m.EncryptSessionKeyFunc(sessionKey, publicKeyPEM)
}

func (m *MockE2EEncryptionService) DecryptSessionKey(encryptedSessionKey []byte, privateKeyPEM string) ([]byte, error) {
	return m.DecryptSessionKeyFunc(encryptedSessionKey, privateKeyPEM)
}

func (m *MockE2EEncryptionService) ParseRSAPublicKeyFromPEM(pemStr string) (*rsa.PublicKey, error) {
	return m.ParseRSAPublicKeyFromPEMFunc(pemStr)
}

func (m *MockE2EEncryptionService) ParseRSAPrivateKeyFromPEM(pemStr string) (*rsa.PrivateKey, error) {
	return m.ParseRSAPrivateKeyFromPEMFunc(pemStr)
}

func TestCreateSession(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - Direct Session", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			CreateChatSessionFunc: func(ctx context.Context, sessionType string, name *string, createdBy *int) (int, error) {
				assert.Equal(t, "direct", sessionType)
				return 1, nil
			},
			AddUserToChatSessionFunc: func(ctx context.Context, sessionID, userID int) error {
				return nil
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		name := "Test Session"
		req := &CreateSessionRequest{
			CreatorID:   1,
			SessionType: "direct",
			Name:        &name,
			MemberIDs:   []int{2, 3},
		}

		resp, err := service.CreateSession(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 1, resp.SessionID)
		assert.Equal(t, "Test Session", resp.Name)
	})

	t.Run("Success - Group Session", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			CreateChatSessionFunc: func(ctx context.Context, sessionType string, name *string, createdBy *int) (int, error) {
				assert.Equal(t, "group", sessionType)
				return 2, nil
			},
			AddUserToChatSessionFunc: func(ctx context.Context, sessionID, userID int) error {
				return nil
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &CreateSessionRequest{
			CreatorID:   1,
			SessionType: "group",
			Name:        nil,
			MemberIDs:   []int{2},
		}

		resp, err := service.CreateSession(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 2, resp.SessionID)
	})

	t.Run("Invalid Session Type", func(t *testing.T) {
		mockRepo := &MockChatRepository{}
		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &CreateSessionRequest{
			CreatorID:   1,
			SessionType: "invalid",
		}

		resp, err := service.CreateSession(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid session type")
	})

	t.Run("Create Session Fails", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			CreateChatSessionFunc: func(ctx context.Context, sessionType string, name *string, createdBy *int) (int, error) {
				return 0, errors.New("database error")
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &CreateSessionRequest{
			CreatorID:   1,
			SessionType: "direct",
		}

		resp, err := service.CreateSession(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to create chat session")
	})

	t.Run("Add Creator To Session Fails", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			CreateChatSessionFunc: func(ctx context.Context, sessionType string, name *string, createdBy *int) (int, error) {
				return 1, nil
			},
			AddUserToChatSessionFunc: func(ctx context.Context, sessionID, userID int) error {
				if userID == 1 {
					return errors.New("failed to add creator")
				}
				return nil
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &CreateSessionRequest{
			CreatorID:   1,
			SessionType: "direct",
		}

		resp, err := service.CreateSession(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to add creator to session")
	})
}

func TestJoinSession(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			GetChatSessionFunc: func(ctx context.Context, sessionID int) (*database.ChatSession, error) {
				return &database.ChatSession{ID: 1}, nil
			},
			AddUserToChatSessionFunc: func(ctx context.Context, sessionID, userID int) error {
				return nil
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &JoinSessionRequest{
			UserID:    2,
			SessionID: 1,
		}

		err := service.JoinSession(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("Session Not Found", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			GetChatSessionFunc: func(ctx context.Context, sessionID int) (*database.ChatSession, error) {
				return nil, errors.New("session not found")
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &JoinSessionRequest{
			UserID:    2,
			SessionID: 999,
		}

		err := service.JoinSession(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("Add User Fails", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			GetChatSessionFunc: func(ctx context.Context, sessionID int) (*database.ChatSession, error) {
				return &database.ChatSession{ID: 1}, nil
			},
			AddUserToChatSessionFunc: func(ctx context.Context, sessionID, userID int) error {
				return errors.New("failed to add user")
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &JoinSessionRequest{
			UserID:    2,
			SessionID: 1,
		}

		err := service.JoinSession(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to join session")
	})
}

func TestLeaveSession(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			RemoveUserFromChatSessionFunc: func(ctx context.Context, sessionID, userID int) error {
				return nil
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &LeaveSessionRequest{
			UserID:    2,
			SessionID: 1,
		}

		err := service.LeaveSession(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("Remove User Fails", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			RemoveUserFromChatSessionFunc: func(ctx context.Context, sessionID, userID int) error {
				return errors.New("failed to remove user")
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &LeaveSessionRequest{
			UserID:    2,
			SessionID: 1,
		}

		err := service.LeaveSession(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to leave session")
	})
}

func TestGetUserSessions(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		now := time.Now()
		mockRepo := &MockChatRepository{
			GetUserChatSessionsFunc: func(ctx context.Context, userID int) ([]*database.ChatSession, error) {
				return []*database.ChatSession{
					{ID: 1, SessionType: "direct", Name: strPtr("Session 1"), CreatedAt: &now},
					{ID: 2, SessionType: "group", Name: strPtr("Session 2"), CreatedAt: &now},
				}, nil
			},
			GetUsersInChatSessionFunc: func(ctx context.Context, sessionID int) ([]int, error) {
				if sessionID == 1 {
					return []int{1, 2}, nil
				}
				return []int{1, 2, 3}, nil
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		sessions, err := service.GetUserSessions(ctx, 1)
		assert.NoError(t, err)
		assert.Len(t, sessions, 2)
		assert.Equal(t, 1, sessions[0].SessionID)
		assert.Equal(t, 2, sessions[0].MemberCount)
		assert.Equal(t, 2, sessions[1].SessionID)
		assert.Equal(t, 3, sessions[1].MemberCount)
	})

	t.Run("Get Sessions Fails", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			GetUserChatSessionsFunc: func(ctx context.Context, userID int) ([]*database.ChatSession, error) {
				return nil, errors.New("database error")
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		sessions, err := service.GetUserSessions(ctx, 1)
		assert.Error(t, err)
		assert.Nil(t, sessions)
		assert.Contains(t, err.Error(), "failed to get user sessions")
	})
}

func TestStoreUserKey(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - New Key", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			GetUserKeyFunc: func(ctx context.Context, userID int) (*database.UserKey, error) {
				return nil, errors.New("key not found")
			},
			StoreUserKeyFunc: func(ctx context.Context, userKey *database.UserKey) error {
				return nil
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &StoreUserKeyRequest{
			UserID:       1,
			PublicKey:    "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----",
			KeyAlgorithm: "RSA-2048",
		}

		err := service.StoreUserKey(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("Success - Replace Existing Key", func(t *testing.T) {
		now := time.Now()
		mockRepo := &MockChatRepository{
			GetUserKeyFunc: func(ctx context.Context, userID int) (*database.UserKey, error) {
				return &database.UserKey{
					ID:        1,
					UserID:    userID,
					IsActive:  true,
					CreatedAt: &now,
				}, nil
			},
			DeactivateUserKeyFunc: func(ctx context.Context, userID int) error {
				return nil
			},
			StoreUserKeyFunc: func(ctx context.Context, userKey *database.UserKey) error {
				return nil
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &StoreUserKeyRequest{
			UserID:       1,
			PublicKey:    "-----BEGIN PUBLIC KEY-----\nnew_key\n-----END PUBLIC KEY-----",
			KeyAlgorithm: "RSA-2048",
		}

		err := service.StoreUserKey(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("Missing Public Key", func(t *testing.T) {
		mockRepo := &MockChatRepository{}
		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &StoreUserKeyRequest{
			UserID:       1,
			PublicKey:    "",
			KeyAlgorithm: "RSA-2048",
		}

		err := service.StoreUserKey(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "public key is required")
	})

	t.Run("Missing Key Algorithm", func(t *testing.T) {
		mockRepo := &MockChatRepository{}
		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &StoreUserKeyRequest{
			UserID:       1,
			PublicKey:    "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----",
			KeyAlgorithm: "",
		}

		err := service.StoreUserKey(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key algorithm is required")
	})

	t.Run("Store Key Fails", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			GetUserKeyFunc: func(ctx context.Context, userID int) (*database.UserKey, error) {
				return nil, errors.New("key not found")
			},
			StoreUserKeyFunc: func(ctx context.Context, userKey *database.UserKey) error {
				return errors.New("database error")
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &StoreUserKeyRequest{
			UserID:       1,
			PublicKey:    "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----",
			KeyAlgorithm: "RSA-2048",
		}

		err := service.StoreUserKey(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to store user key")
	})
}

func TestGetUserPublicKey(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		now := time.Now()
		mockRepo := &MockChatRepository{
			GetUserKeyFunc: func(ctx context.Context, userID int) (*database.UserKey, error) {
				return &database.UserKey{
					ID:           1,
					UserID:       userID,
					PublicKey:    "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----",
					KeyAlgorithm: "RSA-2048",
					IsActive:     true,
					CreatedAt:    &now,
				}, nil
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &GetUserKeyRequest{
			UserID: 1,
		}

		resp, err := service.GetUserPublicKey(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 1, resp.UserID)
		assert.Equal(t, "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----", resp.PublicKey)
		assert.Equal(t, "RSA-2048", resp.Algorithm)
	})

	t.Run("Key Not Found", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			GetUserKeyFunc: func(ctx context.Context, userID int) (*database.UserKey, error) {
				return nil, errors.New("key not found")
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &GetUserKeyRequest{
			UserID: 999,
		}

		resp, err := service.GetUserPublicKey(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to get user key")
	})
}

func TestSendEncryptedMessage(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - Offline Message", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			StoreOfflineEncryptedMessageFunc: func(ctx context.Context, offlineMsg *database.OfflineEncryptedMessage) (int, error) {
				return 1, nil
			},
		}

		mockE2EE := &MockE2EEncryptionService{
			EncryptMessageFunc: func(plaintext []byte, recipientPublicKeyPEM string) ([]byte, []byte, []byte, []byte, error) {
				return []byte("encrypted_session_key"), []byte("encrypted_content"), []byte("iv"), []byte("auth_tag"), nil
			},
		}

		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &SendMessageRequest{
			SenderID:         1,
			RecipientID:      2,
			SessionID:        1,
			PlaintextContent: []byte("Hello, this is a secret message!"),
			RecipientPubKey:  "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----",
		}

		resp, err := service.SendEncryptedMessage(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 1, resp.MessageID)
		assert.Equal(t, []byte("encrypted_session_key"), resp.EncryptedSessionKey)
		assert.Equal(t, []byte("encrypted_content"), resp.EncryptedContent)
		assert.Equal(t, []byte("iv"), resp.Iv)
		assert.Equal(t, []byte("auth_tag"), resp.AuthTag)
		assert.False(t, resp.IsDelivered)
	})

	t.Run("Empty Message Content", func(t *testing.T) {
		mockRepo := &MockChatRepository{}
		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &SendMessageRequest{
			SenderID:         1,
			RecipientID:      2,
			SessionID:        1,
			PlaintextContent: []byte{},
			RecipientPubKey:  "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----",
		}

		resp, err := service.SendEncryptedMessage(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "message content is required")
	})

	t.Run("Missing Recipient Public Key", func(t *testing.T) {
		mockRepo := &MockChatRepository{}
		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &SendMessageRequest{
			SenderID:         1,
			RecipientID:      2,
			SessionID:        1,
			PlaintextContent: []byte("Hello"),
			RecipientPubKey:  "",
		}

		resp, err := service.SendEncryptedMessage(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "recipient public key is required")
	})

	t.Run("Encryption Fails", func(t *testing.T) {
		mockRepo := &MockChatRepository{}
		mockE2EE := &MockE2EEncryptionService{
			EncryptMessageFunc: func(plaintext []byte, recipientPublicKeyPEM string) ([]byte, []byte, []byte, []byte, error) {
				return nil, nil, nil, nil, errors.New("encryption failed")
			},
		}

		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &SendMessageRequest{
			SenderID:         1,
			RecipientID:      2,
			SessionID:        1,
			PlaintextContent: []byte("Hello"),
			RecipientPubKey:  "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----",
		}

		resp, err := service.SendEncryptedMessage(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to encrypt message")
	})

	t.Run("Store Offline Message Fails", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			StoreOfflineEncryptedMessageFunc: func(ctx context.Context, offlineMsg *database.OfflineEncryptedMessage) (int, error) {
				return 0, errors.New("database error")
			},
		}

		mockE2EE := &MockE2EEncryptionService{
			EncryptMessageFunc: func(plaintext []byte, recipientPublicKeyPEM string) ([]byte, []byte, []byte, []byte, error) {
				return []byte("encrypted_session_key"), []byte("encrypted_content"), []byte("iv"), []byte("auth_tag"), nil
			},
		}

		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &SendMessageRequest{
			SenderID:         1,
			RecipientID:      2,
			SessionID:        1,
			PlaintextContent: []byte("Hello"),
			RecipientPubKey:  "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----",
		}

		resp, err := service.SendEncryptedMessage(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to store offline encrypted message")
	})
}

func TestGetOfflineMessages(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		now := time.Now()
		mockRepo := &MockChatRepository{
			GetUndeliveredOfflineMessagesFunc: func(ctx context.Context, recipientID int) ([]*database.OfflineEncryptedMessage, error) {
				return []*database.OfflineEncryptedMessage{
					{
						ID:                  1,
						RecipientID:         recipientID,
						SenderID:            2,
						SessionID:           1,
						EncryptedSessionKey: []byte("key1"),
						EncryptedContent:    []byte("content1"),
						Iv:                  []byte("iv1"),
						AuthTag:             []byte("tag1"),
						CreatedAt:           &now,
						IsDelivered:         false,
					},
					{
						ID:                  2,
						RecipientID:         recipientID,
						SenderID:            3,
						SessionID:           1,
						EncryptedSessionKey: []byte("key2"),
						EncryptedContent:    []byte("content2"),
						Iv:                  []byte("iv2"),
						AuthTag:             []byte("tag2"),
						CreatedAt:           &now,
						IsDelivered:         false,
					},
				}, nil
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &GetOfflineMessagesRequest{
			UserID: 1,
		}

		resp, err := service.GetOfflineMessages(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 2, resp.Count)
		assert.Len(t, resp.Messages, 2)
		assert.Equal(t, 1, resp.Messages[0].MessageID)
		assert.Equal(t, 2, resp.Messages[1].MessageID)
	})

	t.Run("No Messages", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			GetUndeliveredOfflineMessagesFunc: func(ctx context.Context, recipientID int) ([]*database.OfflineEncryptedMessage, error) {
				return []*database.OfflineEncryptedMessage{}, nil
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &GetOfflineMessagesRequest{
			UserID: 1,
		}

		resp, err := service.GetOfflineMessages(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 0, resp.Count)
		assert.Empty(t, resp.Messages)
	})

	t.Run("Get Messages Fails", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			GetUndeliveredOfflineMessagesFunc: func(ctx context.Context, recipientID int) ([]*database.OfflineEncryptedMessage, error) {
				return nil, errors.New("database error")
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		req := &GetOfflineMessagesRequest{
			UserID: 1,
		}

		resp, err := service.GetOfflineMessages(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to get offline messages")
	})
}

func TestMarkMessageDelivered(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			GetUndeliveredOfflineMessagesFunc: func(ctx context.Context, recipientID int) ([]*database.OfflineEncryptedMessage, error) {
				return []*database.OfflineEncryptedMessage{
					{
						ID:          1,
						RecipientID: recipientID,
						SenderID:    2,
						SessionID:   1,
					},
				}, nil
			},
			MarkOfflineMessageDeliveredFunc: func(ctx context.Context, messageID int) error {
				return nil
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		err := service.MarkMessageDelivered(ctx, 1, 1)
		assert.NoError(t, err)
	})

	t.Run("Message Not Found", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			GetUndeliveredOfflineMessagesFunc: func(ctx context.Context, recipientID int) ([]*database.OfflineEncryptedMessage, error) {
				return []*database.OfflineEncryptedMessage{}, nil
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		err := service.MarkMessageDelivered(ctx, 1, 999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message not found or does not belong to user")
	})

	t.Run("Message Belongs To Different User", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			GetUndeliveredOfflineMessagesFunc: func(ctx context.Context, recipientID int) ([]*database.OfflineEncryptedMessage, error) {
				return []*database.OfflineEncryptedMessage{
					{
						ID:          1,
						RecipientID: 999, // Different user
						SenderID:    2,
						SessionID:   1,
					},
				}, nil
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		err := service.MarkMessageDelivered(ctx, 1, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message not found or does not belong to user")
	})

	t.Run("Mark Delivered Fails", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			GetUndeliveredOfflineMessagesFunc: func(ctx context.Context, recipientID int) ([]*database.OfflineEncryptedMessage, error) {
				return []*database.OfflineEncryptedMessage{
					{
						ID:          1,
						RecipientID: recipientID,
						SenderID:    2,
						SessionID:   1,
					},
				}, nil
			},
			MarkOfflineMessageDeliveredFunc: func(ctx context.Context, messageID int) error {
				return errors.New("database error")
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		err := service.MarkMessageDelivered(ctx, 1, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to mark message as delivered")
	})
}

func TestGetUndeliveredMessageCount(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			GetUndeliveredMessageCountFunc: func(ctx context.Context, recipientID int) (int, error) {
				return 5, nil
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		count, err := service.GetUndeliveredMessageCount(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, 5, count)
	})

	t.Run("Get Count Fails", func(t *testing.T) {
		mockRepo := &MockChatRepository{
			GetUndeliveredMessageCountFunc: func(ctx context.Context, recipientID int) (int, error) {
				return 0, errors.New("database error")
			},
		}

		mockE2EE := &MockE2EEncryptionService{}
		service := NewChatService(mockRepo, &MockUserRepository{}, mockE2EE)

		count, err := service.GetUndeliveredMessageCount(ctx, 1)
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to get undelivered message count")
	})
}

func TestChatServiceImplementsInterface(t *testing.T) {
	t.Run("Complete Interface Compliance", func(t *testing.T) {
		// This test verifies that ChatServiceImpl fully implements ChatService
		var service ChatService = &ChatServiceImpl{}
		assert.NotNil(t, service)

		// Verify all methods are accessible through the interface
		assert.NotNil(t, service.CreateSession)
		assert.NotNil(t, service.JoinSession)
		assert.NotNil(t, service.LeaveSession)
		assert.NotNil(t, service.GetUserSessions)
		assert.NotNil(t, service.StoreUserKey)
		assert.NotNil(t, service.GetUserPublicKey)
		assert.NotNil(t, service.SendEncryptedMessage)
		assert.NotNil(t, service.GetOfflineMessages)
		assert.NotNil(t, service.MarkMessageDelivered)
		assert.NotNil(t, service.GetUndeliveredMessageCount)
	})
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}
