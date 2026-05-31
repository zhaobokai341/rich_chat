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
			name:         "successful user creation",
			username:     "testuser",
			passwordHash: "$2a$10$hashedpassword",
			expectedID:   1,
			expectedError: nil,
			setupMocks: func() {
				mockRepo.On("CreateUser", "testuser", "$2a$10$hashedpassword").Return(1, nil)
				mockCache.On("Delete", "user:exists:1").Once()
				mockCache.On("Delete", "user:id:username:testuser").Once()
				mockCache.On("Set", "user:hash:1", "$2a$10$hashedpassword").Once()
			},
		},
		{
			name:         "creation failure",
			username:     "duplicateuser",
			passwordHash: "$2a$10$hashedpassword",
			expectedID:   0,
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
				mockCache.On("Get", "user:id:username:nonexistent").Return("", true)
			},
		},
		{
			name:     "user found - cache miss",
			username: "newuser",
			expectedUser: &User{
				ID:           5,
				Username:     "newuser",
				PasswordHash: "$2a$10$newhash",
			},
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Get", "user:id:username:newuser").Return("", false)
				mockRepo.On("FindByUsername", "newuser").Return(&User{
					ID:           5,
					Username:     "newuser",
					PasswordHash: "$2a$10$newhash",
				}, nil)
				mockCache.On("Set", "user:id:username:newuser", "5").Once()
				mockCache.On("Set", "user:hash:5", "$2a$10$newhash").Once()
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
		name          string
		userID        int
		expectedExist bool
		expectedError error
		setupMocks    func()
	}{
		{
			name:          "user exists - cached",
			userID:        1,
			expectedExist: true,
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Get", "user:exists:1").Return("true", true)
			},
		},
		{
			name:          "user does not exist - cached",
			userID:        999,
			expectedExist: false,
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Get", "user:exists:999").Return("false", true)
			},
		},
		{
			name:          "user exists - cache miss",
			userID:        2,
			expectedExist: true,
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Get", "user:exists:2").Return("", false)
				mockRepo.On("ExistsByID", 2).Return(true, nil)
				mockCache.On("Set", "user:exists:2", "true").Once()
			},
		},
		{
			name:          "database error",
			userID:        3,
			expectedExist: false,
			expectedError: errors.New("database error"),
			setupMocks: func() {
				mockCache.On("Get", "user:exists:3").Return("", false)
				mockRepo.On("ExistsByID", 3).Return(false, errors.New("database error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			
			exists, err := cachedRepo.ExistsByID(tt.userID)
			
			assert.Equal(t, tt.expectedExist, exists)
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

// TestCachedUserRepository_GetUserProfile tests the GetUserProfile method
func TestCachedUserRepository_GetUserProfile(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockCache)

	tests := []struct {
		name          string
		userID        int
		expectedInfo  *UserInfo
		expectedError error
		setupMocks    func()
	}{
		{
			name:   "profile found in cache",
			userID: 1,
			expectedInfo: &UserInfo{
				Username: "cacheduser",
				Nickname: "Cached User",
				Bio:      "Test bio",
			},
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Get", "user:info:1").Return(`{"username":"cacheduser","nickname":"Cached User","bio":"Test bio"}`, true)
			},
		},
		{
			name:          "profile not found - cached",
			userID:        999,
			expectedInfo:  nil,
			expectedError: errors.New("user not found"),
			setupMocks: func() {
				mockCache.On("Get", "user:info:999").Return("", true)
			},
		},
		{
			name:   "profile found in database",
			userID: 2,
			expectedInfo: &UserInfo{
				Username: "dbuser",
				Nickname: "DB User",
				Bio:      "Database bio",
			},
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Get", "user:info:2").Return("", false)
				mockRepo.On("GetUserProfile", 2).Return(&UserInfo{
					Username: "dbuser",
					Nickname: "DB User",
					Bio:      "Database bio",
				}, nil)
				mockCache.On("Set", "user:info:2", mock.MatchedBy(func(s string) bool {
					return len(s) > 0
				})).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			
			info, err := cachedRepo.GetUserProfile(tt.userID)
			
			if tt.expectedInfo != nil {
				assert.NotNil(t, info)
				assert.Equal(t, tt.expectedInfo.Username, info.Username)
				assert.Equal(t, tt.expectedInfo.Nickname, info.Nickname)
			} else {
				assert.Nil(t, info)
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

// TestCachedUserRepository_DeleteUser tests the DeleteUser method
func TestCachedUserRepository_DeleteUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockCache)

	tests := []struct {
		name          string
		userID        int
		expectedError error
		setupMocks    func()
	}{
		{
			name:          "successful deletion",
			userID:        1,
			expectedError: nil,
			setupMocks: func() {
				mockRepo.On("FindByID", 1).Return(&User{
					ID:       1,
					Username: "testuser",
				}, nil)
				mockCache.On("Delete", "user:exists:1").Once()
				mockCache.On("Delete", "user:id:username:testuser").Once()
				mockCache.On("Delete", "user:hash:1").Once()
				mockCache.On("Delete", "user:info:1").Once()
				mockRepo.On("DeleteUser", 1).Return(nil)
			},
		},
		{
			name:          "deletion failure",
			userID:        2,
			expectedError: errors.New("delete failed"),
			setupMocks: func() {
				mockRepo.On("FindByID", 2).Return(&User{
					ID:       2,
					Username: "testuser2",
				}, nil)
				mockCache.On("Delete", "user:exists:2").Once()
				mockCache.On("Delete", "user:id:username:testuser2").Once()
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

// TestRedisTokenRepository_StoreVerifyToken tests token storage
func TestRedisTokenRepository_StoreVerifyToken(t *testing.T) {
	mockCache := new(MockCacheService)
	repo := NewRedisTokenRepository(mockCache, time.Minute*5)

	token := "test-verification-token"
	ttl := time.Minute * 5

	mockCache.On("SetWithTTL", "verify_token:"+token, "valid", 300).Once()

	err := repo.StoreVerifyToken(token, ttl)
	
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

// TestRedisTokenRepository_VerifyAndConsumeToken tests token verification
func TestRedisTokenRepository_VerifyAndConsumeToken(t *testing.T) {
	mockCache := new(MockCacheService)
	repo := NewRedisTokenRepository(mockCache, time.Minute*5)

	tests := []struct {
		name          string
		token         string
		expectedValid bool
		expectedError error
		setupMocks    func()
	}{
		{
			name:          "valid token",
			token:         "valid-token",
			expectedValid: true,
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Get", "verify_token:valid-token").Return("valid", true)
				mockCache.On("Delete", "verify_token:valid-token").Once()
			},
		},
		{
			name:          "invalid token",
			token:         "invalid-token",
			expectedValid: false,
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Get", "verify_token:invalid-token").Return("invalid", true)
				mockCache.On("Delete", "verify_token:invalid-token").Once()
			},
		},
		{
			name:          "expired or non-existent token",
			token:         "expired-token",
			expectedValid: false,
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Get", "verify_token:expired-token").Return("", false)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			
			valid, err := repo.VerifyAndConsumeToken(tt.token)
			
			assert.Equal(t, tt.expectedValid, valid)
			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			
			mockCache.AssertExpectations(t)
		})
	}
}

// TestRedisRateLimitRepository_TrackLoginAttempt tests login attempt tracking
func TestRedisRateLimitRepository_TrackLoginAttempt(t *testing.T) {
	mockCache := new(MockCacheService)
	mockUserRepo := new(MockUserRepository)
	config := Config{
		MAX_LOGIN_ATTEMPTS: 5,
		LOCKOUT_DURATION:   time.Minute * 15,
	}
	repo := NewRedisRateLimitRepository(mockCache, mockUserRepo, config)

	tests := []struct {
		name          string
		identifier    string
		success       bool
		expectedError error
		setupMocks    func()
	}{
		{
			name:          "successful login - reset attempts",
			identifier:    "testuser",
			success:       true,
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Delete", "login_attempts:testuser").Once()
				mockUserRepo.On("UpdateLockStatus", "testuser", (*time.Time)(nil)).Return(nil)
			},
		},
		{
			name:          "failed login - increment attempts",
			identifier:    "testuser",
			success:       false,
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Increment", "login_attempts:testuser").Return(int64(1))
				mockCache.On("SetExpiration", "login_attempts:testuser", time.Minute*15).Once()
			},
		},
		{
			name:          "failed login - account locked after max attempts",
			identifier:    "lockeduser",
			success:       false,
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Increment", "login_attempts:lockeduser").Return(int64(5))
				mockCache.On("SetExpiration", "login_attempts:lockeduser", time.Minute*15).Once()
				mockCache.On("Set", "login_lockout:lockeduser", "locked").Once()
				mockCache.On("SetExpiration", "login_lockout:lockeduser", time.Minute*15).Once()
				mockUserRepo.On("UpdateLockStatus", "lockeduser", mock.MatchedBy(func(t *time.Time) bool {
					return t != nil && t.After(time.Now())
				})).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			
			err := repo.TrackLoginAttempt(tt.identifier, tt.success)
			
			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			
			mockCache.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

// TestRedisRateLimitRepository_CheckAccountLocked tests account lock checking
func TestRedisRateLimitRepository_CheckAccountLocked(t *testing.T) {
	mockCache := new(MockCacheService)
	mockUserRepo := new(MockUserRepository)
	config := Config{}
	repo := NewRedisRateLimitRepository(mockCache, mockUserRepo, config)

	tests := []struct {
		name          string
		identifier    string
		expectedLock  bool
		expectedError error
		setupMocks    func()
	}{
		{
			name:          "account locked in Redis",
			identifier:    "lockeduser",
			expectedLock:  true,
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Get", "login_lockout:lockeduser").Return("locked", true)
			},
		},
		{
			name:          "account not locked",
			identifier:    "normaluser",
			expectedLock:  false,
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Get", "login_lockout:normaluser").Return("", false)
				mockUserRepo.On("GetLockStatus", "normaluser").Return((*time.Time)(nil), nil)
			},
		},
		{
			name:          "account locked in database - expired",
			identifier:    "expiredlock",
			expectedLock:  false,
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Get", "login_lockout:expiredlock").Return("", false)
				expiredTime := time.Now().Add(-time.Hour)
				mockUserRepo.On("GetLockStatus", "expiredlock").Return(&expiredTime, nil)
				mockUserRepo.On("ClearExpiredLock", "expiredlock").Return(nil)
				mockCache.On("Delete", "login_lockout:expiredlock").Once()
				mockCache.On("Delete", "login_attempts:expiredlock").Once()
			},
		},
		{
			name:          "account locked in database - active",
			identifier:    "activelock",
			expectedLock:  true,
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Get", "login_lockout:activelock").Return("", false)
				activeTime := time.Now().Add(time.Hour)
				mockUserRepo.On("GetLockStatus", "activelock").Return(&activeTime, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			
			locked, err := repo.CheckAccountLocked(tt.identifier)
			
			assert.Equal(t, tt.expectedLock, locked)
			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			
			mockCache.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

// TestRedisRateLimitRepository_TrackIPVisit tests IP visit tracking
func TestRedisRateLimitRepository_TrackIPVisit(t *testing.T) {
	config := Config{
		IP_LIMIT_TIME: time.Minute * 10,
	}

	tests := []struct {
		name          string
		ip            string
		expectedCount int64
		expectedError error
		setupMocks    func(*MockCacheService, *MockUserRepository)
	}{
		{
			name:          "first visit",
			ip:            "192.168.1.1",
			expectedCount: 1,
			expectedError: nil,
			setupMocks: func(mockCache *MockCacheService, mockUserRepo *MockUserRepository) {
				mockCache.On("Increment", "ip_visit:192.168.1.1").Return(int64(1))
				mockCache.On("SetExpiration", "ip_visit:192.168.1.1", time.Minute*10).Once()
			},
		},
		{
			name:          "subsequent visit",
			ip:            "192.168.1.1",
			expectedCount: 5,
			expectedError: nil,
			setupMocks: func(mockCache *MockCacheService, mockUserRepo *MockUserRepository) {
				mockCache.On("Increment", "ip_visit:192.168.1.1").Return(int64(5))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCache := new(MockCacheService)
			mockUserRepo := new(MockUserRepository)
			repo := NewRedisRateLimitRepository(mockCache, mockUserRepo, config)

			tt.setupMocks(mockCache, mockUserRepo)

			count, err := repo.TrackIPVisit(tt.ip)

			assert.Equal(t, tt.expectedCount, count)
			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCache.AssertExpectations(t)
		})
	}
}

// TestRedisRateLimitRepository_BlockIP tests IP blocking
func TestRedisRateLimitRepository_BlockIP(t *testing.T) {
	mockCache := new(MockCacheService)
	mockUserRepo := new(MockUserRepository)
	config := Config{}
	repo := NewRedisRateLimitRepository(mockCache, mockUserRepo, config)

	ip := "192.168.1.100"
	reason := "rate limit exceeded"
	duration := time.Hour

	mockCache.On("SetWithTTL", "ip_blocked:192.168.1.100", "blocked", 3600).Once()
	mockCache.On("Delete", "ip_not_blocked:192.168.1.100").Once()

	err := repo.BlockIP(ip, reason, duration)
	
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}
