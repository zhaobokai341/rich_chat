package database

import (
	"context"
)

// chatRepositoryAdapter combines ChatReader and ChatWriter to implement ChatRepository
type chatRepositoryAdapter struct {
	reader ChatReader
	writer ChatWriter
}

// NewChatRepositoryAdapter creates a ChatRepository from separate reader and writer
func NewChatRepositoryAdapter(reader ChatReader, writer ChatWriter) ChatRepository {
	return &chatRepositoryAdapter{
		reader: reader,
		writer: writer,
	}
}

// Read operations
func (a *chatRepositoryAdapter) GetChatSession(ctx context.Context, sessionID int) (*ChatSession, error) {
	return a.reader.GetChatSession(ctx, sessionID)
}

func (a *chatRepositoryAdapter) GetUsersInChatSession(ctx context.Context, sessionID int) ([]int, error) {
	return a.reader.GetUsersInChatSession(ctx, sessionID)
}

func (a *chatRepositoryAdapter) GetUserChatSessions(ctx context.Context, userID int) ([]*ChatSession, error) {
	return a.reader.GetUserChatSessions(ctx, userID)
}

func (a *chatRepositoryAdapter) GetMessageIndex(ctx context.Context, messageID int) (*MessageIndex, error) {
	return a.reader.GetMessageIndex(ctx, messageID)
}

func (a *chatRepositoryAdapter) GetEncryptedMessage(ctx context.Context, messageID int) (*EncryptedMessage, error) {
	return a.reader.GetEncryptedMessage(ctx, messageID)
}

func (a *chatRepositoryAdapter) GetMessagesForSession(ctx context.Context, sessionID int, limit, offset int) ([]*MessageIndex, error) {
	return a.reader.GetMessagesForSession(ctx, sessionID, limit, offset)
}

func (a *chatRepositoryAdapter) GetMessagesForUser(ctx context.Context, userID int, limit, offset int) ([]*MessageIndex, error) {
	return a.reader.GetMessagesForUser(ctx, userID, limit, offset)
}

func (a *chatRepositoryAdapter) GetUnreadMessagesForUser(ctx context.Context, userID int) ([]int, error) {
	return a.reader.GetUnreadMessagesForUser(ctx, userID)
}

func (a *chatRepositoryAdapter) GetReadReceiptsForMessage(ctx context.Context, messageID int) ([]*MessageReadReceipt, error) {
	return a.reader.GetReadReceiptsForMessage(ctx, messageID)
}

func (a *chatRepositoryAdapter) GetGroupChat(ctx context.Context, sessionID int) (*GroupChat, error) {
	return a.reader.GetGroupChat(ctx, sessionID)
}

func (a *chatRepositoryAdapter) GetUserKey(ctx context.Context, userID int) (*UserKey, error) {
	return a.reader.GetUserKey(ctx, userID)
}

func (a *chatRepositoryAdapter) GetUserPublicKey(ctx context.Context, userID int) (string, error) {
	return a.reader.GetUserPublicKey(ctx, userID)
}

func (a *chatRepositoryAdapter) GetUndeliveredOfflineMessages(ctx context.Context, recipientID int) ([]*OfflineEncryptedMessage, error) {
	return a.reader.GetUndeliveredOfflineMessages(ctx, recipientID)
}

func (a *chatRepositoryAdapter) GetOfflineMessageByID(ctx context.Context, messageID int) (*OfflineEncryptedMessage, error) {
	return a.reader.GetOfflineMessageByID(ctx, messageID)
}

func (a *chatRepositoryAdapter) GetUndeliveredMessageCount(ctx context.Context, recipientID int) (int, error) {
	return a.reader.GetUndeliveredMessageCount(ctx, recipientID)
}

// Write operations
func (a *chatRepositoryAdapter) CreateChatSession(ctx context.Context, sessionType string, name *string, createdBy *int) (int, error) {
	return a.writer.CreateChatSession(ctx, sessionType, name, createdBy)
}

func (a *chatRepositoryAdapter) AddUserToChatSession(ctx context.Context, sessionID, userID int) error {
	return a.writer.AddUserToChatSession(ctx, sessionID, userID)
}

func (a *chatRepositoryAdapter) RemoveUserFromChatSession(ctx context.Context, sessionID, userID int) error {
	return a.writer.RemoveUserFromChatSession(ctx, sessionID, userID)
}

func (a *chatRepositoryAdapter) CreateMessageIndex(ctx context.Context, sessionID, senderID int, messageType string, replyToMessageID *int) (int, error) {
	return a.writer.CreateMessageIndex(ctx, sessionID, senderID, messageType, replyToMessageID)
}

func (a *chatRepositoryAdapter) UpdateMessageReadStatus(ctx context.Context, messageID int, isRead bool) error {
	return a.writer.UpdateMessageReadStatus(ctx, messageID, isRead)
}

func (a *chatRepositoryAdapter) DeleteMessage(ctx context.Context, messageID int) error {
	return a.writer.DeleteMessage(ctx, messageID)
}

func (a *chatRepositoryAdapter) StoreEncryptedMessage(ctx context.Context, encryptedMsg *EncryptedMessage) error {
	return a.writer.StoreEncryptedMessage(ctx, encryptedMsg)
}

func (a *chatRepositoryAdapter) MarkMessageAsReadByUser(ctx context.Context, messageID, userID int) error {
	return a.writer.MarkMessageAsReadByUser(ctx, messageID, userID)
}

func (a *chatRepositoryAdapter) CreateGroupChat(ctx context.Context, sessionID int, description, avatarURL *string, maxMembers int, privacyLevel string) error {
	return a.writer.CreateGroupChat(ctx, sessionID, description, avatarURL, maxMembers, privacyLevel)
}

func (a *chatRepositoryAdapter) UpdateGroupChat(ctx context.Context, sessionID int, description, avatarURL *string, maxMembers int, privacyLevel string) error {
	return a.writer.UpdateGroupChat(ctx, sessionID, description, avatarURL, maxMembers, privacyLevel)
}

func (a *chatRepositoryAdapter) StoreUserKey(ctx context.Context, userKey *UserKey) error {
	return a.writer.StoreUserKey(ctx, userKey)
}

func (a *chatRepositoryAdapter) UpdateUserKey(ctx context.Context, userKey *UserKey) error {
	return a.writer.UpdateUserKey(ctx, userKey)
}

func (a *chatRepositoryAdapter) DeactivateUserKey(ctx context.Context, userID int) error {
	return a.writer.DeactivateUserKey(ctx, userID)
}

func (a *chatRepositoryAdapter) StoreOfflineEncryptedMessage(ctx context.Context, offlineMsg *OfflineEncryptedMessage) (int, error) {
	return a.writer.StoreOfflineEncryptedMessage(ctx, offlineMsg)
}

func (a *chatRepositoryAdapter) MarkOfflineMessageDelivered(ctx context.Context, messageID int) error {
	return a.writer.MarkOfflineMessageDelivered(ctx, messageID)
}

func (a *chatRepositoryAdapter) DeleteOfflineMessage(ctx context.Context, messageID int) error {
	return a.writer.DeleteOfflineMessage(ctx, messageID)
}
