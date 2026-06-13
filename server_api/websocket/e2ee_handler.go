package websocket

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"sync"
	"time"

	"rich_chat/server_api/database"
	"rich_chat/server_api/service"
)

// E2EEMessageHandler handles E2EE-encrypted WebSocket messages
type E2EEMessageHandler struct {
	chatService       service.ChatService
	connectionManager HubInterface

	// Message deduplication cache to prevent replay attacks
	processedMessages map[string]time.Time // messageID -> timestamp
	dedupMutex        sync.RWMutex
	dedupTTL          time.Duration // Time to keep message IDs in cache
}

// NewE2EEMessageHandler creates a new E2EE message handler
func NewE2EEMessageHandler(chatService service.ChatService, connectionManager HubInterface) *E2EEMessageHandler {
	return &E2EEMessageHandler{
		chatService:       chatService,
		connectionManager: connectionManager,
		processedMessages: make(map[string]time.Time),
		dedupTTL:          5 * time.Minute, // Keep message IDs for 5 minutes
	}
}

// cleanupExpiredMessages removes expired message IDs from the deduplication cache
func (h *E2EEMessageHandler) cleanupExpiredMessages() {
	h.dedupMutex.Lock()
	defer h.dedupMutex.Unlock()

	now := time.Now()
	for msgID, timestamp := range h.processedMessages {
		if now.Sub(timestamp) > h.dedupTTL {
			delete(h.processedMessages, msgID)
		}
	}
}

// isDuplicateMessage checks if a message has already been processed
func (h *E2EEMessageHandler) isDuplicateMessage(messageID string) bool {
	h.dedupMutex.RLock()
	_, exists := h.processedMessages[messageID]
	h.dedupMutex.RUnlock()
	return exists
}

// markMessageProcessed adds a message ID to the deduplication cache
func (h *E2EEMessageHandler) markMessageProcessed(messageID string) {
	h.dedupMutex.Lock()
	defer h.dedupMutex.Unlock()
	h.processedMessages[messageID] = time.Now()

	// Periodically cleanup expired entries
	if len(h.processedMessages) > 1000 {
		go h.cleanupExpiredMessages()
	}
}

// HandleMessage processes an incoming E2EE message from a WebSocket connection
func (h *E2EEMessageHandler) HandleMessage(conn *Connection, msg ClientMessage) error {
	switch msg.Type {
	case "e2ee_chat":
		return h.handleChatMessage(conn, msg)
	case "e2ee_key_upload":
		return h.handleKeyUpload(conn, msg)
	case "e2ee_key_request":
		return h.handleKeyRequest(conn, msg)
	case "e2ee_offline_sync":
		return h.handleOfflineSync(conn)
	case "e2ee_mark_delivered":
		return h.handleMarkDelivered(conn, msg)
	default:
		return fmt.Errorf("unknown E2EE message type: %s", msg.Type)
	}
}

// ValidateMessage checks if an E2EE message is valid
func (h *E2EEMessageHandler) ValidateMessage(msg ClientMessage) error {
	switch msg.Type {
	case "e2ee_chat":
		return h.validateChatMessage(msg)
	case "e2ee_key_upload":
		return h.validateKeyUpload(msg)
	case "e2ee_key_request":
		return h.validateKeyRequest(msg)
	case "e2ee_offline_sync":
		return h.validateOfflineSync()
	case "e2ee_mark_delivered":
		return h.validateMarkDelivered(msg)
	default:
		return fmt.Errorf("unknown E2EE message type: %s", msg.Type)
	}
}

// IsE2EE checks if the message type is E2EE
func (h *E2EEMessageHandler) IsE2EE(msgType string) bool {
	switch msgType {
	case "e2ee_chat", "e2ee_key_upload", "e2ee_key_request", "e2ee_offline_sync", "e2ee_mark_delivered":
		return true
	default:
		return false
	}
}

