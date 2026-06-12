package service

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"rich_chat/server_api/database"
)

// ChatServiceImpl implements ChatService
type ChatServiceImpl struct {
	chatRepo    database.ChatRepository
	userRepo    database.UserRepository
	e2eeService E2EEncryptionService
}

// NewChatService creates a new chat service
func NewChatService(
	chatRepo database.ChatRepository,
	userRepo database.UserRepository,
	e2eeService E2EEncryptionService,
) *ChatServiceImpl {
	return &ChatServiceImpl{
		chatRepo:    chatRepo,
		userRepo:    userRepo,
		e2eeService: e2eeService,
	}
}

// CreateSession creates a new chat session and adds members
func (s *ChatServiceImpl) CreateSession(ctx context.Context, req *CreateSessionRequest) (*CreateSessionResponse, error) {
	// Validate session type
	if req.SessionType != "direct" && req.SessionType != "group" {
		return nil, fmt.Errorf("invalid session type: %s", req.SessionType)
	}

	// Create the chat session
	sessionID, err := s.chatRepo.CreateChatSession(ctx, req.SessionType, req.Name, &req.CreatorID)
	if err != nil {
		log.WithFields(log.Fields{
			"creator_id":   req.CreatorID,
			"session_type": req.SessionType,
			"error":        err.Error(),
		}).Error("Failed to create chat session")
		return nil, fmt.Errorf("failed to create chat session: %w", err)
	}

	// Add creator to the session
	if err := s.chatRepo.AddUserToChatSession(ctx, sessionID, req.CreatorID); err != nil {
		log.WithFields(log.Fields{
			"session_id": sessionID,
			"user_id":    req.CreatorID,
			"error":      err.Error(),
		}).Error("Failed to add creator to session")
		return nil, fmt.Errorf("failed to add creator to session: %w", err)
	}

	// Add other members to the session
	for _, memberID := range req.MemberIDs {
		if memberID == req.CreatorID {
			continue // Skip creator, already added
		}
		if err := s.chatRepo.AddUserToChatSession(ctx, sessionID, memberID); err != nil {
			log.WithFields(log.Fields{
				"session_id": sessionID,
				"user_id":    memberID,
				"error":      err.Error(),
			}).Error("Failed to add member to session")
			return nil, fmt.Errorf("failed to add member %d to session: %w", memberID, err)
		}
	}

	// Get session name for response
	sessionName := ""
	if req.Name != nil {
		sessionName = *req.Name
	}

	log.WithFields(log.Fields{
		"session_id":   sessionID,
		"creator_id":   req.CreatorID,
		"session_type": req.SessionType,
		"member_count": len(req.MemberIDs) + 1,
	}).Info("Chat session created successfully")

	return &CreateSessionResponse{
		SessionID: sessionID,
		Name:      sessionName,
	}, nil
}

// JoinSession adds a user to an existing chat session
func (s *ChatServiceImpl) JoinSession(ctx context.Context, req *JoinSessionRequest) error {
	// Verify session exists
	_, err := s.chatRepo.GetChatSession(ctx, req.SessionID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id":    req.UserID,
			"session_id": req.SessionID,
			"error":      err.Error(),
		}).Warning("Attempted to join non-existent session")
		return fmt.Errorf("session not found: %w", err)
	}

	// Add user to session
	if err := s.chatRepo.AddUserToChatSession(ctx, req.SessionID, req.UserID); err != nil {
		log.WithFields(log.Fields{
			"user_id":    req.UserID,
			"session_id": req.SessionID,
			"error":      err.Error(),
		}).Error("Failed to join session")
		return fmt.Errorf("failed to join session: %w", err)
	}

	log.WithFields(log.Fields{
		"user_id":    req.UserID,
		"session_id": req.SessionID,
	}).Info("User joined session successfully")

	return nil
}

