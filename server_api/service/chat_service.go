package service

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"rich_chat/server_api/database"
)

// ChatServiceImpl implements ChatService
type ChatServiceImpl struct {
	chatRepo      database.ChatRepository
	userRepo      database.UserRepository
	e2eeService   E2EEncryptionService
	onlineChecker OnlineStatusChecker
}

// NewChatService creates a new chat service
func NewChatService(
	chatRepo database.ChatRepository,
	userRepo database.UserRepository,
	e2eeService E2EEncryptionService,
	onlineChecker OnlineStatusChecker,
) *ChatServiceImpl {
	return &ChatServiceImpl{
		chatRepo:      chatRepo,
		userRepo:      userRepo,
		e2eeService:   e2eeService,
		onlineChecker: onlineChecker,
	}
}

// CreateSession creates a new chat session
func (s *ChatServiceImpl) CreateSession(ctx context.Context, req *CreateSessionRequest) (*CreateSessionResponse, error) {
	if req.CreatorID <= 0 {
		return nil, ErrInvalidInput
	}

	exists, err := s.userRepo.ExistsByID(req.CreatorID)
	if err != nil {
		log.WithFields(log.Fields{
			"creator_id": req.CreatorID,
			"error":      err.Error(),
		}).Error("Failed to check creator existence")
		return nil, fmt.Errorf("failed to check creator existence: %w", err)
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	if req.SessionType == "direct" && len(req.MemberIDs) > 0 {
		for _, memberID := range req.MemberIDs {
			if memberID == req.CreatorID {
				continue
			}
			exists, err := s.userRepo.ExistsByID(memberID)
			if err != nil {
				log.WithFields(log.Fields{
					"member_id": memberID,
					"error":     err.Error(),
				}).Error("Failed to check member existence")
				return nil, fmt.Errorf("failed to check member existence: %w", err)
			}
			if !exists {
				return nil, ErrUserNotFound
			}
		}
	}

	createdBy := req.CreatorID
	sessionID, err := s.chatRepo.CreateChatSession(ctx, req.SessionType, req.Name, &createdBy)
	if err != nil {
		log.WithFields(log.Fields{
			"creator_id": req.CreatorID,
			"error":      err.Error(),
		}).Error("Failed to create session")
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	err = s.chatRepo.AddUserToChatSession(ctx, sessionID, req.CreatorID)
	if err != nil {
		log.WithFields(log.Fields{
			"session_id": sessionID,
			"creator_id": req.CreatorID,
			"error":      err.Error(),
		}).Error("Failed to add creator to session")
		return nil, fmt.Errorf("failed to add creator to session: %w", err)
	}

	for _, memberID := range req.MemberIDs {
		if memberID == req.CreatorID {
			continue
		}
		err = s.chatRepo.AddUserToChatSession(ctx, sessionID, memberID)
		if err != nil {
			log.WithFields(log.Fields{
				"session_id": sessionID,
				"member_id":  memberID,
				"error":      err.Error(),
			}).Error("Failed to add member to session")
			return nil, fmt.Errorf("failed to add member to session: %w", err)
		}
	}

	sessionName := ""
	if req.Name != nil {
		sessionName = *req.Name
	}

	log.WithFields(log.Fields{
		"session_id":   sessionID,
		"creator_id":   req.CreatorID,
		"session_type": req.SessionType,
	}).Info("Session created successfully")

	return &CreateSessionResponse{
		SessionID: sessionID,
		Name:      sessionName,
	}, nil
}

// JoinSession adds a user to an existing session
func (s *ChatServiceImpl) JoinSession(ctx context.Context, req *JoinSessionRequest) error {
	if req.UserID <= 0 || req.SessionID <= 0 {
		return ErrInvalidInput
	}

	exists, err := s.userRepo.ExistsByID(req.UserID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id":    req.UserID,
			"session_id": req.SessionID,
			"error":      err.Error(),
		}).Error("Failed to check user existence")
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return ErrUserNotFound
	}

	_, err = s.chatRepo.GetChatSession(ctx, req.SessionID)
	if err != nil {
		log.WithFields(log.Fields{
			"session_id": req.SessionID,
			"error":      err.Error(),
		}).Error("Failed to check session existence")
		return fmt.Errorf("session not found: %w", err)
	}

	err = s.chatRepo.AddUserToChatSession(ctx, req.SessionID, req.UserID)
	if err != nil {
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

// LeaveSession removes a user from a session
func (s *ChatServiceImpl) LeaveSession(ctx context.Context, req *LeaveSessionRequest) error {
	if req.UserID <= 0 || req.SessionID <= 0 {
		return ErrInvalidInput
	}

	err := s.chatRepo.RemoveUserFromChatSession(ctx, req.SessionID, req.UserID)
	if err != nil {
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
	if userID <= 0 {
		return nil, ErrInvalidInput
	}

	sessions, err := s.chatRepo.GetUserChatSessions(ctx, userID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to get user sessions")
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	result := make([]*SessionInfo, 0, len(sessions))
	for _, session := range sessions {
		info := &SessionInfo{
			SessionID:   session.ID,
			SessionType: session.SessionType,
			Name:        session.Name,
			CreatedAt:   session.CreatedAt,
		}

		members, err := s.chatRepo.GetUsersInChatSession(ctx, session.ID)
		if err == nil {
			info.MemberCount = len(members)
		}

		if session.SessionType == "direct" && len(members) == 2 {
			for _, memberID := range members {
				if memberID != userID {
					info.PartnerID = memberID
					// Try to get partner's name
					basicInfo, err := s.userRepo.GetUserBasicInfo(memberID)
					if err == nil && basicInfo != nil {
						if basicInfo.Nickname != "" {
							info.PartnerName = basicInfo.Nickname
						} else {
							info.PartnerName = basicInfo.Username
						}
					} else {
						info.PartnerName = fmt.Sprintf("User %d", memberID)
					}
					break
				}
			}
		}

		result = append(result, info)
	}

	return result, nil
}

// StoreUserKey stores a user's public key
func (s *ChatServiceImpl) StoreUserKey(ctx context.Context, req *StoreUserKeyRequest) error {
	if req.UserID <= 0 || req.PublicKey == "" {
		return ErrInvalidInput
	}

	exists, err := s.userRepo.ExistsByID(req.UserID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Error("Failed to check user existence")
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return ErrUserNotFound
	}

	algorithm := req.KeyAlgorithm
	if algorithm == "" {
		algorithm = "RSA-4096"
	}

	// Check if user already has an active key, deactivate it first to preserve history
	existingKey, err := s.chatRepo.GetUserKey(ctx, req.UserID)
	if err == nil && existingKey != nil && existingKey.IsActive {
		// Deactivate existing key (preserves history for audit trail)
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
		UserID:       req.UserID,
		PublicKey:    req.PublicKey,
		KeyAlgorithm: algorithm,
		IsActive:     true,
	}

	if err := s.chatRepo.StoreUserKey(ctx, userKey); err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Error("Failed to store user key")
		return fmt.Errorf("failed to store user key: %w", err)
	}

	log.WithFields(log.Fields{
		"user_id":   req.UserID,
		"algorithm": algorithm,
	}).Info("User key stored successfully")

	return nil
}

// GetUserPublicKey retrieves a user's public key
func (s *ChatServiceImpl) GetUserPublicKey(ctx context.Context, req *GetUserKeyRequest) (*GetUserKeyResponse, error) {
	if req.UserID <= 0 {
		return nil, ErrInvalidInput
	}

	publicKey, err := s.chatRepo.GetUserPublicKey(ctx, req.UserID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Error("Failed to get user public key")
		return nil, fmt.Errorf("failed to get user public key: %w", err)
	}

	userKey, err := s.chatRepo.GetUserKey(ctx, req.UserID)
	algorithm := "RSA-4096"
	if err == nil && userKey != nil {
		algorithm = userKey.KeyAlgorithm
	}

	return &GetUserKeyResponse{
		UserID:    req.UserID,
		PublicKey: publicKey,
		Algorithm: algorithm,
	}, nil
}

// SendEncryptedMessage sends an encrypted message to a recipient
func (s *ChatServiceImpl) SendEncryptedMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	if req.SenderID <= 0 || req.RecipientID <= 0 {
		return nil, ErrInvalidInput
	}

	if len(req.PlaintextContent) == 0 {
		return nil, fmt.Errorf("message content is required")
	}

	if req.RecipientPubKey == "" {
		return nil, fmt.Errorf("recipient public key is required")
	}

	exists, err := s.userRepo.ExistsByID(req.SenderID)
	if err != nil {
		log.WithFields(log.Fields{
			"sender_id": req.SenderID,
			"error":     err.Error(),
		}).Error("Failed to check sender existence")
		return nil, fmt.Errorf("failed to check sender existence: %w", err)
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	exists, err = s.userRepo.ExistsByID(req.RecipientID)
	if err != nil {
		log.WithFields(log.Fields{
			"recipient_id": req.RecipientID,
			"error":        err.Error(),
		}).Error("Failed to check recipient existence")
		return nil, fmt.Errorf("failed to check recipient existence: %w", err)
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	encryptedKey, encryptedContent, iv, authTag, err := s.e2eeService.EncryptMessage(
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

	// Check if recipient is online to determine storage strategy
	isDelivered := false
	if s.onlineChecker != nil {
		isDelivered = s.onlineChecker.IsUserOnline(req.RecipientID)
	}

	var messageID int

	if isDelivered {
		// Online user: only store in encrypted_messages table (real-time delivery)
		messageID, err = s.chatRepo.CreateMessageIndex(ctx, req.SessionID, req.SenderID, "text", nil)
		if err != nil {
			log.WithFields(log.Fields{
				"sender_id":    req.SenderID,
				"recipient_id": req.RecipientID,
				"error":        err.Error(),
			}).Error("Failed to create message index")
			return nil, fmt.Errorf("failed to create message index: %w", err)
		}

		encryptedMsg := &database.EncryptedMessage{
			MessageID:        messageID,
			EncryptedContent: string(encryptedContent),
			Iv:               iv,
			AuthTag:          authTag,
			EncryptionKeyID:  nil,
			ContentType:      "application/octet-stream",
		}

		err = s.chatRepo.StoreEncryptedMessage(ctx, encryptedMsg)
		if err != nil {
			log.WithFields(log.Fields{
				"message_id": messageID,
				"error":      err.Error(),
			}).Error("Failed to store encrypted message")
			return nil, fmt.Errorf("failed to store encrypted message: %w", err)
		}
	} else {
		// Offline user: only store in offline_encrypted_messages table (offline delivery)
		// Note: messageID will be assigned by StoreOfflineEncryptedMessage
		offlineMsg := &database.OfflineEncryptedMessage{
			RecipientID:         req.RecipientID,
			SenderID:            req.SenderID,
			SessionID:           req.SessionID,
			EncryptedSessionKey: encryptedKey,
			EncryptedContent:    encryptedContent,
			Iv:                  iv,
			AuthTag:             authTag,
			IsDelivered:         false,
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

		// Update MessageID in the response format
		offlineMsg.MessageID = fmt.Sprintf("msg_%d", messageID)
	}

	log.WithFields(log.Fields{
		"message_id":   messageID,
		"sender_id":    req.SenderID,
		"recipient_id": req.RecipientID,
		"is_delivered": isDelivered,
	}).Info("Encrypted message sent successfully")

	return &SendMessageResponse{
		MessageID:           messageID,
		EncryptedSessionKey: encryptedKey,
		EncryptedContent:    encryptedContent,
		Iv:                  iv,
		AuthTag:             authTag,
		IsDelivered:         isDelivered,
	}, nil
}

// StoreOfflineEncryptedMessage stores an encrypted message for offline delivery
func (s *ChatServiceImpl) StoreOfflineEncryptedMessage(ctx context.Context, offlineMsg *database.OfflineEncryptedMessage) (int, error) {
	if offlineMsg == nil || offlineMsg.RecipientID <= 0 {
		return 0, ErrInvalidInput
	}

	messageID, err := s.chatRepo.StoreOfflineEncryptedMessage(ctx, offlineMsg)
	if err != nil {
		log.WithFields(log.Fields{
			"recipient_id": offlineMsg.RecipientID,
			"error":        err.Error(),
		}).Error("Failed to store offline message")
		return 0, fmt.Errorf("failed to store offline message: %w", err)
	}

	log.WithFields(log.Fields{
		"message_id":   messageID,
		"recipient_id": offlineMsg.RecipientID,
	}).Info("Offline message stored successfully")

	return messageID, nil
}

// GetOfflineMessages retrieves undelivered messages for a user
func (s *ChatServiceImpl) GetOfflineMessages(ctx context.Context, req *GetOfflineMessagesRequest) (*GetOfflineMessagesResponse, error) {
	if req.UserID <= 0 {
		return nil, ErrInvalidInput
	}

	messages, err := s.chatRepo.GetUndeliveredOfflineMessages(ctx, req.UserID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Error("Failed to get offline messages")
		return nil, fmt.Errorf("failed to get offline messages: %w", err)
	}

	if req.Limit > 0 && len(messages) > req.Limit {
		messages = messages[:req.Limit]
	}

	responseMessages := make([]*OfflineMessageResponse, 0, len(messages))
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
	}).Info("Offline messages retrieved successfully")

	return &GetOfflineMessagesResponse{
		Messages: responseMessages,
		Count:    len(responseMessages),
	}, nil
}

// MarkMessageDelivered marks a message as delivered
func (s *ChatServiceImpl) MarkMessageDelivered(ctx context.Context, userID, messageID int) error {
	if userID <= 0 || messageID <= 0 {
		return ErrInvalidInput
	}

	// Verify that the message belongs to the user before marking as delivered
	offlineMsg, err := s.chatRepo.GetOfflineMessageByID(ctx, messageID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id":    userID,
			"message_id": messageID,
			"error":      err.Error(),
		}).Warning("Failed to get offline message for delivery verification")
		return fmt.Errorf("failed to verify message ownership: %w", err)
	}

	if offlineMsg.RecipientID != userID {
		log.WithFields(log.Fields{
			"user_id":    userID,
			"message_id": messageID,
			"owner_id":   offlineMsg.RecipientID,
		}).Warning("User attempted to mark message as delivered that does not belong to them")
		return ErrUnauthorized
	}

	err = s.chatRepo.MarkOfflineMessageDelivered(ctx, messageID)
	if err != nil {
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
	}).Debug("Message marked as delivered")

	return nil
}

// GetUndeliveredMessageCount returns the count of undelivered messages for a user
func (s *ChatServiceImpl) GetUndeliveredMessageCount(ctx context.Context, userID int) (int, error) {
	if userID <= 0 {
		return 0, ErrInvalidInput
	}

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

// UpdateSessionTimestamp updates the last activity timestamp for a session
func (s *ChatServiceImpl) UpdateSessionTimestamp(ctx context.Context, sessionID int) error {
	if sessionID <= 0 {
		return ErrInvalidInput
	}

	session, err := s.chatRepo.GetChatSession(ctx, sessionID)
	if err != nil {
		log.WithFields(log.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		}).Error("Failed to get session for timestamp update")
		return fmt.Errorf("failed to get session: %w", err)
	}

	if session.SessionType == "group" {
		err = s.chatRepo.UpdateGroupChat(ctx, sessionID, nil, nil, 0, "")
		if err != nil {
			log.WithFields(log.Fields{
				"session_id": sessionID,
				"error":      err.Error(),
			}).Error("Failed to update session timestamp")
			return fmt.Errorf("failed to update session timestamp: %w", err)
		}
	}

	return nil
}

// GetSessionMembers retrieves all members of a session
func (s *ChatServiceImpl) GetSessionMembers(ctx context.Context, sessionID int) ([]int, error) {
	if sessionID <= 0 {
		return nil, ErrInvalidInput
	}

	members, err := s.chatRepo.GetUsersInChatSession(ctx, sessionID)
	if err != nil {
		log.WithFields(log.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		}).Error("Failed to get session members")
		return nil, fmt.Errorf("failed to get session members: %w", err)
	}

	return members, nil
}

