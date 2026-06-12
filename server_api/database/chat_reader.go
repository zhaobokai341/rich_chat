package database

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// PostgresChatReader implements ChatReader using PostgreSQL
type PostgresChatReader struct {
	db *sqlx.DB
}

// NewPostgresChatReader creates a new PostgreSQL chat reader
func NewPostgresChatReader(db *sqlx.DB) *PostgresChatReader {
	return &PostgresChatReader{
		db: db,
	}
}

// GetChatSession retrieves a chat session by ID
func (r *PostgresChatReader) GetChatSession(ctx context.Context, sessionID int) (*ChatSession, error) {
	var session ChatSession
	query := `SELECT id, session_type, name, created_by, created_at, updated_at, is_active FROM chat_sessions WHERE id = $1 AND is_active = true`
	err := r.db.GetContext(ctx, &session, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat session: %w", err)
	}
	return &session, nil
}

// GetUsersInChatSession retrieves all users in a chat session
func (r *PostgresChatReader) GetUsersInChatSession(ctx context.Context, sessionID int) ([]int, error) {
	var userIDs []int
	query := `SELECT user_id FROM chat_session_participants WHERE session_id = $1 AND left_at IS NULL`
	err := r.db.SelectContext(ctx, &userIDs, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users in chat session: %w", err)
	}
	return userIDs, nil
}

// GetUserChatSessions retrieves all active chat sessions for a user
func (r *PostgresChatReader) GetUserChatSessions(ctx context.Context, userID int) ([]*ChatSession, error) {
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

// GetMessageIndex retrieves a message index by ID
func (r *PostgresChatReader) GetMessageIndex(ctx context.Context, messageID int) (*MessageIndex, error) {
	var message MessageIndex
	query := `SELECT id, session_id, sender_id, sent_at, updated_at, message_type, is_read, is_deleted, reply_to_message_id FROM message_index WHERE id = $1`
	err := r.db.GetContext(ctx, &message, query, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message index: %w", err)
	}
	return &message, nil
}

// GetEncryptedMessage retrieves an encrypted message from the database
func (r *PostgresChatReader) GetEncryptedMessage(ctx context.Context, messageID int) (*EncryptedMessage, error) {
	var encryptedMsg EncryptedMessage
	query := `SELECT message_id, encrypted_content, encryption_key_id, iv, auth_tag, content_type, file_size, checksum, created_at FROM encrypted_messages WHERE message_id = $1`
	err := r.db.GetContext(ctx, &encryptedMsg, query, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted message: %w", err)
	}
	return &encryptedMsg, nil
}

// GetMessagesForSession retrieves messages for a specific chat session
func (r *PostgresChatReader) GetMessagesForSession(ctx context.Context, sessionID int, limit, offset int) ([]*MessageIndex, error) {
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
func (r *PostgresChatReader) GetMessagesForUser(ctx context.Context, userID int, limit, offset int) ([]*MessageIndex, error) {
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

// GetUnreadMessagesForUser retrieves unread messages for a specific user
func (r *PostgresChatReader) GetUnreadMessagesForUser(ctx context.Context, userID int) ([]int, error) {
	var messageIDs []int
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
func (r *PostgresChatReader) GetReadReceiptsForMessage(ctx context.Context, messageID int) ([]*MessageReadReceipt, error) {
	var receipts []*MessageReadReceipt
	query := `SELECT id, message_id, user_id, read_at FROM message_read_receipts WHERE message_id = $1 ORDER BY read_at DESC`
	err := r.db.SelectContext(ctx, &receipts, query, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get read receipts for message: %w", err)
	}
	return receipts, nil
}

// GetGroupChat retrieves group chat information
func (r *PostgresChatReader) GetGroupChat(ctx context.Context, sessionID int) (*GroupChat, error) {
	var groupChat GroupChat
	query := `SELECT session_id, description, avatar_url, max_members, privacy_level FROM group_chats WHERE session_id = $1`
	err := r.db.GetContext(ctx, &groupChat, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group chat: %w", err)
	}
	return &groupChat, nil
}

// GetUserKey retrieves a user's encryption key pair
func (r *PostgresChatReader) GetUserKey(ctx context.Context, userID int) (*UserKey, error) {
	var userKey UserKey
	query := `
		SELECT id, user_id, public_key, encrypted_private_key, key_algorithm, created_at, updated_at, is_active
		FROM user_keys
		WHERE user_id = $1 AND is_active = true`
	err := r.db.GetContext(ctx, &userKey, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user key: %w", err)
	}
	return &userKey, nil
}

// GetUserPublicKey retrieves only the public key for a user
func (r *PostgresChatReader) GetUserPublicKey(ctx context.Context, userID int) (string, error) {
	var publicKey string
	query := `SELECT public_key FROM user_keys WHERE user_id = $1 AND is_active = true`
	err := r.db.GetContext(ctx, &publicKey, query, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user public key: %w", err)
	}
	return publicKey, nil
}

// GetUndeliveredOfflineMessages retrieves all undelivered encrypted messages for a user
func (r *PostgresChatReader) GetUndeliveredOfflineMessages(ctx context.Context, recipientID int) ([]*OfflineEncryptedMessage, error) {
	var messages []*OfflineEncryptedMessage
	query := `
		SELECT id, recipient_id, sender_id, session_id, encrypted_session_key, encrypted_content, iv, auth_tag, created_at, delivered_at, is_delivered
		FROM offline_encrypted_messages
		WHERE recipient_id = $1 AND is_delivered = false
		ORDER BY created_at ASC`
	err := r.db.SelectContext(ctx, &messages, query, recipientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get undelivered offline messages: %w", err)
	}
	return messages, nil
}

// GetOfflineMessageByID retrieves a specific offline encrypted message by ID
func (r *PostgresChatReader) GetOfflineMessageByID(ctx context.Context, messageID int) (*OfflineEncryptedMessage, error) {
	var message OfflineEncryptedMessage
	query := `
		SELECT id, recipient_id, sender_id, session_id, encrypted_session_key, encrypted_content, iv, auth_tag, created_at, delivered_at, is_delivered
		FROM offline_encrypted_messages
		WHERE id = $1`
	err := r.db.GetContext(ctx, &message, query, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get offline message by ID: %w", err)
	}
	return &message, nil
}

// GetUndeliveredMessageCount returns the count of undelivered messages for a user
func (r *PostgresChatReader) GetUndeliveredMessageCount(ctx context.Context, recipientID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM offline_encrypted_messages WHERE recipient_id = $1 AND is_delivered = false`
	err := r.db.GetContext(ctx, &count, query, recipientID)
	if err != nil {
		return 0, fmt.Errorf("failed to get undelivered message count: %w", err)
	}
	return count, nil
}
