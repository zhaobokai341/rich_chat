package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRedisCacheAdapter_Get tests the Get method
func TestRedisCacheAdapter_Get(t *testing.T) {
	mockRedisManager := new(MockRedisManager)

	// Create a RedisManager instance with the mock methods
	redisManager := RedisManager{
		GetCache:         mockRedisManager.GetCache,
		SetCache:         mockRedisManager.SetCache,
		SetNullCache:     mockRedisManager.SetNullCache,
		SetCacheWithTTL:  mockRedisManager.SetCacheWithTTL,
		IncrementCounter: mockRedisManager.IncrementCounter,
		SetKeyExpiration: mockRedisManager.SetKeyExpiration,
		DeleteCache:      mockRedisManager.DeleteCache,
	}

	cacheAdapter := NewRedisCacheAdapter(redisManager)

	tests := []struct {
		name          string
		key           string
		expectedVal   string
		expectedFound bool
		setupMocks    func()
	}{
		{
			name:          "key exists",
			key:           "existing-key",
			expectedVal:   "value",
			expectedFound: true,
			setupMocks: func() {
				mockRedisManager.On("GetCache", "existing-key").Return("value", true)
			},
		},
		{
			name:          "key does not exist",
			key:           "missing-key",
			expectedVal:   "",
			expectedFound: false,
			setupMocks: func() {
				mockRedisManager.On("GetCache", "missing-key").Return("", false)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			val, found := cacheAdapter.Get(tt.key)

			assert.Equal(t, tt.expectedVal, val)
			assert.Equal(t, tt.expectedFound, found)

			mockRedisManager.AssertExpectations(t)
		})
	}
}

// TestRedisCacheAdapter_Set tests the Set method
func TestRedisCacheAdapter_Set(t *testing.T) {
	mockRedisManager := new(MockRedisManager)

	// Create a RedisManager instance with the mock methods
	redisManager := RedisManager{
		GetCache:         mockRedisManager.GetCache,
		SetCache:         mockRedisManager.SetCache,
		SetNullCache:     mockRedisManager.SetNullCache,
		SetCacheWithTTL:  mockRedisManager.SetCacheWithTTL,
		IncrementCounter: mockRedisManager.IncrementCounter,
		SetKeyExpiration: mockRedisManager.SetKeyExpiration,
		DeleteCache:      mockRedisManager.DeleteCache,
	}

	cacheAdapter := NewRedisCacheAdapter(redisManager)

	tests := []struct {
		name       string
		key        string
		value      string
		setupMocks func()
	}{
		{
			name:  "successful set",
			key:   "set-key",
			value: "set-value",
			setupMocks: func() {
				mockRedisManager.On("SetCache", "set-key", "set-value").Return()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			cacheAdapter.Set(tt.key, tt.value)

			mockRedisManager.AssertExpectations(t)
		})
	}
}

// TestRedisCacheAdapter_SetWithTTL tests the SetWithTTL method
func TestRedisCacheAdapter_SetWithTTL(t *testing.T) {
	mockRedisManager := new(MockRedisManager)

	// Create a RedisManager instance with the mock methods
	redisManager := RedisManager{
		GetCache:         mockRedisManager.GetCache,
		SetCache:         mockRedisManager.SetCache,
		SetNullCache:     mockRedisManager.SetNullCache,
		SetCacheWithTTL:  mockRedisManager.SetCacheWithTTL,
		IncrementCounter: mockRedisManager.IncrementCounter,
		SetKeyExpiration: mockRedisManager.SetKeyExpiration,
		DeleteCache:      mockRedisManager.DeleteCache,
	}

	cacheAdapter := NewRedisCacheAdapter(redisManager)

	tests := []struct {
		name       string
		key        string
		value      string
		ttlSeconds int
		setupMocks func()
	}{
		{
			name:       "successful set with TTL",
			key:        "ttl-key",
			value:      "ttl-value",
			ttlSeconds: 300, // 5 minutes
			setupMocks: func() {
				mockRedisManager.On("SetCacheWithTTL", "ttl-key", "ttl-value", 300).Return()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			cacheAdapter.SetWithTTL(tt.key, tt.value, tt.ttlSeconds)

			mockRedisManager.AssertExpectations(t)
		})
	}
}