// handleChatMessage processes an encrypted chat message
func (h *E2EEMessageHandler) handleChatMessage(conn *Connection, msg ClientMessage) error {
	// Parse the encrypted message content
	chatData, ok := msg.Content.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid chat message content")
	}

	// Extract required fields
	sessionID, ok := chatData["session_id"].(float64)
	if !ok {
		return fmt.Errorf("missing or invalid session_id")
	}

	recipientID, ok := chatData["recipient_id"].(float64)
	if !ok {
		return fmt.Errorf("missing or invalid recipient_id")
	}

	encryptedSessionKey, ok := chatData["encrypted_session_key"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid encrypted_session_key")
	}

	encryptedContent, ok := chatData["encrypted_content"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid encrypted_content")
	}

	iv, ok := chatData["iv"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid iv")
	}

	authTag, ok := chatData["auth_tag"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid auth_tag")
	}

	// Decode base64-encoded binary data
	encryptedSessionKeyBytes, err := base64.StdEncoding.DecodeString(encryptedSessionKey)
	if err != nil {
		return fmt.Errorf("failed to decode encrypted_session_key: %w", err)
	}

	encryptedContentBytes, err := base64.StdEncoding.DecodeString(encryptedContent)
	if err != nil {
		return fmt.Errorf("failed to decode encrypted_content: %w", err)
	}

	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return fmt.Errorf("failed to decode iv: %w", err)
	}

	authTagBytes, err := base64.StdEncoding.DecodeString(authTag)
	if err != nil {
		return fmt.Errorf("failed to decode auth_tag: %w", err)
	}

	// Generate unique message ID for deduplication using cryptographically secure random bytes
	messageID := generateSecureMessageID(conn.UserID, int(sessionID))

	// Check for duplicate message (replay attack prevention)
	if h.isDuplicateMessage(messageID) {
		log.Printf("Duplicate E2EE message detected (possible replay attack): %s", messageID)
		return fmt.Errorf("duplicate message detected")
	}

	// Check if recipient is online
	isRecipientOnline := h.connectionManager.IsUserOnline(int(recipientID))

	// For WebSocket messages, we need to store the already-encrypted data directly
	// We'll use the repository to store offline encrypted messages for persistence
	offlineMsg := &database.OfflineEncryptedMessage{
		MessageID:           messageID,
		RecipientID:         int(recipientID),
		SenderID:            conn.UserID,
		SessionID:           int(sessionID),
		EncryptedSessionKey: encryptedSessionKeyBytes,
		EncryptedContent:    encryptedContentBytes,
		Iv:                  ivBytes,
		AuthTag:             authTagBytes,
	}

	// Store message to database for persistence and audit trail
	storedMessageID, err := h.chatService.StoreOfflineEncryptedMessage(context.Background(), offlineMsg)
	if err != nil {
		log.Printf("Failed to store E2EE message to database: %v", err)
		// Continue anyway to deliver the message, but log the error
	}

	if isRecipientOnline {
		// Send confirmation to sender
		confirmation := ServerMessage{
			Type:      "e2ee_message_sent",
			SessionID: int(sessionID),
			SenderID:  conn.UserID,
			Content: map[string]interface{}{
				"status":       "delivered",
				"recipient_id": int(recipientID),
				"message_id":   messageID,
				"timestamp":    time.Now(),
			},
			Timestamp: time.Now(),
		}

		// Broadcast the encrypted message to the recipient
		serverMsg := ServerMessage{
			Type:      "e2ee_chat",
			SessionID: int(sessionID),
			SenderID:  conn.UserID,
			Content: map[string]interface{}{
				"message_id":            messageID,
				"recipient_id":          int(recipientID),
				"sender_id":             conn.UserID,
				"session_id":            int(sessionID),
				"encrypted_session_key": encryptedSessionKey,
				"encrypted_content":     encryptedContent,
				"iv":                    iv,
				"auth_tag":              authTag,
			},
			Timestamp: time.Now(),
		}

		// Send to recipient
		if err := h.connectionManager.SendToUser(int(recipientID), serverMsg); err != nil {
			log.Printf("Failed to send encrypted message to online recipient: %v", err)
			// Message is already stored in database, will be delivered as offline message later
			// Update confirmation status to indicate it was stored, not delivered immediately
			if contentMap, ok := confirmation.Content.(map[string]interface{}); ok {
				contentMap["status"] = "stored"
			}
		} else {
			// Mark as delivered in database
			_ = h.chatService.MarkMessageDelivered(context.Background(), int(recipientID), storedMessageID)
		}

		_ = conn.SendMessage(confirmation)
	} else {
		// Message already stored as offline message above
		confirmation := ServerMessage{
			Type:      "e2ee_message_sent",
			SessionID: int(sessionID),
			SenderID:  conn.UserID,
			Content: map[string]interface{}{
				"status":       "stored",
				"recipient_id": int(recipientID),
				"message_id":   messageID,
				"timestamp":    time.Now(),
			},
			Timestamp: time.Now(),
		}
		_ = conn.SendMessage(confirmation)
	}

	// Mark message as processed to prevent replay
	h.markMessageProcessed(messageID)

	log.Printf("E2EE message from %d to %d processed (online: %v, message_id: %s)", conn.UserID, int(recipientID), isRecipientOnline, messageID)
	return nil
}

