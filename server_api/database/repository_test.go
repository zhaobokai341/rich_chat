package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

// TestCachedUserRepository_FindByID_CacheHit tests cache hit scenario
func TestCachedUserRepository_FindByID_CacheHit(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockCache)

	userID := 123
	cacheKey := "user:exists:123"

	// Setup expectations
	mockCache.On("Get", cacheKey).Return("true", true)

	// Act - should return from cache without calling database
	exists, err := cachedRepo.ExistsByID(userID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, exists)
	mockRepo.AssertNotCalled(t, "ExistsByID") // Database should not be called
	mockCache.AssertExpectations(t)
}

// TestCachedUserRepository_FindByID_CacheMiss tests cache miss scenario
func TestCachedUserRepository_FindByID_CacheMiss(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockCache)

	userID := 123
	cacheKey := "user:exists:123"
	//expectedUser := &User{ID: userID, Username: "testuser"}

	// Setup expectations
	mockCache.On("Get", cacheKey).Return("", false)
	mockRepo.On("ExistsByID", userID).Return(true, nil)
	mockCache.On("Set", cacheKey, "true")

	// Act
	exists, err := cachedRepo.ExistsByID(userID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, exists)
	mockRepo.AssertExpectations(t) // Database should be called
	mockCache.AssertExpectations(t)
}

// TestCachedUserRepository_CreateUser tests user creation with cache invalidation
func TestCachedUserRepository_CreateUser(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCacheService)
	cachedRepo := NewCachedUserRepository(mockRepo, mockCache)

	username := "testuser"
	passwordHash := "$2a$10$..."
	userID := 1

	// Setup expectations
	mockRepo.On("CreateUser", username, passwordHash).Return(userID, nil)
	mockCache.On("Delete", "user:exists:1")
	mockCache.On("Delete", "user:id:username:testuser")
	mockCache.On("Set", "user:hash:1", passwordHash)

	// Act
	resultID, err := cachedRepo.CreateUser(username, passwordHash)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, userID, resultID)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// TestRedisTokenRepository_VerifyAndConsumeToken tests token verification
func TestRedisTokenRepository_VerifyAndConsumeToken(t *testing.T) {
	// Arrange
	mockCache := new(MockCacheService)
	tokenRepo := NewRedisTokenRepository(mockCache, 5*time.Minute)

	token := "abc123"
	cacheKey := "verify_token:abc123"

	// Setup expectations
	mockCache.On("Get", cacheKey).Return("valid", true)
	mockCache.On("Delete", cacheKey)

	// Act
	valid, err := tokenRepo.VerifyAndConsumeToken(token)

	// Assert
	assert.NoError(t, err)
	assert.True(t, valid)
	mockCache.AssertExpectations(t)
}

// TestRedisTokenRepository_VerifyAndConsumeToken_Invalid tests invalid token
func TestRedisTokenRepository_VerifyAndConsumeToken_Invalid(t *testing.T) {
	// Arrange
	mockCache := new(MockCacheService)
	tokenRepo := NewRedisTokenRepository(mockCache, 5*time.Minute)

	token := "invalid_token"
	cacheKey := "verify_token:invalid_token"

	// Setup expectations - token not found
	mockCache.On("Get", cacheKey).Return("", false)

	// Act
	valid, err := tokenRepo.VerifyAndConsumeToken(token)

	// Assert
	assert.NoError(t, err)
	assert.False(t, valid)
	mockCache.AssertExpectations(t)
}
