package database

import (
	"time"
)

// UserRepository defines the interface for user data access operations
// This interface can be mocked for unit testing
type UserRepository interface {
	// Create operations
	CreateUser(username, passwordHash string) (int, error)

	// Read operations
	FindByID(id int) (*User, error)
	FindByUsername(username string) (*User, error)
	ExistsByID(id int) (bool, error)
	ExistsByUsername(username string) (bool, error)
	GetUserProfile(userID int) (*UserInfo, error)

	// Update operations
	UpdateProfile(userID int, key, value string) error
	UpdateLastLogin(userID int) error
	UpdateLockStatus(identifier string, lockUntil *time.Time) error

	// Delete operations
	DeleteUser(userID int) error

	// Security operations
	GetLockStatus(identifier string) (*time.Time, error)
	ClearExpiredLock(identifier string) error
}

// RateLimitRepository defines the interface for rate limiting operations
type RateLimitRepository interface {
	// Track login attempts
	TrackLoginAttempt(identifier string, success bool) error
	CheckAccountLocked(identifier string) (bool, error)

	// IP rate limiting
	TrackIPVisit(ip string) (int64, error)
	CheckIPBlocked(ip string) (bool, error)
	BlockIP(ip, reason string, duration time.Duration) error
}

// TokenRepository defines the interface for verification token operations
type TokenRepository interface {
	StoreVerifyToken(token string, ttl time.Duration) error
	VerifyAndConsumeToken(token string) (bool, error)
}

// CacheService defines the interface for caching operations
// This abstraction allows switching cache implementations (Redis, in-memory, etc.)
type CacheService interface {
	// Basic cache operations
	Get(key string) (string, bool)
	Set(key, value string)
	SetWithTTL(key, value string, ttlSeconds int)
	Delete(key string)

	// Counter operations
	Increment(key string) int64
	SetExpiration(key string, ttl time.Duration)

	// Special operations
	SetNull(key string) // For caching negative results
}
