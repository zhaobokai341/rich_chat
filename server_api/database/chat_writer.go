package database

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

// PostgresChatWriter implements ChatWriter using PostgreSQL
type PostgresChatWriter struct {
	db *sqlx.DB
}

// NewPostgresChatWriter creates a new PostgreSQL chat writer
func NewPostgresChatWriter(db *sqlx.DB) *PostgresChatWriter {
	return &PostgresChatWriter{
		db: db,
	}
}

// CreateChatSession creates a new chat session
func (w *PostgresChatWriter) CreateChatSession(ctx context.Context, sessionType string, name *string, createdBy *int) (int, error) {
	var sessionID int
	query := `INSERT INTO chat_sessions (session_type, name, created_by, created_at, updated_at, is_active) VALUES ($1, $2, $3, NOW(), NOW(), true) RETURNING id`
	err := w.db.QueryRowContext(ctx, query, sessionType, name, createdBy).Scan(&sessionID)
	if err != nil {
		return 0, fmt.Errorf("failed to create chat session: %w", err)
	}
	return sessionID, nil
}

// AddUserToChatSession adds a user to a chat session
func (w *PostgresChatWriter) AddUserToChatSession(ctx context.Context, sessionID, userID int) error {
	query := `INSERT INTO chat_session_participants (session_id, user_id, joined_at) VALUES ($1, $2, NOW())`
	_, err := w.db.ExecContext(ctx, query, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to add user to chat session: %w", err)
	}

	// Update session's updated_at
	query = `UPDATE chat_sessions SET updated_at = NOW() WHERE id = $1`
	_, err = w.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update chat session timestamp: %w", err)
	}

	return nil
}

// RemoveUserFromChatSession removes a user from a chat session
func (w *PostgresChatWriter) RemoveUserFromChatSession(ctx context.Context, sessionID, userID int) error {
	query := `UPDATE chat_session_participants SET left_at = NOW() WHERE session_id = $1 AND user_id = $2 AND left_at IS NULL`
	result, err := w.db.ExecContext(ctx, query, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user from chat session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found in session or already left")
	}

	return nil
}

// CreateMessageIndex creates a new message index entry
func (w *PostgresChatWriter) CreateMessageIndex(ctx context.Context, sessionID, senderID int, messageType string, replyToMessageID *int) (int, error) {
	var messageID int
	query := `
		INSERT INTO message_index (session_id, sender_id, sent_at, updated_at, message_type, is_read, is_deleted, reply_to_message_id)
		VALUES ($1, $2, NOW(), NOW(), $3, false, false, $4)
		RETURNING id`
	err := w.db.QueryRowContext(ctx, query, sessionID, senderID, messageType, replyToMessageID).Scan(&messageID)
	if err != nil {
		return 0, fmt.Errorf("failed to create message index: %w", err)
	}
	return messageID, nil
}

// UpdateMessageReadStatus updates the read status of a message
func (w *PostgresChatWriter) UpdateMessageReadStatus(ctx context.Context, messageID int, isRead bool) error {
	query := `UPDATE message_index SET is_read = $1, updated_at = NOW() WHERE id = $2`
	_, err := w.db.ExecContext(ctx, query, isRead, messageID)
	if err != nil {
		return fmt.Errorf("failed to update message read status: %w", err)
	}
	return nil
}

// DeleteMessage soft deletes a message
func (w *PostgresChatWriter) DeleteMessage(ctx context.Context, messageID int) error {
	query := `UPDATE message_index SET is_deleted = true, updated_at = NOW() WHERE id = $1`
	result, err := w.db.ExecContext(ctx, query, messageID)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("message not found")
	}

	return nil
}

// StoreEncryptedMessage stores an encrypted message in the database
func (w *PostgresChatWriter) StoreEncryptedMessage(ctx context.Context, encryptedMsg *EncryptedMessage) error {
	query := `
		INSERT INTO encrypted_messages (message_id, encrypted_content, encryption_key_id, iv, auth_tag, content_type, file_size, checksum, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`
	_, err := w.db.ExecContext(ctx, query,
		encryptedMsg.MessageID,
		encryptedMsg.EncryptedContent,
		encryptedMsg.EncryptionKeyID,
		encryptedMsg.Iv,
		encryptedMsg.AuthTag,
		encryptedMsg.ContentType,
		encryptedMsg.FileSize,
		encryptedMsg.Checksum,
	)
	if err != nil {
		return fmt.Errorf("failed to store encrypted message: %w", err)
	}
	return nil
}

// MarkMessageAsReadByUser marks a message as read by a specific user
func (w *PostgresChatWriter) MarkMessageAsReadByUser(ctx context.Context, messageID, userID int) error {
	query := `INSERT INTO message_read_receipts (message_id, user_id, read_at) VALUES ($1, $2, NOW()) ON CONFLICT (message_id, user_id) DO NOTHING`
	_, err := w.db.ExecContext(ctx, query, messageID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark message as read: %w", err)
	}

	// Update message index
	query = `UPDATE message_index SET is_read = true, updated_at = NOW() WHERE id = $1`
	_, err = w.db.ExecContext(ctx, query, messageID)
	if err != nil {
		return fmt.Errorf("failed to update message read status: %w", err)
	}

	return nil
}

