package database

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// PostgresChatRepository implements ChatRepository using PostgreSQL
type PostgresChatRepository struct {
	db *sqlx.DB
}

// NewPostgresChatRepository creates a new PostgreSQL chat repository
func NewPostgresChatRepository(db *sqlx.DB) *PostgresChatRepository {
	return &PostgresChatRepository{
		db: db,
	}
}

// Chat Session Management

// CreateChatSession creates a new chat session (direct or group)
func (r *PostgresChatRepository) CreateChatSession(ctx context.Context, sessionType string, name *string, createdBy *int) (int, error) {
	var sessionID int
	query := `INSERT INTO chat_sessions (session_type, name, created_by) VALUES ($1, $2, $3) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, sessionType, name, createdBy).Scan(&sessionID)
	if err != nil {
		return 0, fmt.Errorf("failed to create chat session: %w", err)
	}
	return sessionID, nil
}

// GetChatSession retrieves a chat session by ID
func (r *PostgresChatRepository) GetChatSession(ctx context.Context, sessionID int) (*ChatSession, error) {
	var session ChatSession
	query := `SELECT id, session_type, name, created_by, created_at, updated_at, is_active FROM chat_sessions WHERE id = $1 AND is_active = true`
	err := r.db.GetContext(ctx, &session, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat session: %w", err)
	}
	return &session, nil
}

// AddUserToChatSession adds a user to a chat session
func (r *PostgresChatRepository) AddUserToChatSession(ctx context.Context, sessionID, userID int) error {
	query := `INSERT INTO chat_session_participants (session_id, user_id) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to add user to chat session: %w", err)
	}
	return nil
}

// RemoveUserFromChatSession removes a user from a chat session
func (r *PostgresChatRepository) RemoveUserFromChatSession(ctx context.Context, sessionID, userID int) error {
	query := `UPDATE chat_session_participants SET left_at = NOW() WHERE session_id = $1 AND user_id = $2 AND left_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user from chat session: %w", err)
	}
	return nil
}

// GetUsersInChatSession retrieves all users in a chat session
func (r *PostgresChatRepository) GetUsersInChatSession(ctx context.Context, sessionID int) ([]int, error) {
	var userIDs []int
	query := `SELECT user_id FROM chat_session_participants WHERE session_id = $1 AND left_at IS NULL`
	err := r.db.SelectContext(ctx, &userIDs, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users in chat session: %w", err)
	}
	return userIDs, nil
}

// GetUserChatSessions retrieves all active chat sessions for a user
func (r *PostgresChatRepository) GetUserChatSessions(ctx context.Context, userID int) ([]*ChatSession, error) {
	var sessions []*ChatSession
	query := `
		SELECT cs.id, cs.session_type, cs.name, cs.created_by, cs.created_at, cs.updated_at, cs.is_active
		FROM chat_sessions cs
		JOIN chat_session_participants csp ON cs.id = csp.session_id
		WHERE csp.user_id = $1 AND csp.left_at IS NULL AND cs.is_active = true
		ORDER BY cs.updated_at DESC`
	err := r.db.SelectContext(ctx, &sessions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user chat sessions: %w", err)
	}
	return sessions, nil
}

// Message Operations

// CreateMessageIndex creates a new message index entry
func (r *PostgresChatRepository) CreateMessageIndex(ctx context.Context, sessionID, senderID int, messageType string, replyToMessageID *int) (int, error) {
	var messageID int
	query := `INSERT INTO message_index (session_id, sender_id, message_type, reply_to_message_id) VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, sessionID, senderID, messageType, replyToMessageID).Scan(&messageID)
	if err != nil {
		return 0, fmt.Errorf("failed to create message index: %w", err)
	}
	return messageID, nil
}

// GetMessageIndex retrieves a message index by ID
func (r *PostgresChatRepository) GetMessageIndex(ctx context.Context, messageID int) (*MessageIndex, error) {
	var message MessageIndex
	query := `SELECT id, session_id, sender_id, sent_at, updated_at, message_type, is_read, is_deleted, reply_to_message_id FROM message_index WHERE id = $1`
	err := r.db.GetContext(ctx, &message, query, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message index: %w", err)
	}
	return &message, nil
}

// UpdateMessageReadStatus updates the read status of a message
func (r *PostgresChatRepository) UpdateMessageReadStatus(ctx context.Context, messageID int, isRead bool) error {
	query := `UPDATE message_index SET is_read = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, isRead, messageID)
	if err != nil {
		return fmt.Errorf("failed to update message read status: %w", err)
	}
	return nil
}

// DeleteMessage marks a message as deleted (soft delete)
func (r *PostgresChatRepository) DeleteMessage(ctx context.Context, messageID int) error {
	query := `UPDATE message_index SET is_deleted = true, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, messageID)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}

// Encrypted Message Operations