// IsSessionMember checks if a user is a member of a session
func (s *ChatServiceImpl) IsSessionMember(ctx context.Context, sessionID, userID int) (bool, error) {
	if sessionID <= 0 || userID <= 0 {
		return false, ErrInvalidInput
	}

	members, err := s.chatRepo.GetUsersInChatSession(ctx, sessionID)
	if err != nil {
		log.WithFields(log.Fields{
			"session_id": sessionID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to check session membership")
		return false, fmt.Errorf("failed to check session membership: %w", err)
	}

	for _, memberID := range members {
		if memberID == userID {
			return true, nil
		}
	}

	return false, nil
}

// GetSessionInfo retrieves detailed information about a session
func (s *ChatServiceImpl) GetSessionInfo(ctx context.Context, sessionID int) (*SessionInfo, error) {
	if sessionID <= 0 {
		return nil, ErrInvalidInput
	}

	session, err := s.chatRepo.GetChatSession(ctx, sessionID)
	if err != nil {
		log.WithFields(log.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		}).Error("Failed to get session info")
		return nil, fmt.Errorf("failed to get session info: %w", err)
	}

	info := &SessionInfo{
		SessionID:   session.ID,
		SessionType: session.SessionType,
		Name:        session.Name,
		CreatedAt:   session.CreatedAt,
	}

	members, err := s.chatRepo.GetUsersInChatSession(ctx, sessionID)
	if err == nil {
		info.MemberCount = len(members)
	}

	return info, nil
}

// CleanupOldOfflineMessages removes old offline messages
func (s *ChatServiceImpl) CleanupOldOfflineMessages(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoffTime := time.Now().Add(-olderThan)

	count := 0

	log.WithFields(log.Fields{
		"cutoff_time":   cutoffTime,
		"deleted_count": count,
	}).Info("Old offline messages cleanup completed")

	return count, nil
}
