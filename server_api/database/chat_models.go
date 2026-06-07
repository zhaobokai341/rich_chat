package database

import (
	"time"
)

// ChatSession represents a chat session (either direct or group chat)
type ChatSession struct {
	ID          int        `db:"id"`
	SessionType string     `db:"session_type"` // 'direct' or 'group'
	Name        *string    `db:"name"`         // for group chats only
	CreatedBy   *int       `db:"created_by"`   // user who created the session (for groups)
	CreatedAt   *time.Time `db:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at"`
	IsActive    bool       `db:"is_active"`
}

// ChatSessionParticipant represents the relationship between users and chat sessions
type ChatSessionParticipant struct {
	SessionID int        `db:"session_id"`
	UserID    int        `db:"user_id"`
	JoinedAt  *time.Time `db:"joined_at"`
	LeftAt    *time.Time `db:"left_at"` // NULL means still participating
}

// MessageIndex represents the unencrypted metadata of a message
type MessageIndex struct {
	ID               int        `db:"id"`
	SessionID        int        `db:"session_id"`
	SenderID         int        `db:"sender_id"`
	SentAt           *time.Time `db:"sent_at"`
	UpdatedAt        *time.Time `db:"updated_at"`
	MessageType      string     `db:"message_type"`        // 'text', 'image', 'file', 'audio', 'video', 'system'
	IsRead           bool       `db:"is_read"`             // for read receipts
	IsDeleted        bool       `db:"is_deleted"`          // soft delete flag
	ReplyToMessageID *int       `db:"reply_to_message_id"` // for reply threads
}

// EncryptedMessage represents the encrypted content of a message
type EncryptedMessage struct {
	MessageID        int        `db:"message_id"`
	EncryptedContent string     `db:"encrypted_content"` // AES-256-GCM encrypted content
	EncryptionKeyID  *string    `db:"encryption_key_id"` // reference to which key was used for encryption
	Iv               []byte     `db:"iv"`                // initialization vector for AES decryption
	AuthTag          []byte     `db:"auth_tag"`          // authentication tag for AES-GCM
	ContentType      string     `db:"content_type"`      // MIME type of original content
	FileSize         *int       `db:"file_size"`         // for file attachments
	Checksum         *string    `db:"checksum"`          // SHA-256 checksum of original content before encryption
	CreatedAt        *time.Time `db:"created_at"`
}

// MessageReadReceipt represents who has read which messages
type MessageReadReceipt struct {
	ID        int        `db:"id"`
	MessageID int        `db:"message_id"`
	UserID    int        `db:"user_id"`
	ReadAt    *time.Time `db:"read_at"`
}

// GroupChat represents additional metadata for group chats
type GroupChat struct {
	SessionID    int     `db:"session_id"`
	Description  *string `db:"description"`
	AvatarURL    *string `db:"avatar_url"`
	MaxMembers   int     `db:"max_members"`
	PrivacyLevel string  `db:"privacy_level"` // 'public', 'private', 'invite_only'
}