// CreateGroupChat creates a new group chat
func (w *PostgresChatWriter) CreateGroupChat(ctx context.Context, sessionID int, description, avatarURL *string, maxMembers int, privacyLevel string) error {
	query := `INSERT INTO group_chats (session_id, description, avatar_url, max_members, privacy_level, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`
	_, err := w.db.ExecContext(ctx, query, sessionID, description, avatarURL, maxMembers, privacyLevel)
	if err != nil {
		return fmt.Errorf("failed to create group chat: %w", err)
	}
	return nil
}

// UpdateGroupChat updates group chat information
func (w *PostgresChatWriter) UpdateGroupChat(ctx context.Context, sessionID int, description, avatarURL *string, maxMembers int, privacyLevel string) error {
	query := `UPDATE group_chats SET description = $1, avatar_url = $2, max_members = $3, privacy_level = $4, updated_at = NOW() WHERE session_id = $5`
	_, err := w.db.ExecContext(ctx, query, description, avatarURL, maxMembers, privacyLevel, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update group chat: %w", err)
	}
	return nil
}

// StoreUserKey stores a user's encryption key pair
// To preserve audit history, this method:
// 1. Deactivates the existing active key (if any)
// 2. Inserts a new key record
func (w *PostgresChatWriter) StoreUserKey(ctx context.Context, userKey *UserKey) error {
	// First deactivate existing active key to preserve audit history
	deactivateQuery := `UPDATE user_keys SET is_active = false, updated_at = NOW() WHERE user_id = $1 AND is_active = true`
	_, err := w.db.ExecContext(ctx, deactivateQuery, userKey.UserID)
	if err != nil {
		return fmt.Errorf("failed to deactivate existing user key: %w", err)
	}

	// Insert new key record
	query := `
		INSERT INTO user_keys (user_id, public_key, encrypted_private_key, key_algorithm, created_at, updated_at, is_active)
		VALUES ($1, $2, $3, $4, NOW(), NOW(), true)`
	_, err = w.db.ExecContext(ctx, query,
		userKey.UserID,
		userKey.PublicKey,
		userKey.EncryptedPrivateKey,
		userKey.KeyAlgorithm,
	)
	if err != nil {
		return fmt.Errorf("failed to store user key: %w", err)
	}
	return nil
}

// UpdateUserKey updates a user's encryption key
func (w *PostgresChatWriter) UpdateUserKey(ctx context.Context, userKey *UserKey) error {
	query := `UPDATE user_keys SET public_key = $1, encrypted_private_key = $2, key_algorithm = $3, updated_at = NOW() WHERE user_id = $4 AND is_active = true`
	result, err := w.db.ExecContext(ctx, query, userKey.PublicKey, userKey.EncryptedPrivateKey, userKey.KeyAlgorithm, userKey.UserID)
	if err != nil {
		return fmt.Errorf("failed to update user key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user key not found")
	}

	return nil
}

// DeactivateUserKey deactivates a user's encryption key
func (w *PostgresChatWriter) DeactivateUserKey(ctx context.Context, userID int) error {
	query := `UPDATE user_keys SET is_active = false, updated_at = NOW() WHERE user_id = $1 AND is_active = true`
	_, err := w.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to deactivate user key: %w", err)
	}
	return nil
}

// StoreOfflineEncryptedMessage stores an offline encrypted message
func (w *PostgresChatWriter) StoreOfflineEncryptedMessage(ctx context.Context, offlineMsg *OfflineEncryptedMessage) (int, error) {
	var messageID int
	query := `
		INSERT INTO offline_encrypted_messages (recipient_id, sender_id, session_id, encrypted_session_key, encrypted_content, iv, auth_tag, created_at, delivered_at, is_delivered)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NULL, false)
		RETURNING id`
	err := w.db.QueryRowContext(ctx, query,
		offlineMsg.RecipientID,
		offlineMsg.SenderID,
		offlineMsg.SessionID,
		offlineMsg.EncryptedSessionKey,
		offlineMsg.EncryptedContent,
		offlineMsg.Iv,
		offlineMsg.AuthTag,
	).Scan(&messageID)
	if err != nil {
		return 0, fmt.Errorf("failed to store offline encrypted message: %w", err)
	}
	return messageID, nil
}

// MarkOfflineMessageDelivered marks an offline message as delivered
func (w *PostgresChatWriter) MarkOfflineMessageDelivered(ctx context.Context, messageID int) error {
	query := `UPDATE offline_encrypted_messages SET delivered_at = NOW(), is_delivered = true WHERE id = $1 AND is_delivered = false`
	result, err := w.db.ExecContext(ctx, query, messageID)
	if err != nil {
		return fmt.Errorf("failed to mark offline message as delivered: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		log.WithFields(log.Fields{
			"message_id": messageID,
		}).Warn("Offline message not found or already delivered")
	}

	return nil
}

// DeleteOfflineMessage deletes an offline message
func (w *PostgresChatWriter) DeleteOfflineMessage(ctx context.Context, messageID int) error {
	query := `DELETE FROM offline_encrypted_messages WHERE id = $1`
	result, err := w.db.ExecContext(ctx, query, messageID)
	if err != nil {
		return fmt.Errorf("failed to delete offline message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("offline message not found")
	}

	return nil
}
