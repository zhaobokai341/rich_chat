package database

import (
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
	cachedRepo := NewCachedUserRepository(mockRepo, mockCache)

	tests := []struct {
		name          string
		username      string
		passwordHash  string
		expectedID    int
		expectedError error
		setupMocks    func()
	}{
		{
			name:          "successful user creation",
			username:      "testuser",
			passwordHash:  "$2a$10$hashedpassword",
			expectedID:    1,
			expectedError: nil,
			setupMocks: func() {
				mockRepo.On("CreateUser", "testuser", "$2a$10$hashedpassword").Return(1, nil)
				mockCache.On("Delete", "user:exists:1").Once()
				mockCache.On("Delete", "user:id:username:testuser").Once()
				mockCache.On("Set", "user:hash:1", "$2a$10$hashedpassword").Once()
			},
		},
		{
			name:          "creation failure",
			username:      "duplicateuser",
			passwordHash:  "$2a$10$hashedpassword",
			expectedID:    0,
			expectedError: errors.New("duplicate key"),
			setupMocks: func() {
				mockRepo.On("CreateUser", "duplicateuser", "$2a$10$hashedpassword").Return(0, errors.New("duplicate key"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			id, err := cachedRepo.CreateUser(tt.username, tt.passwordHash)

			assert.Equal(t, tt.expectedID, id)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestCachedUserRepository_FindByID tests the FindByID method
func TestCachedUserRepository_FindByID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockCache)

	tests := []struct {
		name          string
		userID        int
		expectedUser  *User
		expectedError error
		setupMocks    func()
	}{
		{
			name:   "user found in cache",
			userID: 1,
			expectedUser: &User{
				ID:       1,
				Username: "cacheduser",
			},
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Get", "user:exists:1").Return("true", true)
				mockRepo.On("FindByID", 1).Return(&User{
					ID:       1,
					Username: "cacheduser",
				}, nil)
				mockCache.On("Set", "user:exists:1", "true").Once()
			},
		},
		{
			name:          "user not found - cached negative result",
			userID:        999,
			expectedUser:  nil,
			expectedError: errors.New("user not found"),
			setupMocks: func() {
				mockCache.On("Get", "user:exists:999").Return("false", true)
			},
		},
		{
			name:   "user found in database - cache miss",
			userID: 2,
			expectedUser: &User{
				ID:       2,
				Username: "dbuser",
			},
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Get", "user:exists:2").Return("", false)
				mockRepo.On("FindByID", 2).Return(&User{
					ID:       2,
					Username: "dbuser",
				}, nil)
				mockCache.On("Set", "user:exists:2", "true").Once()
			},
		},
		{
			name:          "database error",
			userID:        3,
			expectedUser:  nil,
			expectedError: errors.New("database error"),
			setupMocks: func() {
				mockCache.On("Get", "user:exists:3").Return("", false)
				mockRepo.On("FindByID", 3).Return(nil, errors.New("database error"))
				mockCache.On("Set", "user:exists:3", "false").Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			user, err := cachedRepo.FindByID(tt.userID)

			if tt.expectedUser != nil {
				assert.NotNil(t, user)
				assert.Equal(t, tt.expectedUser.ID, user.ID)
				assert.Equal(t, tt.expectedUser.Username, user.Username)
			} else {
				assert.Nil(t, user)
			}

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestCachedUserRepository_FindByUsername tests the FindByUsername method
func TestCachedUserRepository_FindByUsername(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockCache)

	tests := []struct {
		name          string
		username      string
		expectedUser  *User
		expectedError error
		setupMocks    func()
	}{
		{
			name:     "user found - cache hit",
			username: "testuser",
			expectedUser: &User{
				ID:           1,
				Username:     "testuser",
				PasswordHash: "$2a$10$hash",
			},
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Get", "user:id:username:testuser").Return("1", true)
				mockRepo.On("FindByUsername", "testuser").Return(&User{
					ID:           1,
					Username:     "testuser",
					PasswordHash: "$2a$10$hash",
				}, nil)
				mockCache.On("Set", "user:id:username:testuser", "1").Once()
				mockCache.On("Set", "user:hash:1", "$2a$10$hash").Once()
			},
		},
		{
			name:          "user not found - cached",
			username:      "nonexistent",
			expectedUser:  nil,
			expectedError: errors.New("user not found"),
			setupMocks: func() {
				mockCache.On("Get", "user:id:username:nonexistent").Return("", false)
				mockRepo.On("FindByUsername", "nonexistent").Return(nil, errors.New("user not found"))
				mockCache.On("SetNull", "user:id:username:nonexistent").Once()
			},
		},
		{
			name:     "user found - cache miss",
			username: "anotheruser",
			expectedUser: &User{
				ID:           2,
				Username:     "anotheruser",
				PasswordHash: "$2a$10$anotherhash",
			},
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Get", "user:id:username:anotheruser").Return("", false)
				mockRepo.On("FindByUsername", "anotheruser").Return(&User{
					ID:           2,
					Username:     "anotheruser",
					PasswordHash: "$2a$10$anotherhash",
				}, nil)
				mockCache.On("Set", "user:id:username:anotheruser", "2").Once()
				mockCache.On("Set", "user:hash:2", "$2a$10$anotherhash").Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			user, err := cachedRepo.FindByUsername(tt.username)

			if tt.expectedUser != nil {
				assert.NotNil(t, user)
				assert.Equal(t, tt.expectedUser.ID, user.ID)
				assert.Equal(t, tt.expectedUser.Username, user.Username)
				assert.Equal(t, tt.expectedUser.PasswordHash, user.PasswordHash)
			} else {
				assert.Nil(t, user)
			}

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestCachedUserRepository_ExistsByID tests the ExistsByID method
func TestCachedUserRepository_ExistsByID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockCache)

	tests := []struct {
		name         string
		userID       int
		expectedBool bool
		expectedErr  error
		setupMocks   func()
	}{
		{
			name:         "user exists - cached positive",
			userID:       1,
			expectedBool: true,
			expectedErr:  nil,
			setupMocks: func() {
				mockCache.On("Get", "user:exists:1").Return("true", true)
			},
		},
		{
			name:         "user does not exist - cached negative",
			userID:       2,
			expectedBool: false,
			expectedErr:  nil,
			setupMocks: func() {
				mockCache.On("Get", "user:exists:2").Return("false", true)
			},
		},
		{
			name:         "cache miss - user exists in DB",
			userID:       3,
			expectedBool: true,
			expectedErr:  nil,
			setupMocks: func() {
				mockCache.On("Get", "user:exists:3").Return("", false)
				mockRepo.On("ExistsByID", 3).Return(true, nil)
				mockCache.On("Set", "user:exists:3", "true").Once()
			},
		},
		{
			name:         "cache miss - user does not exist in DB",
			userID:       4,
			expectedBool: false,
			expectedErr:  nil,
			setupMocks: func() {
				mockCache.On("Get", "user:exists:4").Return("", false)
				mockRepo.On("ExistsByID", 4).Return(false, nil)
				mockCache.On("Set", "user:exists:4", "false").Once()
			},
		},
		{
			name:         "error case",
			userID:       5,
			expectedBool: false,
			expectedErr:  errors.New("database error"),
			setupMocks: func() {
				mockCache.On("Get", "user:exists:5").Return("", false)
				mockRepo.On("ExistsByID", 5).Return(false, errors.New("database error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			exists, err := cachedRepo.ExistsByID(tt.userID)

			assert.Equal(t, tt.expectedBool, exists)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestCachedUserRepository_GetUserProfile tests the GetUserProfile method
func TestCachedUserRepository_GetUserProfile(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockCache)

	tests := []struct {
		name         string
		userID       int
		expectedUser *UserInfo
		expectedErr  error
		setupMocks   func()
	}{
		{
			name:   "profile found in cache",
			userID: 1,
			expectedUser: &UserInfo{
				Username: "testuser",
				Email:    "test@example.com",
				Nickname: "Test User",
				Bio:      "Test bio",
			},
			expectedErr: nil,
			setupMocks: func() {
				mockCache.On("Get", "user:info:1").Return(`{"Username":"testuser","Email":"test@example.com","Nickname":"Test User","Bio":"Test bio"}`, true)
			},
		},
		{
			name:         "profile not found in cache - fetch from DB",
			userID:       2,
			expectedUser: nil,
			expectedErr:  errors.New("profile not found"),
			setupMocks: func() {
				mockCache.On("Get", "user:info:2").Return("", false)
				mockRepo.On("GetUserProfile", 2).Return(nil, errors.New("profile not found"))
				mockCache.On("SetNull", "user:info:2").Once()
			},
		},
		{
			name:   "profile found in DB - cache miss",
			userID: 3,
			expectedUser: &UserInfo{
				Username: "anotheruser",
				Email:    "another@example.com",
				Nickname: "Another User",
				Bio:      "Another bio",
			},
			expectedErr: nil,
			setupMocks: func() {
				mockCache.On("Get", "user:info:3").Return("", false)
				mockRepo.On("GetUserProfile", 3).Return(&UserInfo{
					Username: "anotheruser",
					Email:    "another@example.com",
					Nickname: "Another User",
					Bio:      "Another bio",
				}, nil)
				mockCache.On("Set", "user:info:3", mock.AnythingOfType("string")).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			profile, err := cachedRepo.GetUserProfile(tt.userID)

			if tt.expectedUser != nil {
				assert.NotNil(t, profile)
				assert.Equal(t, tt.expectedUser.Username, profile.Username)
				assert.Equal(t, tt.expectedUser.Email, profile.Email)
				assert.Equal(t, tt.expectedUser.Nickname, profile.Nickname)
				assert.Equal(t, tt.expectedUser.Bio, profile.Bio)
			} else {
				assert.Nil(t, profile)
			}

			if tt.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestCachedUserRepository_UpdateProfile tests the UpdateProfile method
func TestCachedUserRepository_UpdateProfile(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockCache)

	tests := []struct {
		name        string
		userID      int
		key         string
		value       string
		expectedErr error
		setupMocks  func()
	}{
		{
			name:        "successful profile update",
			userID:      1,
			key:         "email",
			value:       "updated@example.com",
			expectedErr: nil,
			setupMocks: func() {
				mockRepo.On("UpdateProfile", 1, "email", "updated@example.com").Return(nil)
				mockCache.On("Delete", "user:info:1").Once()
			},
		},
		{
			name:        "update fails",
			userID:      2,
			key:         "nickname",
			value:       "Invalid Name",
			expectedErr: errors.New("update failed"),
			setupMocks: func() {
				mockRepo.On("UpdateProfile", 2, "nickname", "Invalid Name").Return(errors.New("update failed"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := cachedRepo.UpdateProfile(tt.userID, tt.key, tt.value)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestCachedUserRepository_UpdatePassword tests the UpdatePassword method
func TestCachedUserRepository_UpdatePassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockCache)

	tests := []struct {
		name        string
		userID      int
		newPassword string
		expectedErr error
		setupMocks  func()
	}{
		{
			name:        "successful password update",
			userID:      1,
			newPassword: "$2a$10$newhash",
			expectedErr: nil,
			setupMocks: func() {
				mockRepo.On("UpdatePassword", 1, "$2a$10$newhash").Return(nil)
				mockCache.On("Delete", "user:hash:1").Once()
			},
		},
		{
			name:        "password update fails",
			userID:      2,
			newPassword: "$2a$10$badhash",
			expectedErr: errors.New("password update failed"),
			setupMocks: func() {
				mockRepo.On("UpdatePassword", 2, "$2a$10$badhash").Return(errors.New("password update failed"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := cachedRepo.UpdatePassword(tt.userID, tt.newPassword)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestCachedUserRepository_UpdateLastLogin tests the UpdateLastLogin method
func TestCachedUserRepository_UpdateLastLogin(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockCache)

	tests := []struct {
		name        string
		userID      int
		expectedErr error
		setupMocks  func()
	}{
		{
			name:        "successful last login update",
			userID:      1,
			expectedErr: nil,
			setupMocks: func() {
				mockRepo.On("UpdateLastLogin", 1).Return(nil)
			},
		},
		{
			name:        "last login update fails",
			userID:      2,
			expectedErr: errors.New("update failed"),
			setupMocks: func() {
				mockRepo.On("UpdateLastLogin", 2).Return(errors.New("update failed"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := cachedRepo.UpdateLastLogin(tt.userID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestCachedUserRepository_UpdateLockStatus tests the UpdateLockStatus method
func TestCachedUserRepository_UpdateLockStatus(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockCache)

	tests := []struct {
		name        string
		identifier  string
		lockUntil   *time.Time
		expectedErr error
		setupMocks  func()
	}{
		{
			name:        "lock user account",
			identifier:  "user1",
			lockUntil:   &time.Time{},
			expectedErr: nil,
			setupMocks: func() {
				mockRepo.On("UpdateLockStatus", "user1", mock.AnythingOfType("*time.Time")).Return(nil)
			},
		},
		{
			name:        "unlock user account",
			identifier:  "user2",
			lockUntil:   nil,
			expectedErr: nil,
			setupMocks: func() {
				mockRepo.On("UpdateLockStatus", "user2", (*time.Time)(nil)).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := cachedRepo.UpdateLockStatus(tt.identifier, tt.lockUntil)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestCachedUserRepository_GetLockStatus tests the GetLockStatus method
func TestCachedUserRepository_GetLockStatus(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockCache)

	tests := []struct {
		name         string
		identifier   string
		expectedTime *time.Time
		expectedErr  error
		setupMocks   func()
	}{
		{
			name:         "user lock status exists",
			identifier:   "user1",
			expectedTime: &time.Time{},
			expectedErr:  nil,
			setupMocks: func() {
				mockRepo.On("GetLockStatus", "user1").Return(&time.Time{}, nil)
			},
		},
		{
			name:         "user has no lock",
			identifier:   "user2",
			expectedTime: nil,
			expectedErr:  nil,
			setupMocks: func() {
				mockRepo.On("GetLockStatus", "user2").Return((*time.Time)(nil), nil)
			},
		},
		{
			name:         "lock status found in DB",
			identifier:   "user3",
			expectedTime: &time.Time{},
			expectedErr:  nil,
			setupMocks: func() {
				mockRepo.On("GetLockStatus", "user3").Return(&time.Time{}, nil)
			},
		},
		{
			name:         "DB error",
			identifier:   "user4",
			expectedTime: nil,
			expectedErr:  errors.New("DB error"),
			setupMocks: func() {
				mockRepo.On("GetLockStatus", "user4").Return((*time.Time)(nil), errors.New("DB error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			timeResult, err := cachedRepo.GetLockStatus(tt.identifier)

			if tt.expectedTime != nil && timeResult != nil {
				assert.Equal(t, tt.expectedTime.String(), timeResult.String()) // Compare string representations
			} else {
				assert.Equal(t, tt.expectedTime, timeResult)
			}

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestCachedUserRepository_ClearExpiredLock tests the ClearExpiredLock method
func TestCachedUserRepository_ClearExpiredLock(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockCache)

	tests := []struct {
		name        string
		identifier  string
		expectedErr error
		setupMocks  func()
	}{
		{
			name:        "successful expired lock clear",
			identifier:  "user1",
			expectedErr: nil,
			setupMocks: func() {
				mockRepo.On("ClearExpiredLock", "user1").Return(nil)
			},
		},
		{
			name:        "failed to clear expired lock",
			identifier:  "user2",
			expectedErr: errors.New("clear failed"),
			setupMocks: func() {
				mockRepo.On("ClearExpiredLock", "user2").Return(errors.New("clear failed"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := cachedRepo.ClearExpiredLock(tt.identifier)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestCachedUserRepository_DeleteUser tests the DeleteUser method
func TestCachedUserRepository_DeleteUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockCache)

	tests := []struct {
		name        string
		userID      int
		expectedErr error
		setupMocks  func()
	}{
		{
			name:        "successful user deletion",
			userID:      1,
			expectedErr: nil,
			setupMocks: func() {
				mockRepo.On("FindByID", 1).Return(&User{ID: 1, Username: "testuser"}, nil)
				mockRepo.On("DeleteUser", 1).Return(nil)
				mockCache.On("Delete", "user:exists:1").Once()
				mockCache.On("Delete", "user:id:username:testuser").Once()
				mockCache.On("Delete", "user:hash:1").Once()
				mockCache.On("Delete", "user:info:1").Once()
			},
		},
		{
			name:        "delete user fails",
			userID:      2,
			expectedErr: errors.New("delete failed"),
			setupMocks: func() {
				mockRepo.On("FindByID", 2).Return(&User{ID: 2, Username: "user2"}, nil)
				mockCache.On("Delete", "user:exists:2").Once()
				mockCache.On("Delete", "user:id:username:user2").Once()
				mockCache.On("Delete", "user:hash:2").Once()
				mockCache.On("Delete", "user:info:2").Once()
				mockRepo.On("DeleteUser", 2).Return(errors.New("delete failed"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := cachedRepo.DeleteUser(tt.userID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}