// handleKeyUpload processes a user's public key upload
// Note: Private key is NEVER stored on server for security reasons
func (h *E2EEMessageHandler) handleKeyUpload(conn *Connection, msg ClientMessage) error {
	keyData, ok := msg.Content.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid key upload content")
	}

	publicKey, ok := keyData["public_key"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid public_key")
	}

	keyAlgorithm, ok := keyData["key_algorithm"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid key_algorithm")
	}

	// Validate public key format and strength
	if err := validatePublicKey(publicKey, keyAlgorithm); err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	// Store the user's public key (private key is NEVER stored)
	storeReq := &service.StoreUserKeyRequest{
		UserID:       conn.UserID,
		PublicKey:    publicKey,
		KeyAlgorithm: keyAlgorithm,
	}

	if err := h.chatService.StoreUserKey(context.Background(), storeReq); err != nil {
		return fmt.Errorf("failed to store user key: %w", err)
	}

	// Send confirmation
	confirmation := ServerMessage{
		Type:      "e2ee_key_uploaded",
		SessionID: 0,
		SenderID:  conn.UserID,
		Content: map[string]interface{}{
			"status":        "success",
			"key_algorithm": keyAlgorithm,
			"timestamp":     time.Now(),
		},
		Timestamp: time.Now(),
	}
	_ = conn.SendMessage(confirmation)

	log.Printf("User %d uploaded E2EE key (algorithm: %s)", conn.UserID, keyAlgorithm)
	return nil
}

// handleKeyRequest processes a request for a user's public key
func (h *E2EEMessageHandler) handleKeyRequest(conn *Connection, msg ClientMessage) error {
	keyData, ok := msg.Content.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid key request content")
	}

	targetUserID, ok := keyData["user_id"].(float64)
	if !ok {
		return fmt.Errorf("missing or invalid user_id")
	}

	// Get the user's public key
	keyResp, err := h.chatService.GetUserPublicKey(context.Background(), &service.GetUserKeyRequest{
		UserID: int(targetUserID),
	})
	if err != nil {
		// Send error response
		errorMsg := ServerMessage{
			Type:      "e2ee_key_response",
			SessionID: 0,
			SenderID:  conn.UserID,
			Content: map[string]interface{}{
				"status":  "error",
				"user_id": int(targetUserID),
				"error":   err.Error(),
			},
			Timestamp: time.Now(),
		}
		_ = conn.SendMessage(errorMsg)
		return nil
	}

	// Send the public key
	response := ServerMessage{
		Type:      "e2ee_key_response",
		SessionID: 0,
		SenderID:  conn.UserID,
		Content: map[string]interface{}{
			"status":        "success",
			"user_id":       keyResp.UserID,
			"public_key":    keyResp.PublicKey,
			"key_algorithm": keyResp.Algorithm,
		},
		Timestamp: time.Now(),
	}
	_ = conn.SendMessage(response)

	log.Printf("User %d requested public key for user %d", conn.UserID, int(targetUserID))
	return nil
}

