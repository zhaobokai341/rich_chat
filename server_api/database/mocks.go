package database

import (
	"time"

	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(username, passwordHash string) (int, error) {
	args := m.Called(username, passwordHash)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) FindByID(id int) (*User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(username string) (*User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) ExistsByID(id int) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByUsername(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) GetUserProfile(userID int) (*UserInfo, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserInfo), args.Error(1)
}

func (m *MockUserRepository) UpdateProfile(userID int, key, value string) error {
	args := m.Called(userID, key, value)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(userID int) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLockStatus(identifier string, lockUntil *time.Time) error {
	args := m.Called(identifier, lockUntil)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePassword(userID int, newPasswordHash string) error {
	args := m.Called(userID, newPasswordHash)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(userID int) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockUserRepository) GetLockStatus(identifier string) (*time.Time, error) {
	args := m.Called(identifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*time.Time), args.Error(1)
}

func (m *MockUserRepository) ClearExpiredLock(identifier string) error {
	args := m.Called(identifier)
	return args.Error(0)
}

// MockRateLimitRepository is a mock implementation of RateLimitRepository for testing
type MockRateLimitRepository struct {
	mock.Mock
}

func (m *MockRateLimitRepository) TrackLoginAttempt(identifier string, success bool) error {
	args := m.Called(identifier, success)
	return args.Error(0)
}

func (m *MockRateLimitRepository) CheckAccountLocked(identifier string) (bool, error) {
	args := m.Called(identifier)
	return args.Bool(0), args.Error(1)
}

func (m *MockRateLimitRepository) TrackIPVisit(ip string) (int64, error) {
	args := m.Called(ip)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRateLimitRepository) CheckIPBlocked(ip string) (bool, error) {
	args := m.Called(ip)
	return args.Bool(0), args.Error(1)
}

func (m *MockRateLimitRepository) BlockIP(ip, reason string, duration time.Duration) error {
	args := m.Called(ip, reason, duration)
	return args.Error(0)
}

// MockTokenRepository is a mock implementation of TokenRepository for testing
type MockTokenRepository struct {
	mock.Mock
}

func (m *MockTokenRepository) StoreVerifyToken(token string, ttl time.Duration) error {
	args := m.Called(token, ttl)
	return args.Error(0)
}

func (m *MockTokenRepository) VerifyAndConsumeToken(token string) (bool, error) {
	args := m.Called(token)
	return args.Bool(0), args.Error(1)
}

// MockCacheService is a mock implementation of CacheService for testing
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Get(key string) (string, bool) {
	args := m.Called(key)
	return args.String(0), args.Bool(1)
}

func (m *MockCacheService) Set(key, value string) {
	m.Called(key, value)
}

func (m *MockCacheService) SetWithTTL(key, value string, ttlSeconds int) {
	m.Called(key, value, ttlSeconds)
}

func (m *MockCacheService) Delete(key string) {
	m.Called(key)
}

func (m *MockCacheService) Increment(key string) int64 {
	args := m.Called(key)
	return args.Get(0).(int64)
}

func (m *MockCacheService) SetExpiration(key string, ttl time.Duration) {
	m.Called(key, ttl)
}

func (m *MockCacheService) SetNull(key string) {
	m.Called(key)
}

// MockRedisManager is a mock implementation of RedisManager for testing
type MockRedisManager struct {
	mock.Mock
}

// NewMockRedisManager creates a new MockRedisManager with default implementations
func NewMockRedisManager() *MockRedisManager {
	return &MockRedisManager{}
}

// MockRedisManagerWithSetup creates a new MockRedisManager with custom implementations
func MockRedisManagerWithSetup(setup func(*MockRedisManager)) *MockRedisManager {
	mock := NewMockRedisManager()
	if setup != nil {
		setup(mock)
	}
	return mock
}

// Implement RedisManager interface methods
func (m *MockRedisManager) GetCache(key string) (string, bool) {
	args := m.Called(key)
	return args.String(0), args.Bool(1)
}

func (m *MockRedisManager) GetIntValue(key string) (int64, bool) {
	args := m.Called(key)
	return args.Get(0).(int64), args.Bool(1)
}

func (m *MockRedisManager) SetCache(key string, value string) {
	m.Called(key, value)
}

func (m *MockRedisManager) SetNullCache(key string) {
	m.Called(key)
}

func (m *MockRedisManager) SetCacheWithTTL(key string, value string, ttl int) {
	m.Called(key, value, ttl)
}

func (m *MockRedisManager) SetLockoutCache(key string) {
	m.Called(key)
}

func (m *MockRedisManager) IncrementCounter(key string) int64 {
	args := m.Called(key)
	return args.Get(0).(int64)
}

func (m *MockRedisManager) SetKeyExpiration(key string, ttl time.Duration) {
	m.Called(key, ttl)
}

func (m *MockRedisManager) DeleteCache(key string) {
	m.Called(key)
}

// ToRedisManager converts MockRedisManager to RedisManager struct for compatibility
func (m *MockRedisManager) ToRedisManager() RedisManager {
	return RedisManager{
		GetCache:         m.GetCache,
		GetIntValue:      m.GetIntValue,
		SetCache:         m.SetCache,
		SetNullCache:     m.SetNullCache,
		SetCacheWithTTL:  m.SetCacheWithTTL,
		SetLockoutCache:  m.SetLockoutCache,
		IncrementCounter: m.IncrementCounter,
		SetKeyExpiration: m.SetKeyExpiration,
		DeleteCache:      m.DeleteCache,
	}
}