// LeaveSession removes a user from a chat session
func (s *ChatServiceImpl) LeaveSession(ctx context.Context, req *LeaveSessionRequest) error {
	// Remove user from session
	if err := s.chatRepo.RemoveUserFromChatSession(ctx, req.SessionID, req.UserID); err != nil {
		log.WithFields(log.Fields{
			"user_id":    req.UserID,
			"session_id": req.SessionID,
			"error":      err.Error(),
		}).Error("Failed to leave session")
		return fmt.Errorf("failed to leave session: %w", err)
	}

	log.WithFields(log.Fields{
		"user_id":    req.UserID,
		"session_id": req.SessionID,
	}).Info("User left session successfully")

	return nil
}

// GetUserSessions retrieves all sessions for a user
func (s *ChatServiceImpl) GetUserSessions(ctx context.Context, userID int) ([]*SessionInfo, error) {
	sessions, err := s.chatRepo.GetUserChatSessions(ctx, userID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to get user sessions")
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	var sessionInfos []*SessionInfo
	for _, session := range sessions {
		// Get member count for each session
		memberCount, err := s.chatRepo.GetUsersInChatSession(ctx, session.ID)
		if err != nil {
			log.WithFields(log.Fields{
				"session_id": session.ID,
				"error":      err.Error(),
			}).Warning("Failed to get member count for session")
			memberCount = []int{}
		}

		sessionInfo := &SessionInfo{
			SessionID:   session.ID,
			SessionType: session.SessionType,
			Name:        session.Name,
			CreatedAt:   session.CreatedAt,
			MemberCount: len(memberCount),
		}

		// For direct chats, find the partner info
		if session.SessionType == "direct" && len(memberCount) == 2 {
			for _, memberID := range memberCount {
				if memberID != userID {
					sessionInfo.PartnerID = memberID
					// Try to get partner's name
					userInfo, err := s.userRepo.GetUserBasicInfo(memberID)
					if err == nil && userInfo != nil {
						if userInfo.Nickname != "" {
							sessionInfo.PartnerName = userInfo.Nickname
						} else {
							sessionInfo.PartnerName = userInfo.Username
						}
					} else {
						sessionInfo.PartnerName = fmt.Sprintf("User %d", memberID)
					}
					break
				}
			}
		}

		sessionInfos = append(sessionInfos, sessionInfo)
	}

	return sessionInfos, nil
}

// StoreUserKey stores a user's public key (private key is NEVER stored)
func (s *ChatServiceImpl) StoreUserKey(ctx context.Context, req *StoreUserKeyRequest) error {
	// Validate input
	if req.PublicKey == "" {
		return fmt.Errorf("public key is required")
	}
	if req.KeyAlgorithm == "" {
		return fmt.Errorf("key algorithm is required")
	}

	// Check if user already has an active key, deactivate it first
	existingKey, err := s.chatRepo.GetUserKey(ctx, req.UserID)
	if err == nil && existingKey != nil && existingKey.IsActive {
		// Deactivate existing key
		if err := s.chatRepo.DeactivateUserKey(ctx, req.UserID); err != nil {
			log.WithFields(log.Fields{
				"user_id": req.UserID,
				"error":   err.Error(),
			}).Warning("Failed to deactivate existing user key")
			// Continue anyway, as the new key will be stored
		}
	}

	// Store the new key (private key is NEVER stored on server)
	userKey := &database.UserKey{
		UserID:              req.UserID,
		PublicKey:           req.PublicKey,
		EncryptedPrivateKey: nil, // Private key is never stored
		KeyAlgorithm:        req.KeyAlgorithm,
		IsActive:            true,
	}

	if err := s.chatRepo.StoreUserKey(ctx, userKey); err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Error("Failed to store user key")
		return fmt.Errorf("failed to store user key: %w", err)
	}

	log.WithFields(log.Fields{
		"user_id":       req.UserID,
		"key_algorithm": req.KeyAlgorithm,
	}).Info("User encryption key stored successfully")

	return nil
}

