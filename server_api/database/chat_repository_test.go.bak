package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgresChatRepository(t *testing.T) {
	t.Run("Interface Implementation", func(t *testing.T) {
		// Verify that PostgresChatRepository implements ChatRepository
		var _ ChatRepository = (*PostgresChatRepository)(nil)

		// Just verify the constructor exists and is accessible
		assert.NotNil(t, NewPostgresChatRepository)
	})
}

func TestChatModels(t *testing.T) {
	t.Run("Model Structures", func(t *testing.T) {
		// Test that our model structs are properly defined
		session := ChatSession{
			ID:          1,
			SessionType: "direct",
		}
		assert.Equal(t, 1, session.ID)
		assert.Equal(t, "direct", session.SessionType)

		messageIndex := MessageIndex{
			MessageType: "text",
		}
		assert.Equal(t, "text", messageIndex.MessageType)

		encryptedMsg := EncryptedMessage{
			EncryptedContent: "encrypted_data_here",
		}
		assert.Equal(t, "encrypted_data_here", encryptedMsg.EncryptedContent)

		readReceipt := MessageReadReceipt{
			MessageID: 1,
		}
		assert.Equal(t, 1, readReceipt.MessageID)
	})
}

func TestChatRepositoryInterfaceMethods(t *testing.T) {
	t.Run("Method Signatures", func(t *testing.T) {
		// This test ensures that all interface methods have the correct signatures
		// by attempting to assign them to function variables with the expected signatures

		// Define function variables with expected signatures
		var createChatSessionFunc func(context.Context, string, *string, *int) (int, error)
		var getUserChatSessionsFunc func(context.Context, int) ([]*ChatSession, error)
		var createMessageIndexFunc func(context.Context, int, int, string, *int) (int, error)
		var getMessagesForSessionFunc func(context.Context, int, int, int) ([]*MessageIndex, error)
		var storeEncryptedMessageFunc func(context.Context, *EncryptedMessage) error
		var markMessageAsReadByUserFunc func(context.Context, int, int) error

		// Assign the interface methods to these function variables
		// This verifies that the signatures match
		repo := &PostgresChatRepository{} // Using zero value for signature checking only

		// We're not calling these, just assigning to verify signatures match
		createChatSessionFunc = repo.CreateChatSession
		getUserChatSessionsFunc = repo.GetUserChatSessions
		createMessageIndexFunc = repo.CreateMessageIndex
		getMessagesForSessionFunc = repo.GetMessagesForSession
		storeEncryptedMessageFunc = repo.StoreEncryptedMessage
		markMessageAsReadByUserFunc = repo.MarkMessageAsReadByUser

		// Verify functions are assigned (this should always pass if signatures match)
		assert.NotNil(t, createChatSessionFunc)
		assert.NotNil(t, getUserChatSessionsFunc)
		assert.NotNil(t, createMessageIndexFunc)
		assert.NotNil(t, getMessagesForSessionFunc)
		assert.NotNil(t, storeEncryptedMessageFunc)
		assert.NotNil(t, markMessageAsReadByUserFunc)
	})
}

func TestE2EEModels(t *testing.T) {
	t.Run("UserKey Model", func(t *testing.T) {
		userKey := UserKey{
			ID:                  1,
			UserID:              42,
			PublicKey:           "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----",
			EncryptedPrivateKey: []byte("encrypted_private_key_bytes"),
			KeyAlgorithm:        "RSA-2048",
			IsActive:            true,
		}
		assert.Equal(t, 1, userKey.ID)
		assert.Equal(t, 42, userKey.UserID)
		assert.Equal(t, "RSA-2048", userKey.KeyAlgorithm)
		assert.True(t, userKey.IsActive)
		assert.NotEmpty(t, userKey.PublicKey)
		assert.NotEmpty(t, userKey.EncryptedPrivateKey)
	})

	t.Run("OfflineEncryptedMessage Model", func(t *testing.T) {
		offlineMsg := OfflineEncryptedMessage{
			ID:                  1,
			RecipientID:         42,
			SenderID:            100,
			SessionID:           5,
			EncryptedSessionKey: []byte("encrypted_session_key"),
			EncryptedContent:    []byte("encrypted_content"),
			IsDelivered:         false,
		}
		assert.Equal(t, 1, offlineMsg.ID)
		assert.Equal(t, 42, offlineMsg.RecipientID)
		assert.Equal(t, 100, offlineMsg.SenderID)
		assert.Equal(t, 5, offlineMsg.SessionID)
		assert.False(t, offlineMsg.IsDelivered)
		assert.NotEmpty(t, offlineMsg.EncryptedSessionKey)
		assert.NotEmpty(t, offlineMsg.EncryptedContent)
	})
}