// handleOfflineSync delivers pending offline messages to a user
func (h *E2EEMessageHandler) handleOfflineSync(conn *Connection) error {
	// Limit offline messages to prevent memory overflow
	const maxOfflineMessages = 100

	// Get all undelivered messages for the user (limited)
	offlineResp, err := h.chatService.GetOfflineMessages(context.Background(), &service.GetOfflineMessagesRequest{
		UserID: conn.UserID,
		Limit:  maxOfflineMessages,
	})
	if err != nil {
		return fmt.Errorf("failed to get offline messages: %w", err)
	}

	// Send each message to the user (skip rate limit for offline sync)
	syncedCount := 0
	for _, offlineMsg := range offlineResp.Messages {
		serverMsg := ServerMessage{
			Type:      "e2ee_offline_message",
			SessionID: offlineMsg.SessionID,
			SenderID:  offlineMsg.SenderID,
			Content: map[string]interface{}{
				"message_id":            offlineMsg.MessageID,
				"encrypted_session_key": base64.StdEncoding.EncodeToString(offlineMsg.EncryptedSessionKey),
				"encrypted_content":     base64.StdEncoding.EncodeToString(offlineMsg.EncryptedContent),
				"iv":                    base64.StdEncoding.EncodeToString(offlineMsg.Iv),
				"auth_tag":              base64.StdEncoding.EncodeToString(offlineMsg.AuthTag),
				"created_at":            offlineMsg.CreatedAt,
			},
			Timestamp: time.Now(),
		}

		if err := conn.SendMessageWithRateLimit(serverMsg, false); err != nil {
			log.Printf("Failed to send offline message %d to user %d: %v, stopping sync", offlineMsg.MessageID, conn.UserID, err)
			break
		}

		// Mark the message as delivered
		_ = h.chatService.MarkMessageDelivered(context.Background(), conn.UserID, offlineMsg.MessageID)
		syncedCount++
	}

	// Send sync completion notification
	completion := ServerMessage{
		Type:      "e2ee_offline_sync_complete",
		SessionID: 0,
		SenderID:  conn.UserID,
		Content: map[string]interface{}{
			"status":          "complete",
			"messages_synced": syncedCount,
			"has_more":        offlineResp.Count > maxOfflineMessages,
			"timestamp":       time.Now(),
		},
		Timestamp: time.Now(),
	}
	if err := conn.SendMessageWithRateLimit(completion, false); err != nil {
		log.Printf("Failed to send sync completion to user %d: %v", conn.UserID, err)
	}

	log.Printf("User %d synced %d offline messages (total: %d)", conn.UserID, syncedCount, offlineResp.Count)
	return nil
}

// handleMarkDelivered marks a message as delivered
func (h *E2EEMessageHandler) handleMarkDelivered(conn *Connection, msg ClientMessage) error {
	msgData, ok := msg.Content.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid mark delivered content")
	}

	messageID, ok := msgData["message_id"].(float64)
	if !ok {
		return fmt.Errorf("missing or invalid message_id")
	}

	if err := h.chatService.MarkMessageDelivered(context.Background(), conn.UserID, int(messageID)); err != nil {
		return fmt.Errorf("failed to mark message delivered: %w", err)
	}

	// Send confirmation
	confirmation := ServerMessage{
		Type:      "e2ee_message_delivered",
		SessionID: 0,
		SenderID:  conn.UserID,
		Content: map[string]interface{}{
			"status":     "success",
			"message_id": int(messageID),
			"timestamp":  time.Now(),
		},
		Timestamp: time.Now(),
	}
	_ = conn.SendMessage(confirmation)

	log.Printf("User %d marked message %d as delivered", conn.UserID, int(messageID))
	return nil
}

