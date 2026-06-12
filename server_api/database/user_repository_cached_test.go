package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestCachedUserRepository_CreateUser tests the CreateUser method
func TestCachedUserRepository_CreateUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockRepo, mockCache)

	t.Run("successful user creation", func(t *testing.T) {
		mockRepo.On("CreateUser", "testuser", "$2a$10$hashedpassword").Return(1, nil)
		mockCache.On("Delete", "user:exists:1").Once()
		mockCache.On("Delete", "user:id:username:testuser").Once()

		id, err := cachedRepo.CreateUser("testuser", "$2a$10$hashedpassword")

		assert.NoError(t, err)
		assert.Equal(t, 1, id)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("creation failure", func(t *testing.T) {
		mockRepo.On("CreateUser", "duplicateuser", "$2a$10$hashedpassword").Return(0, errors.New("duplicate key"))

		id, err := cachedRepo.CreateUser("duplicateuser", "$2a$10$hashedpassword")

		assert.Error(t, err)
		assert.Equal(t, 0, id)
		mockRepo.AssertExpectations(t)
	})
}

// TestCachedUserRepository_FindByID tests the FindByID method
func TestCachedUserRepository_FindByID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockRepo, mockCache)

	t.Run("user found in cache", func(t *testing.T) {
		mockCache.On("Get", "user:exists:1").Return("true", true)
		mockRepo.On("FindByID", 1).Return(&User{
			ID:       1,
			Username: "cacheduser",
		}, nil)
		mockCache.On("SetWithTTL", "user:exists:1", "true", 300).Once()

		user, err := cachedRepo.FindByID(1)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, 1, user.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found - cached negative result", func(t *testing.T) {
		mockCache.On("Get", "user:exists:999").Return("false", true)

		user, err := cachedRepo.FindByID(999)

		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("user found in database - cache miss", func(t *testing.T) {
		mockCache.On("Get", "user:exists:2").Return("", false)
		mockRepo.On("FindByID", 2).Return(&User{
			ID:       2,
			Username: "dbuser",
		}, nil)
		mockCache.On("SetWithTTL", "user:exists:2", "true", 300).Once()

		user, err := cachedRepo.FindByID(2)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, 2, user.ID)
	})

	t.Run("database error", func(t *testing.T) {
		mockCache.On("Get", "user:exists:3").Return("", false)
		mockRepo.On("FindByID", 3).Return(nil, errors.New("database error"))
		mockCache.On("SetWithTTL", "user:exists:3", "false", 300).Once()

		user, err := cachedRepo.FindByID(3)

		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

// TestCachedUserRepository_FindByUsername tests the FindByUsername method
func TestCachedUserRepository_FindByUsername(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockRepo, mockCache)

	t.Run("user found - cache hit", func(t *testing.T) {
		mockCache.On("Get", "user:id:username:testuser").Return("1", true)
		mockRepo.On("FindByUsername", "testuser").Return(&User{
			ID:           1,
			Username:     "testuser",
			PasswordHash: "$2a$10$hash",
		}, nil)
		mockCache.On("SetWithTTL", "user:id:username:testuser", "1", 300).Once()

		user, err := cachedRepo.FindByUsername("testuser")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, 1, user.ID)
	})

	t.Run("user not found - cached", func(t *testing.T) {
		mockCache.On("Get", "user:id:username:nonexistent").Return("", false)
		mockRepo.On("FindByUsername", "nonexistent").Return(nil, errors.New("user not found"))
		mockCache.On("SetNull", "user:id:username:nonexistent").Once()
		mockCache.On("SetExpiration", "user:id:username:nonexistent", time.Duration(60)*time.Second).Once()

		user, err := cachedRepo.FindByUsername("nonexistent")

		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

// TestCachedUserRepository_ExistsByID tests the ExistsByID method
func TestCachedUserRepository_ExistsByID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockRepo, mockCache)

	t.Run("user exists - cached positive", func(t *testing.T) {
		mockCache.On("Get", "user:exists:1").Return("true", true)

		exists, err := cachedRepo.ExistsByID(1)

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("user does not exist - cached negative", func(t *testing.T) {
		mockCache.On("Get", "user:exists:2").Return("false", true)

		exists, err := cachedRepo.ExistsByID(2)

		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("cache miss - user exists in DB", func(t *testing.T) {
		mockCache.On("Get", "user:exists:3").Return("", false)
		mockRepo.On("ExistsByID", 3).Return(true, nil)
		mockCache.On("SetWithTTL", "user:exists:3", "true", 300).Once()

		exists, err := cachedRepo.ExistsByID(3)

		assert.NoError(t, err)
		assert.True(t, exists)
	})
}

// TestCachedUserRepository_GetUserProfile tests the GetUserProfile method
func TestCachedUserRepository_GetUserProfile(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockRepo, mockCache)

	t.Run("profile found in cache", func(t *testing.T) {
		mockCache.On("Get", "user:info:1").Return(`{"Username":"testuser","Email":"test@example.com","Nickname":"Test User","Bio":"Test bio"}`, true)

		profile, err := cachedRepo.GetUserProfile(1)

		assert.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, "testuser", profile.Username)
	})

	t.Run("profile not found in cache - fetch from DB", func(t *testing.T) {
		mockCache.On("Get", "user:info:2").Return("", false)
		mockRepo.On("GetUserProfile", 2).Return(nil, errors.New("profile not found"))
		mockCache.On("SetNull", "user:info:2").Once()
		mockCache.On("SetExpiration", "user:info:2", time.Duration(60)*time.Second).Once()

		profile, err := cachedRepo.GetUserProfile(2)

		assert.Error(t, err)
		assert.Nil(t, profile)
	})
}

// TestCachedUserRepository_UpdateProfile tests the UpdateProfile method
func TestCachedUserRepository_UpdateProfile(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockRepo, mockCache)

	t.Run("successful profile update", func(t *testing.T) {
		mockRepo.On("UpdateProfile", 1, "email", "updated@example.com").Return(nil)
		mockCache.On("Delete", "user:info:1").Once()
		mockCache.On("Delete", "user:basic:1").Once()

		err := cachedRepo.UpdateProfile(1, "email", "updated@example.com")

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("update fails", func(t *testing.T) {
		mockRepo.On("UpdateProfile", 2, "nickname", "Invalid Name").Return(errors.New("update failed"))

		err := cachedRepo.UpdateProfile(2, "nickname", "Invalid Name")

		assert.Error(t, err)
	})
}

// TestCachedUserRepository_UpdatePassword tests the UpdatePassword method
func TestCachedUserRepository_UpdatePassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockRepo, mockCache)

	t.Run("successful password update", func(t *testing.T) {
		mockRepo.On("UpdatePassword", 1, "$2a$10$newhash").Return(nil)
		mockCache.On("Delete", "user:info:1").Once()
		mockCache.On("Delete", "user:basic:1").Once()

		err := cachedRepo.UpdatePassword(1, "$2a$10$newhash")

		assert.NoError(t, err)
	})

	t.Run("password update fails", func(t *testing.T) {
		mockRepo.On("UpdatePassword", 2, "$2a$10$badhash").Return(errors.New("password update failed"))

		err := cachedRepo.UpdatePassword(2, "$2a$10$badhash")

		assert.Error(t, err)
	})
}

// TestCachedUserRepository_UpdateLastLogin tests the UpdateLastLogin method
func TestCachedUserRepository_UpdateLastLogin(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockRepo, mockCache)

	t.Run("successful last login update", func(t *testing.T) {
		mockRepo.On("UpdateLastLogin", 1).Return(nil)

		err := cachedRepo.UpdateLastLogin(1)

		assert.NoError(t, err)
	})

	t.Run("last login update fails", func(t *testing.T) {
		mockRepo.On("UpdateLastLogin", 2).Return(errors.New("update failed"))

		err := cachedRepo.UpdateLastLogin(2)

		assert.Error(t, err)
	})
}

// TestCachedUserRepository_UpdateLockStatus tests the UpdateLockStatus method
func TestCachedUserRepository_UpdateLockStatus(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockRepo, mockCache)

	t.Run("lock user account", func(t *testing.T) {
		mockRepo.On("UpdateLockStatus", "user1", mock.AnythingOfType("*time.Time")).Return(nil)
		mockCache.On("Delete", "login_lockout:user1").Once()

		err := cachedRepo.UpdateLockStatus("user1", &time.Time{})

		assert.NoError(t, err)
	})

	t.Run("unlock user account", func(t *testing.T) {
		mockRepo.On("UpdateLockStatus", "user2", (*time.Time)(nil)).Return(nil)
		mockCache.On("Delete", "login_lockout:user2").Once()

		err := cachedRepo.UpdateLockStatus("user2", nil)

		assert.NoError(t, err)
	})
}

// TestCachedUserRepository_GetLockStatus tests the GetLockStatus method
func TestCachedUserRepository_GetLockStatus(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockRepo, mockCache)

	t.Run("user lock status exists", func(t *testing.T) {
		mockRepo.On("GetLockStatus", "user1").Return(&time.Time{}, nil)

		timeResult, err := cachedRepo.GetLockStatus("user1")

		assert.NoError(t, err)
		assert.NotNil(t, timeResult)
	})

	t.Run("user has no lock", func(t *testing.T) {
		mockRepo.On("GetLockStatus", "user2").Return((*time.Time)(nil), nil)

		timeResult, err := cachedRepo.GetLockStatus("user2")

		assert.NoError(t, err)
		assert.Nil(t, timeResult)
	})
}