func TestE2EEInterfaceMethods(t *testing.T) {
	t.Run("E2EE Key Management Method Signatures", func(t *testing.T) {
		repo := &PostgresChatRepository{}

		// Verify E2EE key management methods exist with correct signatures
		var storeUserKeyFunc func(context.Context, *UserKey) error
		var getUserKeyFunc func(context.Context, int) (*UserKey, error)
		var getUserPublicKeyFunc func(context.Context, int) (string, error)
		var updateUserKeyFunc func(context.Context, *UserKey) error
		var deactivateUserKeyFunc func(context.Context, int) error

		storeUserKeyFunc = repo.StoreUserKey
		getUserKeyFunc = repo.GetUserKey
		getUserPublicKeyFunc = repo.GetUserPublicKey
		updateUserKeyFunc = repo.UpdateUserKey
		deactivateUserKeyFunc = repo.DeactivateUserKey

		assert.NotNil(t, storeUserKeyFunc)
		assert.NotNil(t, getUserKeyFunc)
		assert.NotNil(t, getUserPublicKeyFunc)
		assert.NotNil(t, updateUserKeyFunc)
		assert.NotNil(t, deactivateUserKeyFunc)
	})

	t.Run("E2EE Offline Message Method Signatures", func(t *testing.T) {
		repo := &PostgresChatRepository{}

		// Verify E2EE offline message methods exist with correct signatures
		var storeOfflineMsgFunc func(context.Context, *OfflineEncryptedMessage) (int, error)
		var getUndeliveredMsgsFunc func(context.Context, int) ([]*OfflineEncryptedMessage, error)
		var markDeliveredFunc func(context.Context, int) error
		var deleteOfflineMsgFunc func(context.Context, int) error
		var getUndeliveredCountFunc func(context.Context, int) (int, error)

		storeOfflineMsgFunc = repo.StoreOfflineEncryptedMessage
		getUndeliveredMsgsFunc = repo.GetUndeliveredOfflineMessages
		markDeliveredFunc = repo.MarkOfflineMessageDelivered
		deleteOfflineMsgFunc = repo.DeleteOfflineMessage
		getUndeliveredCountFunc = repo.GetUndeliveredMessageCount

		assert.NotNil(t, storeOfflineMsgFunc)
		assert.NotNil(t, getUndeliveredMsgsFunc)
		assert.NotNil(t, markDeliveredFunc)
		assert.NotNil(t, deleteOfflineMsgFunc)
		assert.NotNil(t, getUndeliveredCountFunc)
	})
}

func TestChatRepositoryImplementsInterface(t *testing.T) {
	t.Run("Complete Interface Compliance", func(t *testing.T) {
		// This test verifies that PostgresChatRepository fully implements ChatRepository
		var repo ChatRepository = &PostgresChatRepository{}
		assert.NotNil(t, repo)

		// Verify all methods are accessible through the interface
		assert.NotNil(t, repo.CreateChatSession)
		assert.NotNil(t, repo.GetChatSession)
		assert.NotNil(t, repo.AddUserToChatSession)
		assert.NotNil(t, repo.RemoveUserFromChatSession)
		assert.NotNil(t, repo.GetUsersInChatSession)
		assert.NotNil(t, repo.GetUserChatSessions)
		assert.NotNil(t, repo.CreateMessageIndex)
		assert.NotNil(t, repo.GetMessageIndex)
		assert.NotNil(t, repo.UpdateMessageReadStatus)
		assert.NotNil(t, repo.DeleteMessage)
		assert.NotNil(t, repo.StoreEncryptedMessage)
		assert.NotNil(t, repo.GetEncryptedMessage)
		assert.NotNil(t, repo.GetMessagesForSession)
		assert.NotNil(t, repo.GetMessagesForUser)
		assert.NotNil(t, repo.MarkMessageAsReadByUser)
		assert.NotNil(t, repo.GetUnreadMessagesForUser)
		assert.NotNil(t, repo.GetReadReceiptsForMessage)
		assert.NotNil(t, repo.CreateGroupChat)
		assert.NotNil(t, repo.UpdateGroupChat)
		assert.NotNil(t, repo.GetGroupChat)
		assert.NotNil(t, repo.StoreUserKey)
		assert.NotNil(t, repo.GetUserKey)
		assert.NotNil(t, repo.GetUserPublicKey)
		assert.NotNil(t, repo.UpdateUserKey)
		assert.NotNil(t, repo.DeactivateUserKey)
		assert.NotNil(t, repo.StoreOfflineEncryptedMessage)
		assert.NotNil(t, repo.GetUndeliveredOfflineMessages)
		assert.NotNil(t, repo.MarkOfflineMessageDelivered)
		assert.NotNil(t, repo.DeleteOfflineMessage)
		assert.NotNil(t, repo.GetUndeliveredMessageCount)
	})
}
