package database

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRedisRateLimitRepository_TrackLoginAttempt tests the TrackLoginAttempt method
func TestRedisRateLimitRepository_TrackLoginAttempt(t *testing.T) {
	mockCache := new(MockCacheService)
	mockUserRepo := new(MockUserRepository)
	config := Config{
		MAX_LOGIN_ATTEMPTS: 5,
		LOCKOUT_DURATION:   time.Hour,
	}
	rateLimitRepo := NewRedisRateLimitRepository(mockCache, mockUserRepo, config)

	tests := []struct {
		name          string
		username      string
		success       bool
		expectedError error
		setupMocks    func()
	}{
		{
			name:          "successful login attempt tracking",
			username:      "testuser",
			success:       true,
			expectedError: nil,
			setupMocks: func() {
				mockCache.On("Delete", "login_attempts:testuser").Return()
				mockUserRepo.On("UpdateLockStatus", "testuser", (*time.Time)(nil)).Return(nil)
			},
		},
		{
			name:          "failed login attempt tracking",
			username:      "testuser",
			success:       false,
			expectedError: nil, // Note: TrackLoginAttempt doesn't return errors in the implementation
			setupMocks: func() {
				mockCache.On("Increment", "login_attempts:testuser").Return(int64(1))
				mockCache.On("SetExpiration", "login_attempts:testuser", time.Hour).Return()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := rateLimitRepo.TrackLoginAttempt(tt.username, tt.success)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockCache.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

// TestRedisRateLimitRepository_CheckAccountLocked tests the CheckAccountLocked method
func TestRedisRateLimitRepository_CheckAccountLocked(t *testing.T) {
	mockCache := new(MockCacheService)
	mockUserRepo := new(MockUserRepository)
	config := Config{
		MAX_LOGIN_ATTEMPTS: 5,
		LOCKOUT_DURATION:   time.Hour,
	}
	rateLimitRepo := NewRedisRateLimitRepository(mockCache, mockUserRepo, config)

	tests := []struct {
		name             string
		username         string
		expectedIsLocked bool
		expectedError    error
		setupMocks       func()
	}{
		{
			name:             "account not locked - Redis cache miss, DB not locked",
			username:         "activeuser",
			expectedIsLocked: false,
			expectedError:    nil,
			setupMocks: func() {
				mockCache.On("Get", "login_lockout:activeuser").Return("", false)             // Not locked in Redis
				mockUserRepo.On("GetLockStatus", "activeuser").Return((*time.Time)(nil), nil) // Not locked in DB
			},
		},
		{
			name:             "account locked - in Redis",
			username:         "lockeduser",
			expectedIsLocked: true,
			expectedError:    nil,
			setupMocks: func() {
				mockCache.On("Get", "login_lockout:lockeduser").Return("locked", true) // Locked in Redis
			},
		},
		{
			name:             "account locked - in DB",
			username:         "dblockeduser",
			expectedIsLocked: true,
			expectedError:    nil,
			setupMocks: func() {
				futureTime := time.Now().Add(1 * time.Hour)                               // Future time means still locked
				mockCache.On("Get", "login_lockout:dblockeduser").Return("", false)       // Not locked in Redis
				mockUserRepo.On("GetLockStatus", "dblockeduser").Return(&futureTime, nil) // Locked in DB with future time
			},
		},
		{
			name:             "expired lock - cleared from DB",
			username:         "expireduser",
			expectedIsLocked: false,
			expectedError:    nil,
			setupMocks: func() {
				pastTime := time.Now().Add(-1 * time.Hour)                             // Past time means expired
				mockCache.On("Get", "login_lockout:expireduser").Return("", false)     // Not locked in Redis
				mockUserRepo.On("GetLockStatus", "expireduser").Return(&pastTime, nil) // Expired lock in DB
				mockUserRepo.On("ClearExpiredLock", "expireduser").Return(nil)         // Should clear expired lock
				mockCache.On("Delete", "login_lockout:expireduser").Return()           // Delete Redis lock
				mockCache.On("Delete", "login_attempts:expireduser").Return()          // Delete attempts counter
			},
		},
		{
			name:             "DB error",
			username:         "erroruser",
			expectedIsLocked: false,
			expectedError:    errors.New("database error"),
			setupMocks: func() {
				mockCache.On("Get", "login_lockout:erroruser").Return("", false)                                      // Not locked in Redis
				mockUserRepo.On("GetLockStatus", "erroruser").Return((*time.Time)(nil), errors.New("database error")) // DB error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			isLocked, err := rateLimitRepo.CheckAccountLocked(tt.username)

			assert.Equal(t, tt.expectedIsLocked, isLocked)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockCache.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

// TestRedisRateLimitRepository_TrackIPVisit tests the TrackIPVisit method
func TestRedisRateLimitRepository_TrackIPVisit(t *testing.T) {
	mockCache := new(MockCacheService)
	mockUserRepo := new(MockUserRepository)
	config := Config{
		MAX_LOGIN_ATTEMPTS: 5,
		LOCKOUT_DURATION:   time.Hour,
		IP_LIMIT_TIME:      24 * time.Hour, // Add IP limit time to config
	}
	rateLimitRepo := NewRedisRateLimitRepository(mockCache, mockUserRepo, config)

	tests := []struct {
		name        string
		ipAddress   string
		expectedErr error
		setupMocks  func()
	}{
		{
			name:        "successful IP visit tracking - first visit",
			ipAddress:   "192.168.1.1",
			expectedErr: nil,
			setupMocks: func() {
				mockCache.On("Increment", "ip_visit:192.168.1.1").Return(int64(1))           // First visit
				mockCache.On("SetExpiration", "ip_visit:192.168.1.1", 24*time.Hour).Return() // Set expiration
			},
		},
		{
			name:        "successful IP visit tracking - subsequent visit",
			ipAddress:   "192.168.1.2",
			expectedErr: nil,
			setupMocks: func() {
				mockCache.On("Increment", "ip_visit:192.168.1.2").Return(int64(2)) // Second visit, no SetExpiration called
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			_, err := rateLimitRepo.TrackIPVisit(tt.ipAddress)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockCache.AssertExpectations(t)
		})
	}
}

// TestRedisRateLimitRepository_BlockIP tests the BlockIP method
func TestRedisRateLimitRepository_BlockIP(t *testing.T) {
	mockCache := new(MockCacheService)
	mockUserRepo := new(MockUserRepository)
	config := Config{
		MAX_LOGIN_ATTEMPTS: 5,
		LOCKOUT_DURATION:   time.Hour,
	}
	rateLimitRepo := NewRedisRateLimitRepository(mockCache, mockUserRepo, config)

	tests := []struct {
		name        string
		ipAddress   string
		reason      string
		duration    time.Duration
		expectedErr error
		setupMocks  func()
	}{
		{
			name:        "successful IP blocking",
			ipAddress:   "192.168.1.1",
			reason:      "brute_force",
			duration:    time.Hour,
			expectedErr: nil,
			setupMocks: func() {
				mockCache.On("SetWithTTL", "ip_blocked:192.168.1.1", "blocked", 3600).Return() // TTL in seconds
				mockCache.On("Delete", "ip_not_blocked:192.168.1.1").Return()                  // Delete negative cache
			},
		},
		{
			name:        "IP blocking with different duration",
			ipAddress:   "192.168.1.2",
			reason:      "suspicious_activity",
			duration:    2 * time.Hour,
			expectedErr: nil,
			setupMocks: func() {
				mockCache.On("SetWithTTL", "ip_blocked:192.168.1.2", "blocked", 7200).Return() // 2 hours = 7200 seconds
				mockCache.On("Delete", "ip_not_blocked:192.168.1.2").Return()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := rateLimitRepo.BlockIP(tt.ipAddress, tt.reason, tt.duration)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockCache.AssertExpectations(t)
		})
	}
}

// TestRedisRateLimitRepository_CheckIPBlocked tests the CheckIPBlocked method
func TestRedisRateLimitRepository_CheckIPBlocked(t *testing.T) {
	mockCache := new(MockCacheService)
	mockUserRepo := new(MockUserRepository)
	config := Config{
		MAX_LOGIN_ATTEMPTS: 5,
		LOCKOUT_DURATION:   time.Hour,
	}
	rateLimitRepo := NewRedisRateLimitRepository(mockCache, mockUserRepo, config)

	tests := []struct {
		name            string
		ipAddress       string
		expectedBlocked bool
		expectedError   error
		setupMocks      func()
	}{
		{
			name:            "IP not blocked - cache miss",
			ipAddress:       "192.168.1.1",
			expectedBlocked: false,
			expectedError:   nil,
			setupMocks: func() {
				mockCache.On("Get", "ip_blocked:192.168.1.1").Return("", false)                      // Not found in blocked cache
				mockCache.On("Get", "ip_not_blocked:192.168.1.1").Return("", false)                  // Not found in negative cache
				mockCache.On("SetWithTTL", "ip_not_blocked:192.168.1.1", "not_blocked", 60).Return() // Set negative cache
			},
		},
		{
			name:            "IP blocked",
			ipAddress:       "192.168.1.2",
			expectedBlocked: true,
			expectedError:   nil,
			setupMocks: func() {
				mockCache.On("Get", "ip_blocked:192.168.1.2").Return("blocked", true) // Found in blocked cache
			},
		},
		{
			name:            "IP not blocked - negative cache hit",
			ipAddress:       "192.168.1.3",
			expectedBlocked: false,
			expectedError:   nil,
			setupMocks: func() {
				mockCache.On("Get", "ip_blocked:192.168.1.3").Return("", false)               // Not found in blocked cache
				mockCache.On("Get", "ip_not_blocked:192.168.1.3").Return("not_blocked", true) // Found in negative cache
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			blocked, err := rateLimitRepo.CheckIPBlocked(tt.ipAddress)

			assert.Equal(t, tt.expectedBlocked, blocked)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockCache.AssertExpectations(t)
		})
	}
}