// TestRedisCacheAdapter_Delete tests the Delete method
func TestRedisCacheAdapter_Delete(t *testing.T) {
	mockRedisManager := new(MockRedisManager)

	// Create a RedisManager instance with the mock methods
	redisManager := RedisManager{
		GetCache:         mockRedisManager.GetCache,
		SetCache:         mockRedisManager.SetCache,
		SetNullCache:     mockRedisManager.SetNullCache,
		SetCacheWithTTL:  mockRedisManager.SetCacheWithTTL,
		IncrementCounter: mockRedisManager.IncrementCounter,
		SetKeyExpiration: mockRedisManager.SetKeyExpiration,
		DeleteCache:      mockRedisManager.DeleteCache,
	}

	cacheAdapter := NewRedisCacheAdapter(redisManager)

	tests := []struct {
		name       string
		key        string
		setupMocks func()
	}{
		{
			name: "successful delete",
			key:  "delete-key",
			setupMocks: func() {
				mockRedisManager.On("DeleteCache", "delete-key").Return()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			cacheAdapter.Delete(tt.key)

			mockRedisManager.AssertExpectations(t)
		})
	}
}

// TestRedisCacheAdapter_Increment tests the Increment method
func TestRedisCacheAdapter_Increment(t *testing.T) {
	mockRedisManager := new(MockRedisManager)

	// Create a RedisManager instance with the mock methods
	redisManager := RedisManager{
		GetCache:         mockRedisManager.GetCache,
		SetCache:         mockRedisManager.SetCache,
		SetNullCache:     mockRedisManager.SetNullCache,
		SetCacheWithTTL:  mockRedisManager.SetCacheWithTTL,
		IncrementCounter: mockRedisManager.IncrementCounter,
		SetKeyExpiration: mockRedisManager.SetKeyExpiration,
		DeleteCache:      mockRedisManager.DeleteCache,
	}

	cacheAdapter := NewRedisCacheAdapter(redisManager)

	tests := []struct {
		name        string
		key         string
		expectedVal int64
		setupMocks  func()
	}{
		{
			name:        "successful increment",
			key:         "inc-key",
			expectedVal: 1,
			setupMocks: func() {
				mockRedisManager.On("IncrementCounter", "inc-key").Return(int64(1))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			val := cacheAdapter.Increment(tt.key)

			assert.Equal(t, tt.expectedVal, val)

			mockRedisManager.AssertExpectations(t)
		})
	}
}

// TestRedisCacheAdapter_SetExpiration tests the SetExpiration method
func TestRedisCacheAdapter_SetExpiration(t *testing.T) {
	mockRedisManager := new(MockRedisManager)

	// Create a RedisManager instance with the mock methods
	redisManager := RedisManager{
		GetCache:         mockRedisManager.GetCache,
		SetCache:         mockRedisManager.SetCache,
		SetNullCache:     mockRedisManager.SetNullCache,
		SetCacheWithTTL:  mockRedisManager.SetCacheWithTTL,
		IncrementCounter: mockRedisManager.IncrementCounter,
		SetKeyExpiration: mockRedisManager.SetKeyExpiration,
		DeleteCache:      mockRedisManager.DeleteCache,
	}

	cacheAdapter := NewRedisCacheAdapter(redisManager)

	tests := []struct {
		name       string
		key        string
		ttl        time.Duration
		setupMocks func()
	}{
		{
			name: "successful expiration set",
			key:  "exp-key",
			ttl:  time.Hour,
			setupMocks: func() {
				mockRedisManager.On("SetKeyExpiration", "exp-key", time.Hour).Return()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			cacheAdapter.SetExpiration(tt.key, tt.ttl)

			mockRedisManager.AssertExpectations(t)
		})
	}
}

// TestRedisCacheAdapter_SetNull tests the SetNull method
func TestRedisCacheAdapter_SetNull(t *testing.T) {
	mockRedisManager := new(MockRedisManager)

	// Create a RedisManager instance with the mock methods
	redisManager := RedisManager{
		GetCache:         mockRedisManager.GetCache,
		SetCache:         mockRedisManager.SetCache,
		SetNullCache:     mockRedisManager.SetNullCache,
		SetCacheWithTTL:  mockRedisManager.SetCacheWithTTL,
		IncrementCounter: mockRedisManager.IncrementCounter,
		SetKeyExpiration: mockRedisManager.SetKeyExpiration,
		DeleteCache:      mockRedisManager.DeleteCache,
	}

	cacheAdapter := NewRedisCacheAdapter(redisManager)

	tests := []struct {
		name       string
		key        string
		setupMocks func()
	}{
		{
			name: "successful null set",
			key:  "null-key",
			setupMocks: func() {
				mockRedisManager.On("SetNullCache", "null-key").Return()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			cacheAdapter.SetNull(tt.key)

			mockRedisManager.AssertExpectations(t)
		})
	}
}