// StoreEncryptedMessage stores an encrypted message in the database
func (r *PostgresChatRepository) StoreEncryptedMessage(ctx context.Context, encryptedMsg *EncryptedMessage) error {
	query := `
		INSERT INTO encrypted_messages (message_id, encrypted_content, encryption_key_id, iv, auth_tag, content_type, file_size, checksum)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, query,
		encryptedMsg.MessageID,
		encryptedMsg.EncryptedContent,
		encryptedMsg.EncryptionKeyID,
		encryptedMsg.Iv,
		encryptedMsg.AuthTag,
		encryptedMsg.ContentType,
		encryptedMsg.FileSize,
		encryptedMsg.Checksum)
	if err != nil {
		return fmt.Errorf("failed to store encrypted message: %w", err)
	}
	return nil
}

// GetEncryptedMessage retrieves an encrypted message from the database
func (r *PostgresChatRepository) GetEncryptedMessage(ctx context.Context, messageID int) (*EncryptedMessage, error) {
	var encryptedMsg EncryptedMessage
	query := `SELECT message_id, encrypted_content, encryption_key_id, iv, auth_tag, content_type, file_size, checksum, created_at FROM encrypted_messages WHERE message_id = $1`
	err := r.db.GetContext(ctx, &encryptedMsg, query, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted message: %w", err)
	}
	return &encryptedMsg, nil
}

// Message History

// GetMessagesForSession retrieves messages for a specific chat session
func (r *PostgresChatRepository) GetMessagesForSession(ctx context.Context, sessionID int, limit, offset int) ([]*MessageIndex, error) {
	var messages []*MessageIndex
	query := `
		SELECT id, session_id, sender_id, sent_at, updated_at, message_type, is_read, is_deleted, reply_to_message_id
		FROM message_index
		WHERE session_id = $1 AND is_deleted = false
		ORDER BY sent_at DESC
		LIMIT $2 OFFSET $3`
	err := r.db.SelectContext(ctx, &messages, query, sessionID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages for session: %w", err)
	}
	return messages, nil
}

// GetMessagesForUser retrieves messages sent by a specific user
func (r *PostgresChatRepository) GetMessagesForUser(ctx context.Context, userID int, limit, offset int) ([]*MessageIndex, error) {
	var messages []*MessageIndex
	query := `
		SELECT id, session_id, sender_id, sent_at, updated_at, message_type, is_read, is_deleted, reply_to_message_id
		FROM message_index
		WHERE sender_id = $1 AND is_deleted = false
		ORDER BY sent_at DESC
		LIMIT $2 OFFSET $3`
	err := r.db.SelectContext(ctx, &messages, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages for user: %w", err)
	}
	return messages, nil
}

// Read Receipts

// MarkMessageAsReadByUser marks a message as read by a specific user
func (r *PostgresChatRepository) MarkMessageAsReadByUser(ctx context.Context, messageID, userID int) error {
	query := `INSERT INTO message_read_receipts (message_id, user_id) VALUES ($1, $2) ON CONFLICT (message_id, user_id) DO UPDATE SET read_at = NOW()`
	_, err := r.db.ExecContext(ctx, query, messageID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark message as read by user: %w", err)
	}
	return nil
}

// GetUnreadMessagesForUser retrieves unread messages for a specific user
func (r *PostgresChatRepository) GetUnreadMessagesForUser(ctx context.Context, userID int) ([]int, error) {
	var messageIDs []int
	// This query finds messages in sessions the user participates in that are not marked as read by the user
	query := `
		SELECT mi.id
		FROM message_index mi
		JOIN chat_session_participants csp ON mi.session_id = csp.session_id
		LEFT JOIN message_read_receipts mrr ON mi.id = mrr.message_id AND mrr.user_id = $1
		WHERE csp.user_id = $1 AND csp.left_at IS NULL AND mi.is_read = false AND mrr.message_id IS NULL
		AND mi.is_deleted = false
		ORDER BY mi.sent_at DESC`
	err := r.db.SelectContext(ctx, &messageIDs, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread messages for user: %w", err)
	}
	return messageIDs, nil
}

// GetReadReceiptsForMessage retrieves read receipts for a specific message
func (r *PostgresChatRepository) GetReadReceiptsForMessage(ctx context.Context, messageID int) ([]*MessageReadReceipt, error) {
	var receipts []*MessageReadReceipt
	query := `SELECT id, message_id, user_id, read_at FROM message_read_receipts WHERE message_id = $1 ORDER BY read_at DESC`
	err := r.db.SelectContext(ctx, &receipts, query, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get read receipts for message: %w", err)
	}
	return receipts, nil
}

// Group Chat Operations

// CreateGroupChat creates a new group chat
func (r *PostgresChatRepository) CreateGroupChat(ctx context.Context, sessionID int, description, avatarURL *string, maxMembers int, privacyLevel string) error {
	query := `INSERT INTO group_chats (session_id, description, avatar_url, max_members, privacy_level) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, sessionID, description, avatarURL, maxMembers, privacyLevel)
	if err != nil {
		return fmt.Errorf("failed to create group chat: %w", err)
	}
	return nil
}

// UpdateGroupChat updates group chat information
func (r *PostgresChatRepository) UpdateGroupChat(ctx context.Context, sessionID int, description, avatarURL *string, maxMembers int, privacyLevel string) error {
	query := `UPDATE group_chats SET description = $1, avatar_url = $2, max_members = $3, privacy_level = $4, updated_at = NOW() WHERE session_id = $5`
	_, err := r.db.ExecContext(ctx, query, description, avatarURL, maxMembers, privacyLevel, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update group chat: %w", err)
	}
	return nil
}

// GetGroupChat retrieves group chat information
func (r *PostgresChatRepository) GetGroupChat(ctx context.Context, sessionID int) (*GroupChat, error) {
	var groupChat GroupChat
	query := `SELECT session_id, description, avatar_url, max_members, privacy_level FROM group_chats WHERE session_id = $1`
	err := r.db.GetContext(ctx, &groupChat, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group chat: %w", err)
	}
	return &groupChat, nil
}