// TestCachedUserRepository_ClearExpiredLock tests the ClearExpiredLock method
func TestCachedUserRepository_ClearExpiredLock(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockRepo, mockCache)

	t.Run("successful expired lock clear", func(t *testing.T) {
		mockRepo.On("ClearExpiredLock", "user1").Return(nil)
		mockCache.On("Delete", "login_lockout:user1").Once()
		mockCache.On("Delete", "login_attempts:user1").Once()

		err := cachedRepo.ClearExpiredLock("user1")

		assert.NoError(t, err)
	})

	t.Run("failed to clear expired lock", func(t *testing.T) {
		mockRepo.On("ClearExpiredLock", "user2").Return(errors.New("clear failed"))

		err := cachedRepo.ClearExpiredLock("user2")

		assert.Error(t, err)
	})
}

// TestCachedUserRepository_DeleteUser tests the DeleteUser method
func TestCachedUserRepository_DeleteUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockRepo, mockCache)

	t.Run("successful user deletion", func(t *testing.T) {
		mockRepo.On("FindByID", 1).Return(&User{ID: 1, Username: "testuser"}, nil)
		mockRepo.On("DeleteUser", 1).Return(nil)
		mockCache.On("Delete", "user:exists:1").Once()
		mockCache.On("Delete", "user:id:username:testuser").Once()
		mockCache.On("Delete", "user:info:1").Once()
		mockCache.On("Delete", "user:basic:1").Once()

		err := cachedRepo.DeleteUser(1)

		assert.NoError(t, err)
	})

	t.Run("deletion fails", func(t *testing.T) {
		mockRepo.On("FindByID", 2).Return(&User{ID: 2, Username: "testuser2"}, nil)
		mockRepo.On("DeleteUser", 2).Return(errors.New("delete failed"))
		mockCache.On("Delete", "user:exists:2").Once()
		mockCache.On("Delete", "user:id:username:testuser2").Once()
		mockCache.On("Delete", "user:info:2").Once()
		mockCache.On("Delete", "user:basic:2").Once()

		err := cachedRepo.DeleteUser(2)

		assert.Error(t, err)
	})
}