// GetUserPublicKey retrieves a user's public key for E2EE
func (s *ChatServiceImpl) GetUserPublicKey(ctx context.Context, req *GetUserKeyRequest) (*GetUserKeyResponse, error) {
	// Get the user's key
	userKey, err := s.chatRepo.GetUserKey(ctx, req.UserID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Warning("Failed to get user key")
		return nil, fmt.Errorf("failed to get user key: %w", err)
	}

	return &GetUserKeyResponse{
		UserID:    req.UserID,
		PublicKey: userKey.PublicKey,
		Algorithm: userKey.KeyAlgorithm,
	}, nil
}

// SendEncryptedMessage encrypts and sends a message to a recipient
func (s *ChatServiceImpl) SendEncryptedMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	// Validate input
	if len(req.PlaintextContent) == 0 {
		return nil, fmt.Errorf("message content is required")
	}
	if req.RecipientPubKey == "" {
		return nil, fmt.Errorf("recipient public key is required")
	}

	// Encrypt the message using E2EE service
	encryptedSessionKey, encryptedContent, iv, authTag, err := s.e2eeService.EncryptMessage(
		req.PlaintextContent,
		req.RecipientPubKey,
	)
	if err != nil {
		log.WithFields(log.Fields{
			"sender_id":    req.SenderID,
			"recipient_id": req.RecipientID,
			"error":        err.Error(),
		}).Error("Failed to encrypt message")
		return nil, fmt.Errorf("failed to encrypt message: %w", err)
	}

	// Check if recipient is online (this would be done via WebSocket service in practice)
	// For now, we'll store as offline message and let the WebSocket service handle delivery
	isOnline := false // This would be checked via WebSocketService.IsUserOnline()

	var messageID int

	if isOnline {
		// For online users, we could store in regular message index
		// For now, we'll still store as encrypted message for audit trail
		messageID, err = s.chatRepo.CreateMessageIndex(ctx, req.SessionID, req.SenderID, "encrypted", nil)
		if err != nil {
			log.WithFields(log.Fields{
				"session_id": req.SessionID,
				"sender_id":  req.SenderID,
				"error":      err.Error(),
			}).Error("Failed to create message index")
			return nil, fmt.Errorf("failed to create message index: %w", err)
		}

		// Store encrypted content
		encryptedMsg := &database.EncryptedMessage{
			MessageID:        messageID,
			EncryptedContent: string(encryptedContent),
			Iv:               iv,
			AuthTag:          authTag,
		}

		if err := s.chatRepo.StoreEncryptedMessage(ctx, encryptedMsg); err != nil {
			log.WithFields(log.Fields{
				"message_id": messageID,
				"error":      err.Error(),
			}).Error("Failed to store encrypted message")
			return nil, fmt.Errorf("failed to store encrypted message: %w", err)
		}
	} else {
		// Store as offline encrypted message for later delivery
		offlineMsg := &database.OfflineEncryptedMessage{
			RecipientID:         req.RecipientID,
			SenderID:            req.SenderID,
			SessionID:           req.SessionID,
			EncryptedSessionKey: encryptedSessionKey,
			EncryptedContent:    encryptedContent,
			Iv:                  iv,
			AuthTag:             authTag,
		}

		messageID, err = s.chatRepo.StoreOfflineEncryptedMessage(ctx, offlineMsg)
		if err != nil {
			log.WithFields(log.Fields{
				"sender_id":    req.SenderID,
				"recipient_id": req.RecipientID,
				"error":        err.Error(),
			}).Error("Failed to store offline encrypted message")
			return nil, fmt.Errorf("failed to store offline encrypted message: %w", err)
		}
	}

	log.WithFields(log.Fields{
		"message_id":   messageID,
		"sender_id":    req.SenderID,
		"recipient_id": req.RecipientID,
		"is_delivered": isOnline,
	}).Info("Encrypted message sent successfully")

	return &SendMessageResponse{
		MessageID:           messageID,
		EncryptedSessionKey: encryptedSessionKey,
		EncryptedContent:    encryptedContent,
		Iv:                  iv,
		AuthTag:             authTag,
		IsDelivered:         isOnline,
	}, nil
}