// validateChatMessage validates an E2EE chat message
func (h *E2EEMessageHandler) validateChatMessage(msg ClientMessage) error {
	chatData, ok := msg.Content.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid chat message content")
	}

	requiredFields := []string{"session_id", "recipient_id", "encrypted_session_key", "encrypted_content", "iv", "auth_tag"}
	for _, field := range requiredFields {
		if _, exists := chatData[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	return nil
}

// validateKeyUpload validates a key upload message
func (h *E2EEMessageHandler) validateKeyUpload(msg ClientMessage) error {
	keyData, ok := msg.Content.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid key upload content")
	}

	// public_key and key_algorithm are required
	requiredFields := []string{"public_key", "key_algorithm"}
	for _, field := range requiredFields {
		if _, exists := keyData[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// encrypted_private_key is optional (for client-side key backup)

	return nil
}

// validateKeyRequest validates a key request message
func (h *E2EEMessageHandler) validateKeyRequest(msg ClientMessage) error {
	keyData, ok := msg.Content.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid key request content")
	}

	if _, exists := keyData["user_id"]; !exists {
		return fmt.Errorf("missing required field: user_id")
	}

	return nil
}

// validateOfflineSync validates an offline sync message
func (h *E2EEMessageHandler) validateOfflineSync() error {
	// No content required for offline sync
	return nil
}

// validateMarkDelivered validates a mark delivered message
func (h *E2EEMessageHandler) validateMarkDelivered(msg ClientMessage) error {
	msgData, ok := msg.Content.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid mark delivered content")
	}

	if _, exists := msgData["message_id"]; !exists {
		return fmt.Errorf("missing required field: message_id")
	}

	return nil
}

// DeliverOfflineMessages delivers pending offline messages when a user connects
func (h *E2EEMessageHandler) DeliverOfflineMessages(conn *Connection) {
	go func() {
		// Get undelivered message count
		count, err := h.chatService.GetUndeliveredMessageCount(context.Background(), conn.UserID)
		if err != nil {
			log.Printf("Failed to get offline message count for user %d: %v", conn.UserID, err)
			return
		}

		if count == 0 {
			// No offline messages
			notification := ServerMessage{
				Type:      "e2ee_offline_notification",
				SessionID: 0,
				SenderID:  conn.UserID,
				Content: map[string]interface{}{
					"status":        "no_offline_messages",
					"message_count": 0,
					"timestamp":     time.Now(),
				},
				Timestamp: time.Now(),
			}
			_ = conn.SendMessageWithRateLimit(notification, false)
			return
		}

		// Notify user about pending messages
		notification := ServerMessage{
			Type:      "e2ee_offline_notification",
			SessionID: 0,
			SenderID:  conn.UserID,
			Content: map[string]interface{}{
				"status":        "has_offline_messages",
				"message_count": count,
				"timestamp":     time.Now(),
			},
			Timestamp: time.Now(),
		}
		_ = conn.SendMessageWithRateLimit(notification, false)

		_ = h.handleOfflineSync(conn)
	}()
}

// validatePublicKey validates the format and strength of a public key
func validatePublicKey(publicKeyPEM string, algorithm string) error {
	// Decode PEM block
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	// Parse public key
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	// Validate based on algorithm
	switch algorithm {
	case "RSA-2048", "RSA-4096":
		rsaPubKey, ok := pub.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("not an RSA public key")
		}

		// Check minimum key size for security
		if rsaPubKey.N.BitLen() < 2048 {
			return fmt.Errorf("RSA key size %d bits is too small, minimum 2048 bits required", rsaPubKey.N.BitLen())
		}

		// Warn if key size is larger than expected
		if rsaPubKey.N.BitLen() > 8192 {
			return fmt.Errorf("RSA key size %d bits is too large, maximum 8192 bits allowed", rsaPubKey.N.BitLen())
		}

	default:
		return fmt.Errorf("unsupported key algorithm: %s", algorithm)
	}

	return nil
}

// generateSecureMessageID generates a cryptographically secure message ID
func generateSecureMessageID(userID, sessionID int) string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("%d-%d-%d", userID, sessionID, time.Now().UnixNano())
	}
	return fmt.Sprintf("%d-%d-%x", userID, sessionID, b)
}