// TestCachedUserRepository_GetUserBasicInfo tests the GetUserBasicInfo method
func TestCachedUserRepository_GetUserBasicInfo(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockRepo, mockCache)

	t.Run("basic info found in cache", func(t *testing.T) {
		mockCache.On("Get", "user:basic:1").Return(`{"ID":1,"Username":"testuser","Nickname":"Test User"}`, true)

		basicInfo, err := cachedRepo.GetUserBasicInfo(1)

		assert.NoError(t, err)
		assert.NotNil(t, basicInfo)
		assert.Equal(t, "testuser", basicInfo.Username)
	})

	t.Run("basic info not found in cache - fetch from DB", func(t *testing.T) {
		mockCache.On("Get", "user:basic:2").Return("", false)
		mockRepo.On("GetUserBasicInfo", 2).Return(nil, errors.New("not found"))
		mockCache.On("SetNull", "user:basic:2").Once()
		mockCache.On("SetExpiration", "user:basic:2", time.Duration(60)*time.Second).Once()

		basicInfo, err := cachedRepo.GetUserBasicInfo(2)

		assert.Error(t, err)
		assert.Nil(t, basicInfo)
	})
}

// TestCachedUserRepository_InterfaceCompliance tests that the cached repository implements the interface
func TestCachedUserRepository_InterfaceCompliance(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockRepo, mockCache)

	// Verify that the cached repository implements UserRepository
	var _ UserRepository = cachedRepo
	assert.NotNil(t, cachedRepo)
}

// TestChatRepositoryAdapter_InterfaceCompliance tests that the chat repository adapter implements the interface
func TestChatRepositoryAdapter_InterfaceCompliance(t *testing.T) {
	mockReader := &MockChatReader{}
	mockWriter := &MockChatWriter{}
	adapter := NewChatRepositoryAdapter(mockReader, mockWriter)

	// Verify that the adapter implements ChatRepository
	var _ ChatRepository = adapter
	assert.NotNil(t, adapter)
}

