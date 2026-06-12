package database

import (
	"context"
	"time"
)

// UserReader defines read operations for user data
type UserReader interface {
	FindByID(id int) (*User, error)
	FindByUsername(username string) (*User, error)
	ExistsByID(id int) (bool, error)
	ExistsByUsername(username string) (bool, error)
	GetUserProfile(userID int) (*UserInfo, error)
	GetUserBasicInfo(userID int) (*UserBasicInfo, error)
	GetLockStatus(identifier string) (*time.Time, error)
	GetPasswordHash(userID int) (string, error)
}

// UserWriter defines write operations for user data
type UserWriter interface {
	CreateUser(username, passwordHash string) (int, error)
	UpdateProfile(userID int, key, value string) error
	UpdateLastLogin(userID int) error
	UpdateLockStatus(identifier string, lockUntil *time.Time) error
	UpdatePassword(userID int, newPasswordHash string) error
	DeleteUser(userID int) error
	ClearExpiredLock(identifier string) error
}

// UserRepository defines the interface for user data access operations
// This interface can be mocked for unit testing
// Deprecated: Use UserReader and UserWriter instead
type UserRepository interface {
	UserReader
	UserWriter
}

// ChatReader defines read operations for chat data
type ChatReader interface {
	GetChatSession(ctx context.Context, sessionID int) (*ChatSession, error)
	GetUsersInChatSession(ctx context.Context, sessionID int) ([]int, error)
	GetUserChatSessions(ctx context.Context, userID int) ([]*ChatSession, error)
	GetMessageIndex(ctx context.Context, messageID int) (*MessageIndex, error)
	GetEncryptedMessage(ctx context.Context, messageID int) (*EncryptedMessage, error)
	GetMessagesForSession(ctx context.Context, sessionID int, limit, offset int) ([]*MessageIndex, error)
	GetMessagesForUser(ctx context.Context, userID int, limit, offset int) ([]*MessageIndex, error)
	GetUnreadMessagesForUser(ctx context.Context, userID int) ([]int, error)
	GetReadReceiptsForMessage(ctx context.Context, messageID int) ([]*MessageReadReceipt, error)
	GetGroupChat(ctx context.Context, sessionID int) (*GroupChat, error)
	GetUserKey(ctx context.Context, userID int) (*UserKey, error)
	GetUserPublicKey(ctx context.Context, userID int) (string, error)
	GetUndeliveredOfflineMessages(ctx context.Context, recipientID int) ([]*OfflineEncryptedMessage, error)
	GetOfflineMessageByID(ctx context.Context, messageID int) (*OfflineEncryptedMessage, error)
	GetUndeliveredMessageCount(ctx context.Context, recipientID int) (int, error)
}

// ChatWriter defines write operations for chat data
type ChatWriter interface {
	CreateChatSession(ctx context.Context, sessionType string, name *string, createdBy *int) (int, error)
	AddUserToChatSession(ctx context.Context, sessionID, userID int) error
	RemoveUserFromChatSession(ctx context.Context, sessionID, userID int) error
	CreateMessageIndex(ctx context.Context, sessionID, senderID int, messageType string, replyToMessageID *int) (int, error)
	UpdateMessageReadStatus(ctx context.Context, messageID int, isRead bool) error
	DeleteMessage(ctx context.Context, messageID int) error
	StoreEncryptedMessage(ctx context.Context, encryptedMsg *EncryptedMessage) error
	MarkMessageAsReadByUser(ctx context.Context, messageID, userID int) error
	CreateGroupChat(ctx context.Context, sessionID int, description, avatarURL *string, maxMembers int, privacyLevel string) error
	UpdateGroupChat(ctx context.Context, sessionID int, description, avatarURL *string, maxMembers int, privacyLevel string) error
	StoreUserKey(ctx context.Context, userKey *UserKey) error
	UpdateUserKey(ctx context.Context, userKey *UserKey) error
	DeactivateUserKey(ctx context.Context, userID int) error
	StoreOfflineEncryptedMessage(ctx context.Context, offlineMsg *OfflineEncryptedMessage) (int, error)
	MarkOfflineMessageDelivered(ctx context.Context, messageID int) error
	DeleteOfflineMessage(ctx context.Context, messageID int) error
}

// ChatRepository defines the interface for chat-related database operations
// Deprecated: Use ChatReader and ChatWriter instead
// Note: This interface is defined in chat_repository.go

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
