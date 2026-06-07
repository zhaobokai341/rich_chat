package database

import (
	"context"
)

// ChatRepository defines the interface for chat-related database operations
type ChatRepository interface {
	// Chat Session Management
	CreateChatSession(ctx context.Context, sessionType string, name *string, createdBy *int) (int, error)
	GetChatSession(ctx context.Context, sessionID int) (*ChatSession, error)
	AddUserToChatSession(ctx context.Context, sessionID, userID int) error
	RemoveUserFromChatSession(ctx context.Context, sessionID, userID int) error
	GetUsersInChatSession(ctx context.Context, sessionID int) ([]int, error)
	GetUserChatSessions(ctx context.Context, userID int) ([]*ChatSession, error)

	// Message Operations
	CreateMessageIndex(ctx context.Context, sessionID, senderID int, messageType string, replyToMessageID *int) (int, error)
	GetMessageIndex(ctx context.Context, messageID int) (*MessageIndex, error)
	UpdateMessageReadStatus(ctx context.Context, messageID int, isRead bool) error
	DeleteMessage(ctx context.Context, messageID int) error

	// Encrypted Message Operations
	StoreEncryptedMessage(ctx context.Context, encryptedMsg *EncryptedMessage) error
	GetEncryptedMessage(ctx context.Context, messageID int) (*EncryptedMessage, error)

	// Message History
	GetMessagesForSession(ctx context.Context, sessionID int, limit, offset int) ([]*MessageIndex, error)
	GetMessagesForUser(ctx context.Context, userID int, limit, offset int) ([]*MessageIndex, error)

	// Read Receipts
	MarkMessageAsReadByUser(ctx context.Context, messageID, userID int) error
	GetUnreadMessagesForUser(ctx context.Context, userID int) ([]int, error)
	GetReadReceiptsForMessage(ctx context.Context, messageID int) ([]*MessageReadReceipt, error)

	// Group Chat Operations
	CreateGroupChat(ctx context.Context, sessionID int, description, avatarURL *string, maxMembers int, privacyLevel string) error
	UpdateGroupChat(ctx context.Context, sessionID int, description, avatarURL *string, maxMembers int, privacyLevel string) error
	GetGroupChat(ctx context.Context, sessionID int) (*GroupChat, error)
}