// TestChatRepositoryAdapter_CreateChatSession tests the CreateChatSession method
func TestChatRepositoryAdapter_CreateChatSession(t *testing.T) {
	mockReader := &MockChatReader{}
	mockWriter := &MockChatWriter{}
	adapter := NewChatRepositoryAdapter(mockReader, mockWriter)

	ctx := context.Background()
	sessionType := "direct"
	name := "Test Session"
	createdBy := 1

	mockWriter.On("CreateChatSession", ctx, sessionType, &name, &createdBy).Return(1, nil)

	id, err := adapter.CreateChatSession(ctx, sessionType, &name, &createdBy)

	assert.NoError(t, err)
	assert.Equal(t, 1, id)
	mockWriter.AssertExpectations(t)
}

// TestChatRepositoryAdapter_GetChatSession tests the GetChatSession method
func TestChatRepositoryAdapter_GetChatSession(t *testing.T) {
	mockReader := &MockChatReader{}
	mockWriter := &MockChatWriter{}
	adapter := NewChatRepositoryAdapter(mockReader, mockWriter)

	ctx := context.Background()
	sessionID := 1

	expectedSession := &ChatSession{
		ID:          sessionID,
		SessionType: "direct",
	}
	mockReader.On("GetChatSession", ctx, sessionID).Return(expectedSession, nil)

	session, err := adapter.GetChatSession(ctx, sessionID)

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, sessionID, session.ID)
	mockReader.AssertExpectations(t)
}

// TestChatRepositoryAdapter_StoreEncryptedMessage tests the StoreEncryptedMessage method
func TestChatRepositoryAdapter_StoreEncryptedMessage(t *testing.T) {
	mockReader := &MockChatReader{}
	mockWriter := &MockChatWriter{}
	adapter := NewChatRepositoryAdapter(mockReader, mockWriter)

	ctx := context.Background()
	encryptedMsg := &EncryptedMessage{
		MessageID:        1,
		EncryptedContent: "encrypted_content",
	}

	mockWriter.On("StoreEncryptedMessage", ctx, encryptedMsg).Return(nil)

	err := adapter.StoreEncryptedMessage(ctx, encryptedMsg)

	assert.NoError(t, err)
	mockWriter.AssertExpectations(t)
}

// TestChatRepositoryAdapter_GetUserPublicKey tests the GetUserPublicKey method
func TestChatRepositoryAdapter_GetUserPublicKey(t *testing.T) {
	mockReader := &MockChatReader{}
	mockWriter := &MockChatWriter{}
	adapter := NewChatRepositoryAdapter(mockReader, mockWriter)

	ctx := context.Background()
	userID := 1
	expectedKey := "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----"

	mockReader.On("GetUserPublicKey", ctx, userID).Return(expectedKey, nil)

	key, err := adapter.GetUserPublicKey(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedKey, key)
	mockReader.AssertExpectations(t)
}

// TestChatRepositoryAdapter_StoreOfflineEncryptedMessage tests the StoreOfflineEncryptedMessage method
func TestChatRepositoryAdapter_StoreOfflineEncryptedMessage(t *testing.T) {
	mockReader := &MockChatReader{}
	mockWriter := &MockChatWriter{}
	adapter := NewChatRepositoryAdapter(mockReader, mockWriter)

	ctx := context.Background()
	offlineMsg := &OfflineEncryptedMessage{
		RecipientID:      1,
		SenderID:         2,
		SessionID:        1,
		EncryptedContent: []byte("encrypted"),
	}

	mockWriter.On("StoreOfflineEncryptedMessage", ctx, offlineMsg).Return(1, nil)

	id, err := adapter.StoreOfflineEncryptedMessage(ctx, offlineMsg)

	assert.NoError(t, err)
	assert.Equal(t, 1, id)
	mockWriter.AssertExpectations(t)
}

// TestChatRepositoryAdapter_GetUndeliveredOfflineMessages tests the GetUndeliveredOfflineMessages method
func TestChatRepositoryAdapter_GetUndeliveredOfflineMessages(t *testing.T) {
	mockReader := &MockChatReader{}
	mockWriter := &MockChatWriter{}
	adapter := NewChatRepositoryAdapter(mockReader, mockWriter)

	ctx := context.Background()
	recipientID := 1

	expectedMessages := []*OfflineEncryptedMessage{
		{
			ID:          1,
			RecipientID: recipientID,
			SenderID:    2,
		},
	}
	mockReader.On("GetUndeliveredOfflineMessages", ctx, recipientID).Return(expectedMessages, nil)

	messages, err := adapter.GetUndeliveredOfflineMessages(ctx, recipientID)

	assert.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, 1, messages[0].ID)
	mockReader.AssertExpectations(t)
}