// GetOfflineMessages retrieves all undelivered encrypted messages for a user
func (s *ChatServiceImpl) GetOfflineMessages(ctx context.Context, req *GetOfflineMessagesRequest) (*GetOfflineMessagesResponse, error) {
	// Get undelivered messages
	messages, err := s.chatRepo.GetUndeliveredOfflineMessages(ctx, req.UserID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Error("Failed to get offline messages")
		return nil, fmt.Errorf("failed to get offline messages: %w", err)
	}

	var responseMessages []*OfflineMessageResponse
	for _, msg := range messages {
		responseMessages = append(responseMessages, &OfflineMessageResponse{
			MessageID:           msg.ID,
			SenderID:            msg.SenderID,
			SessionID:           msg.SessionID,
			EncryptedSessionKey: msg.EncryptedSessionKey,
			EncryptedContent:    msg.EncryptedContent,
			Iv:                  msg.Iv,
			AuthTag:             msg.AuthTag,
			CreatedAt:           msg.CreatedAt,
		})
	}

	log.WithFields(log.Fields{
		"user_id":       req.UserID,
		"message_count": len(responseMessages),
	}).Info("Retrieved offline messages successfully")

	return &GetOfflineMessagesResponse{
		Messages: responseMessages,
		Count:    len(responseMessages),
	}, nil
}

// MarkMessageDelivered marks an offline message as delivered
func (s *ChatServiceImpl) MarkMessageDelivered(ctx context.Context, userID, messageID int) error {
	// Verify the message belongs to the user
	messages, err := s.chatRepo.GetUndeliveredOfflineMessages(ctx, userID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id":    userID,
			"message_id": messageID,
			"error":      err.Error(),
		}).Error("Failed to verify message ownership")
		return fmt.Errorf("failed to verify message ownership: %w", err)
	}

	// Check if message belongs to user
	found := false
	for _, msg := range messages {
		if msg.ID == messageID && msg.RecipientID == userID {
			found = true
			break
		}
	}

	if !found {
		log.WithFields(log.Fields{
			"user_id":    userID,
			"message_id": messageID,
		}).Warning("Attempted to mark non-owned message as delivered")
		return fmt.Errorf("message not found or does not belong to user")
	}

	// Mark as delivered
	if err := s.chatRepo.MarkOfflineMessageDelivered(ctx, messageID); err != nil {
		log.WithFields(log.Fields{
			"user_id":    userID,
			"message_id": messageID,
			"error":      err.Error(),
		}).Error("Failed to mark message as delivered")
		return fmt.Errorf("failed to mark message as delivered: %w", err)
	}

	log.WithFields(log.Fields{
		"user_id":    userID,
		"message_id": messageID,
	}).Info("Message marked as delivered")

	return nil
}

// GetUndeliveredMessageCount returns the count of undelivered messages for a user
func (s *ChatServiceImpl) GetUndeliveredMessageCount(ctx context.Context, userID int) (int, error) {
	count, err := s.chatRepo.GetUndeliveredMessageCount(ctx, userID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to get undelivered message count")
		return 0, fmt.Errorf("failed to get undelivered message count: %w", err)
	}

	return count, nil
}

// StoreOfflineEncryptedMessage stores an encrypted message for offline delivery
func (s *ChatServiceImpl) StoreOfflineEncryptedMessage(ctx context.Context, offlineMsg *database.OfflineEncryptedMessage) (int, error) {
	messageID, err := s.chatRepo.StoreOfflineEncryptedMessage(ctx, offlineMsg)
	if err != nil {
		log.WithFields(log.Fields{
			"sender_id":    offlineMsg.SenderID,
			"recipient_id": offlineMsg.RecipientID,
			"session_id":   offlineMsg.SessionID,
			"error":        err.Error(),
		}).Error("Failed to store offline encrypted message")
		return 0, fmt.Errorf("failed to store offline encrypted message: %w", err)
	}

	log.WithFields(log.Fields{
		"message_id":   messageID,
		"sender_id":    offlineMsg.SenderID,
		"recipient_id": offlineMsg.RecipientID,
		"session_id":   offlineMsg.SessionID,
	}).Info("Offline encrypted message stored successfully")

	return messageID, nil
}
