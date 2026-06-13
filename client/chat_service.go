package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ChatMessage represents a chat message
type ChatMessage struct {
	MessageID           int       `json:"message_id"`
	SenderID            int       `json:"sender_id"`
	SessionID           int       `json:"session_id"`
	EncryptedSessionKey string    `json:"encrypted_session_key"`
	EncryptedContent    string    `json:"encrypted_content"`
	Iv                  string    `json:"iv"`
	AuthTag             string    `json:"auth_tag"`
	CreatedAt           time.Time `json:"created_at"`
}

// ChatSession represents a chat session
type ChatSession struct {
	SessionID   int        `json:"session_id"`
	SessionType string     `json:"session_type"`
	Name        *string    `json:"name"`
	CreatedAt   *time.Time `json:"created_at"`
	MemberCount int        `json:"member_count"`
	PartnerID   int        `json:"partner_id,omitempty"`
	PartnerName string     `json:"partner_name,omitempty"`
}

// ChatAPIClientInterface defines the interface for chat API operations
type ChatAPIClientInterface interface {
	GetUserBasicInfo(currentUserID int, token, verifyToken string, targetUserID int) (*UserBasicInfo, error)
	CreateChatSession(userID int, token, verifyToken string, recipientID int) (int, error)
	GetUserSessions(userID int, token, verifyToken string) ([]*ChatSession, error)
	GetUserPublicKey(currentUserID int, token, verifyToken string, targetUserID int) (string, string, error)
	StoreUserKey(userID int, token, publicKey string, encryptedPrivateKey []byte, keyAlgorithm string, verifyToken string) error
}

// ChatAPIClient handles chat-related API operations
type ChatAPIClient struct {
	client       *HTTPClient
	baseURL      string
	languagePack *LanguagePackWrapper
}

// Ensure ChatAPIClient implements ChatAPIClientInterface
var _ ChatAPIClientInterface = (*ChatAPIClient)(nil)

// NewChatAPIClient creates a new chat API client
func NewChatAPIClient(client *HTTPClient, baseURL string, languagePack *LanguagePackWrapper) *ChatAPIClient {
	return &ChatAPIClient{
		client:       client,
		baseURL:      baseURL,
		languagePack: languagePack,
	}
}

// GetUserBasicInfo retrieves basic user info
func (c *ChatAPIClient) GetUserBasicInfo(currentUserID int, token, verifyToken string, targetUserID int) (*UserBasicInfo, error) {
	resp, err := c.client.R().
		SetHeaders(map[string]string{
			"user_token": token,
			"user_id":    fmt.Sprintf("%d", currentUserID),
		}).
		SetQueryParams(map[string]string{
			"verify_token": verifyToken,
			"language":     getCurrentLanguage(),
		}).
		Get(fmt.Sprintf("%s/api/users/%d/basic", c.baseURL, targetUserID))

	if err != nil {
		return nil, fmt.Errorf("%s: %w", c.languagePack.Get("get_user_info_failed"), err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("%s: %d", c.languagePack.Get("http_status_code_error"), resp.StatusCode())
	}

	var result struct {
		User *UserBasicInfo `json:"user"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("%s", c.languagePack.Get("failed_to_parse_response"))
	}

	return result.User, nil
}

// CreateChatSession creates a new direct chat session
func (c *ChatAPIClient) CreateChatSession(userID int, token, verifyToken string, recipientID int) (int, error) {
	resp, err := c.client.R().
		SetHeaders(map[string]string{
			"user_token": token,
			"user_id":    fmt.Sprintf("%d", userID),
		}).
		SetHeader("Content-Type", "application/json").
		SetQueryParam("verify_token", verifyToken).
		SetBody(map[string]interface{}{
			"recipient_id": recipientID,
		}).
		Post(fmt.Sprintf("%s/api/users/%d/sessions", c.baseURL, userID))

	if err != nil {
		return 0, fmt.Errorf("%s: %w", c.languagePack.Get("create_session_failed"), err)
	}

	if resp.StatusCode() != 200 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(resp.Body(), &errorResp); err == nil {
			if errorResp.Message != "" {
				return 0, fmt.Errorf("%s", errorResp.Message)
			}
		}
		return 0, fmt.Errorf("%s: %d", c.languagePack.Get("http_status_code_error"), resp.StatusCode())
	}

	var result struct {
		SessionID int    `json:"session_id"`
		Message   string `json:"message"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return 0, fmt.Errorf("%s", c.languagePack.Get("failed_to_parse_response"))
	}

	return result.SessionID, nil
}

// GetUserSessions retrieves all chat sessions for a user
func (c *ChatAPIClient) GetUserSessions(userID int, token, verifyToken string) ([]*ChatSession, error) {
	resp, err := c.client.R().
		SetHeaders(map[string]string{
			"user_token": token,
			"user_id":    fmt.Sprintf("%d", userID),
		}).
		SetQueryParams(map[string]string{
			"verify_token": verifyToken,
			"language":     getCurrentLanguage(),
		}).
		Get(fmt.Sprintf("%s/api/users/%d/sessions", c.baseURL, userID))

	if err != nil {
		return nil, fmt.Errorf("%s: %w", c.languagePack.Get("get_sessions_failed"), err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("%s: %d", c.languagePack.Get("http_status_code_error"), resp.StatusCode())
	}

	var result struct {
		Sessions []*ChatSession `json:"sessions"`
		Count    int            `json:"count"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("%s", c.languagePack.Get("failed_to_parse_response"))
	}

	return result.Sessions, nil
}

// GetUserPublicKey retrieves a user's public encryption key
func (c *ChatAPIClient) GetUserPublicKey(currentUserID int, token, verifyToken string, targetUserID int) (string, string, error) {
	resp, err := c.client.R().
		SetHeaders(map[string]string{
			"user_token": token,
			"user_id":    fmt.Sprintf("%d", currentUserID),
		}).
		SetQueryParams(map[string]string{
			"verify_token": verifyToken,
			"language":     getCurrentLanguage(),
		}).
		Get(fmt.Sprintf("%s/api/users/%d/public-key", c.baseURL, targetUserID))

	if err != nil {
		return "", "", fmt.Errorf("%s: %w", c.languagePack.Get("get_public_key_failed"), err)
	}

	if resp.StatusCode() != 200 {
		return "", "", fmt.Errorf("%s: %d", c.languagePack.Get("http_status_code_error"), resp.StatusCode())
	}

	var result struct {
		UserID    int    `json:"user_id"`
		PublicKey string `json:"public_key"`
		Algorithm string `json:"algorithm"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return "", "", fmt.Errorf("%s", c.languagePack.Get("failed_to_parse_response"))
	}

	return result.PublicKey, result.Algorithm, nil
}

// StoreUserKey stores a user's encryption key
func (c *ChatAPIClient) StoreUserKey(userID int, token, publicKey string, encryptedPrivateKey []byte, keyAlgorithm string, verifyToken string) error {
	resp, err := c.client.R().
		SetHeaders(map[string]string{
			"user_token": token,
			"user_id":    fmt.Sprintf("%d", userID),
		}).
		SetHeader("Content-Type", "application/json").
		SetQueryParam("verify_token", verifyToken).
		SetBody(map[string]interface{}{
			"public_key":            publicKey,
			"encrypted_private_key": encryptedPrivateKey,
			"key_algorithm":         keyAlgorithm,
		}).
		Post(fmt.Sprintf("%s/api/users/%d/keys", c.baseURL, userID))

	if err != nil {
		return fmt.Errorf("%s: %w", c.languagePack.Get("store_key_failed"), err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("%s: %d", c.languagePack.Get("http_status_code_error"), resp.StatusCode())
	}

	return nil
}

// UserBasicInfo represents basic user info for chat
type UserBasicInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Bio      string `json:"bio"`
}

// WebSocketClient manages a WebSocket connection for chat
type WebSocketClient struct {
	conn          *websocket.Conn
	mu            sync.Mutex
	messages      chan *ChatMessage
	done          chan struct{}
	closeOnce     *sync.Once // Ensures done channel is closed only once
	isConnected   bool
	currentUserID int

	// Reconnection fields
	baseURL        string
	token          string
	reconnect      bool // Whether to auto-reconnect
	reconnectMu    sync.Mutex
	reconnecting   bool // Prevent concurrent reconnection attempts
	reconnectingMu sync.Mutex
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient() *WebSocketClient {
	return &WebSocketClient{
		messages:  make(chan *ChatMessage, 100),
		done:      make(chan struct{}),
		closeOnce: &sync.Once{},
	}
}

// Connect establishes a WebSocket connection
func (w *WebSocketClient) Connect(baseURL, token string, userID int) error {
	// Convert http/https to ws/wss
	var urlSchema string
	if len(baseURL) >= 7 && baseURL[:7] == "http://" {
		urlSchema = "ws"
	} else if len(baseURL) >= 8 && baseURL[:8] == "https://" {
		urlSchema = "wss"
	}

	wsURL := fmt.Sprintf("%s://%s:%d", urlSchema, URL_DOMAIN, URL_PORT)

	wsFullURL := fmt.Sprintf("%s/ws/chat", wsURL)

	// Log connection attempt details
	tokenPreview := ""
	if len(token) > 10 {
		tokenPreview = token[:10] + "..."
	} else {
		tokenPreview = token
	}
	log.Printf("[WebSocket] Attempting to connect to: %s", wsFullURL)
	log.Printf("[WebSocket] User ID: %d, Token preview: %s, Token length: %d", userID, tokenPreview, len(token))

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)
	header.Set("User-Agent", USER_AGENT)

	conn, resp, err := websocket.DefaultDialer.Dial(wsFullURL, header)
	if err != nil {
		log.Printf("[WebSocket] Connection failed!")
		log.Printf("[WebSocket] Error: %v", err)
		if resp != nil {
			log.Printf("[WebSocket] HTTP Status Code: %d", resp.StatusCode)
			log.Printf("[WebSocket] Response Headers: %v", resp.Header)
			// Try to read response body for more details
			if resp.Body != nil {
				body := make([]byte, 1024)
				n, _ := resp.Body.Read(body)
				if n > 0 {
					log.Printf("[WebSocket] Response Body: %s", string(body[:n]))
				}
			}
		}
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	log.Printf("[WebSocket] Successfully connected!")

	w.mu.Lock()
	w.conn = conn
	w.isConnected = true
	w.currentUserID = userID
	w.baseURL = baseURL
	w.token = token
	w.reconnect = true
	// Reset done channel for reconnection (create a new one if it was closed)
	w.done = make(chan struct{})
	// Reset closeOnce for the new connection
	w.closeOnce = &sync.Once{}
	w.mu.Unlock()

	// === Ping/Pong Setup ===
	// Design: Both client and server send pings to each other
	// - Client pings every 30s → Server responds with pong → Server's PongHandler resets server's ReadDeadline
	// - Server pings every 54s → Client responds with pong → Client's PongHandler resets client's ReadDeadline
	// WriteControl is thread-safe and can be called concurrently with NextWriter

	// PongHandler: called when we receive a pong from server (in response to our ping)
	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// PingHandler: called when we receive a ping from server
	// We must respond with pong AND reset read deadline
	conn.SetPingHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return conn.WriteControl(
			websocket.PongMessage,
			[]byte(appData),
			time.Now().Add(5*time.Second),
		)
	})

	// Start reading messages
	go w.readMessages()

	// Start ping sender (sends pings to server to keep server's ReadDeadline alive)
	go w.pingSender()

	return nil
}

// Close closes the WebSocket connection
func (w *WebSocketClient) Close() error {
	w.mu.Lock()

	// Disable reconnection first
	w.reconnect = false

	if w.conn != nil {
		w.isConnected = false
		w.mu.Unlock()
		// Use closeOnce to prevent panic from double close
		w.closeOnce.Do(func() {
			close(w.done)
		})
		return w.conn.Close()
	}
	w.mu.Unlock()
	return nil
}

// SendE2EEMessage sends an E2EE-encrypted message
func (w *WebSocketClient) SendE2EEMessage(sessionID, recipientID int, encryptedSessionKey, encryptedContent, iv, authTag string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.isConnected || w.conn == nil {
		return fmt.Errorf("not connected")
	}

	msg := map[string]interface{}{
		"type": "e2ee_chat",
		"content": map[string]interface{}{
			"session_id":            sessionID,
			"recipient_id":          recipientID,
			"encrypted_session_key": encryptedSessionKey,
			"encrypted_content":     encryptedContent,
			"iv":                    iv,
			"auth_tag":              authTag,
		},
		"timestamp": time.Now(),
	}

	return w.conn.WriteJSON(msg)
}

// SendKeyUpload uploads a user's encryption key
func (w *WebSocketClient) SendKeyUpload(publicKey string, encryptedPrivateKey []byte, keyAlgorithm string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.isConnected || w.conn == nil {
		return fmt.Errorf("not connected")
	}

	msg := map[string]interface{}{
		"type": "e2ee_key_upload",
		"content": map[string]interface{}{
			"public_key":            publicKey,
			"encrypted_private_key": base64.StdEncoding.EncodeToString(encryptedPrivateKey),
			"key_algorithm":         keyAlgorithm,
		},
		"timestamp": time.Now(),
	}

	return w.conn.WriteJSON(msg)
}

// SendKeyRequest requests another user's public key
func (w *WebSocketClient) SendKeyRequest(targetUserID int) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.isConnected || w.conn == nil {
		return fmt.Errorf("not connected")
	}

	msg := map[string]interface{}{
		"type": "e2ee_key_request",
		"content": map[string]interface{}{
			"user_id": targetUserID,
		},
		"timestamp": time.Now(),
	}

	return w.conn.WriteJSON(msg)
}

// SendOfflineSync requests offline message synchronization
func (w *WebSocketClient) SendOfflineSync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.isConnected || w.conn == nil {
		return fmt.Errorf("not connected")
	}

	msg := map[string]interface{}{
		"type":      "e2ee_offline_sync",
		"content":   nil,
		"timestamp": time.Now(),
	}

	return w.conn.WriteJSON(msg)
}

// pingSender periodically sends ping messages to keep the connection alive
func (w *WebSocketClient) pingSender() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.done:
			return
		case <-ticker.C:
			w.mu.Lock()
			conn := w.conn
			connected := w.isConnected
			w.mu.Unlock()

			if !connected || conn == nil {
				return
			}

			// WriteControl is thread-safe and can be called concurrently with NextWriter
			err := conn.WriteControl(
				websocket.PingMessage,
				nil,
				time.Now().Add(10*time.Second),
			)
			if err != nil {
				log.Printf("Failed to send ping: %v", err)
				return
			}
		}
	}
}

// readMessages reads messages from the WebSocket connection
func (w *WebSocketClient) readMessages() {
	defer func() {
		w.mu.Lock()
		w.isConnected = false
		w.mu.Unlock()
	}()

	// Set initial read deadline (must be longer than server ping interval)
	// Server sends ping every 54s, we need to respond within 60s
	w.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		_, message, err := w.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			} else {
				log.Printf("WebSocket closed: %v", err)
			}
			return
		}

		// Parse the message
		var wsMsg map[string]interface{}
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Failed to parse WebSocket message: %v", err)
			continue
		}

		// Handle different message types
		msgType, _ := wsMsg["type"].(string)
		switch msgType {
		case "e2ee_chat", "e2ee_offline_message":
			content, ok := wsMsg["content"].(map[string]interface{})
			if !ok {
				log.Printf("Invalid message content format")
				continue
			}

			// Safe type assertions to prevent panic
			senderID, ok := wsMsg["sender_id"].(float64)
			if !ok {
				log.Printf("Missing or invalid sender_id in message")
				continue
			}
			sessionID, ok := wsMsg["session_id"].(float64)
			if !ok {
				log.Printf("Missing or invalid session_id in message")
				continue
			}

			chatMsg := &ChatMessage{
				SenderID:  int(senderID),
				SessionID: int(sessionID),
			}

			if msgID, ok := content["message_id"].(float64); ok {
				chatMsg.MessageID = int(msgID)
			}
			if esk, ok := content["encrypted_session_key"].(string); ok {
				chatMsg.EncryptedSessionKey = esk
			} else if eskBytes, ok := content["encrypted_session_key"].([]interface{}); ok {
				chatMsg.EncryptedSessionKey = base64.StdEncoding.EncodeToString(interfaceToByteSlice(eskBytes))
			} else if eskBytes, ok := content["encrypted_session_key"].([]uint8); ok {
				chatMsg.EncryptedSessionKey = base64.StdEncoding.EncodeToString(eskBytes)
			}
			if ec, ok := content["encrypted_content"].(string); ok {
				chatMsg.EncryptedContent = ec
			} else if ecBytes, ok := content["encrypted_content"].([]interface{}); ok {
				chatMsg.EncryptedContent = base64.StdEncoding.EncodeToString(interfaceToByteSlice(ecBytes))
			} else if ecBytes, ok := content["encrypted_content"].([]uint8); ok {
				chatMsg.EncryptedContent = base64.StdEncoding.EncodeToString(ecBytes)
			}
			if iv, ok := content["iv"].(string); ok {
				chatMsg.Iv = iv
			} else if ivBytes, ok := content["iv"].([]interface{}); ok {
				chatMsg.Iv = base64.StdEncoding.EncodeToString(interfaceToByteSlice(ivBytes))
			} else if ivBytes, ok := content["iv"].([]uint8); ok {
				chatMsg.Iv = base64.StdEncoding.EncodeToString(ivBytes)
			}
			if at, ok := content["auth_tag"].(string); ok {
				chatMsg.AuthTag = at
			} else if atBytes, ok := content["auth_tag"].([]interface{}); ok {
				chatMsg.AuthTag = base64.StdEncoding.EncodeToString(interfaceToByteSlice(atBytes))
			} else if atBytes, ok := content["auth_tag"].([]uint8); ok {
				chatMsg.AuthTag = base64.StdEncoding.EncodeToString(atBytes)
			}

			// Use blocking send with timeout to prevent message loss
			select {
			case w.messages <- chatMsg:
				// Message sent successfully
			case <-time.After(5 * time.Second):
				log.Printf("Message channel full, dropping message")
			}

		case "e2ee_offline_notification":
			content, _ := wsMsg["content"].(map[string]interface{})
			if content != nil {
				count, _ := content["message_count"].(float64)
				status, _ := content["status"].(string)
				log.Printf("Offline notification: status=%s, count=%.0f", status, count)
			}

		case "e2ee_offline_sync_complete":
			content, _ := wsMsg["content"].(map[string]interface{})
			if content != nil {
				count, _ := content["messages_synced"].(float64)
				log.Printf("Offline sync complete: %d messages synced", int(count))
			}

		case "e2ee_message_sent":
			content, _ := wsMsg["content"].(map[string]interface{})
			if content != nil {
				status, _ := content["status"].(string)
				log.Printf("Message sent: status=%s", status)
			}

		case "e2ee_key_response":
			content, _ := wsMsg["content"].(map[string]interface{})
			if content != nil {
				status, _ := content["status"].(string)
				if status == "success" {
					log.Printf("Received public key for user %v", content["user_id"])
				}
			}

		case "error":
			content, _ := wsMsg["content"].(map[string]interface{})
			if content != nil {
				errMsg, _ := content["error"].(string)
				log.Printf("Server error: %s", errMsg)
			}
		}
	}
}

// GetMessages returns the message channel
func (w *WebSocketClient) GetMessages() <-chan *ChatMessage {
	return w.messages
}

// IsConnected returns whether the client is connected
func (w *WebSocketClient) IsConnected() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.isConnected
}

// reconnectMonitor monitors the connection and reconnects if needed
func (w *WebSocketClient) reconnectMonitor() {
	retryDelay := 1 * time.Second
	maxRetryDelay := 30 * time.Second

	for {
		select {
		case <-w.done:
			return
		default:
			time.Sleep(1 * time.Second)

			w.reconnectMu.Lock()
			shouldReconnect := w.reconnect
			w.reconnectMu.Unlock()

			if !shouldReconnect {
				return
			}

			w.mu.Lock()
			isConnected := w.isConnected
			w.mu.Unlock()

			if !isConnected && w.baseURL != "" && w.token != "" {
				// Check if already reconnecting
				w.reconnectingMu.Lock()
				if w.reconnecting {
					w.reconnectingMu.Unlock()
					continue // Another goroutine is handling reconnection
				}
				w.reconnecting = true
				w.reconnectingMu.Unlock()

				log.Printf("[WebSocket] Connection lost, attempting to reconnect in %v...", retryDelay)
				time.Sleep(retryDelay)

				// Increase retry delay with exponential backoff
				retryDelay = retryDelay * 2
				if retryDelay > maxRetryDelay {
					retryDelay = maxRetryDelay
				}

				// Attempt to reconnect
				err := w.Connect(w.baseURL, w.token, w.currentUserID)
				if err != nil {
					log.Printf("[WebSocket] Reconnection failed: %v", err)
				} else {
					log.Printf("[WebSocket] Reconnected successfully!")
					retryDelay = 1 * time.Second // Reset retry delay on success
				}

				// Release reconnecting lock
				w.reconnectingMu.Lock()
				w.reconnecting = false
				w.reconnectingMu.Unlock()
			} else if isConnected {
				retryDelay = 1 * time.Second // Reset retry delay when connected
			}
		}
	}
}

// SendPing sends a ping message
func (w *WebSocketClient) SendPing() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.isConnected || w.conn == nil {
		return fmt.Errorf("not connected")
	}

	return w.conn.WriteMessage(websocket.PingMessage, nil)
}

// interfaceToByteSlice converts []interface{} to []byte
func interfaceToByteSlice(arr []interface{}) []byte {
	result := make([]byte, len(arr))
	for i, v := range arr {
		if b, ok := v.(float64); ok {
			result[i] = byte(b)
		}
	}
	return result
}

// ChatService manages chat operations with encryption
type ChatService struct {
	apiClient    ChatAPIClientInterface
	wsClient     *WebSocketClient
	configMgr    ConfigManager
	languagePack *LanguagePackWrapper
	privateKey   *rsa.PrivateKey
	publicKeyPEM string
	mu           sync.Mutex
}

// NewChatService creates a new chat service
func NewChatService(apiClient ChatAPIClientInterface, configMgr ConfigManager, lp *LanguagePackWrapper) *ChatService {
	return &ChatService{
		apiClient:    apiClient,
		configMgr:    configMgr,
		languagePack: lp,
	}
}

// Connect connects to the WebSocket server
func (s *ChatService) Connect(token string, userID int) error {
	s.wsClient = NewWebSocketClient()
	err := s.wsClient.Connect(url_root, token, userID)
	if err != nil {
		return err
	}
	// Start reconnection monitor (only once during initial connection)
	go s.wsClient.reconnectMonitor()
	return nil
}

// Disconnect disconnects from the WebSocket server
func (s *ChatService) Disconnect() error {
	if s.wsClient != nil {
		return s.wsClient.Close()
	}
	return nil
}

// GetMessages returns the message channel
func (s *ChatService) GetMessages() <-chan *ChatMessage {
	if s.wsClient != nil {
		return s.wsClient.GetMessages()
	}
	return nil
}

// HasEncryptionKey checks if the user has encryption keys
func (s *ChatService) HasEncryptionKey(userID int) (bool, error) {
	return s.configMgr.HasEncryptionKey(userID)
}

// LoadEncryptionKey loads the user's encryption key from config
func (s *ChatService) LoadEncryptionKey(userID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	privateKeyPEM, err := s.configMgr.GetEncryptionKey(userID)
	if err != nil {
		return err
	}

	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	s.privateKey = key
	s.publicKeyPEM = privateKeyPEM
	return nil
}

// SaveEncryptionKey saves a private key to local storage
func (s *ChatService) SaveEncryptionKey(userID int, privateKeyPEM string) error {
	return s.configMgr.SaveEncryptionKey(userID, privateKeyPEM)
}

// UploadPublicKeyOnly uploads existing local public key to server
func (s *ChatService) UploadPublicKeyOnly(userID int, token string, verifyToken string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Load local private key
	privateKeyPEM, err := s.configMgr.GetEncryptionKey(userID)
	if err != nil {
		return fmt.Errorf("failed to load local private key: %w", err)
	}

	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	// Extract public key from private key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	// Upload public key to server
	err = s.apiClient.StoreUserKey(userID, token, string(publicKeyPEM), nil, "RSA-2048", verifyToken)
	if err != nil {
		return fmt.Errorf(s.languagePack.Get("key_upload_failed"), err)
	}

	// Store in memory
	s.privateKey = key
	s.publicKeyPEM = string(publicKeyPEM)

	return nil
}

// GenerateAndUploadKeys generates RSA key pair and uploads public key
func (s *ChatService) GenerateAndUploadKeys(userID int, token string, verifyToken string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate RSA-4096 key pair (consistent with server default)
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf(s.languagePack.Get("key_generation_failed"), err)
	}

	// Marshal private key to PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Marshal public key to PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf(s.languagePack.Get("key_marshal_failed"), err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	// Store private key locally
	err = s.configMgr.SaveEncryptionKey(userID, string(privateKeyPEM))
	if err != nil {
		return fmt.Errorf(s.languagePack.Get("key_save_failed"), err)
	}

	// Upload public key to server
	err = s.apiClient.StoreUserKey(userID, token, string(publicKeyPEM), nil, "RSA-4096", verifyToken)
	if err != nil {
		return fmt.Errorf(s.languagePack.Get("key_upload_failed"), err)
	}

	// Store in memory
	s.privateKey = privateKey
	s.publicKeyPEM = string(publicKeyPEM)

	return nil
}

// SendMessage encrypts and sends a chat message
func (s *ChatService) SendMessage(sessionID, recipientID int, message string, recipientPubKeyPEM string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.wsClient == nil || !s.wsClient.IsConnected() {
		return fmt.Errorf("not connected to chat server")
	}

	// Parse recipient's public key
	block, _ := pem.Decode([]byte(recipientPubKeyPEM))
	if block == nil {
		return fmt.Errorf("failed to decode recipient's public key")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse recipient's public key: %w", err)
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("not an RSA public key")
	}

	// Generate AES-256 session key
	sessionKey := make([]byte, 32)
	if _, err := rand.Read(sessionKey); err != nil {
		return fmt.Errorf("failed to generate session key: %w", err)
	}

	// Encrypt session key with recipient's RSA public key
	encryptedSessionKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPubKey, sessionKey, nil)
	if err != nil {
		return fmt.Errorf("failed to encrypt session key: %w", err)
	}

	// Encrypt message with AES-256-GCM
	blockCipher, err := aes.NewCipher(sessionKey)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and seal
	ciphertext := gcm.Seal(nil, nonce, []byte(message), nil)

	// Extract auth tag (last 16 bytes)
	authTag := ciphertext[len(ciphertext)-16:]
	encryptedContent := ciphertext[:len(ciphertext)-16]

	// Send via WebSocket
	err = s.wsClient.SendE2EEMessage(
		sessionID,
		recipientID,
		base64.StdEncoding.EncodeToString(encryptedSessionKey),
		base64.StdEncoding.EncodeToString(encryptedContent),
		base64.StdEncoding.EncodeToString(nonce),
		base64.StdEncoding.EncodeToString(authTag),
	)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// DecryptMessage decrypts an incoming chat message
func (s *ChatService) DecryptMessage(msg *ChatMessage) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.privateKey == nil {
		return "", fmt.Errorf("private key not loaded")
	}

	// Decode encrypted session key
	encryptedSessionKey, err := base64.StdEncoding.DecodeString(msg.EncryptedSessionKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted session key: %w", err)
	}

	// Decrypt session key with private key
	sessionKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, s.privateKey, encryptedSessionKey, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt session key: %w", err)
	}

	// Decode IV, encrypted content, and auth tag
	iv, err := base64.StdEncoding.DecodeString(msg.Iv)
	if err != nil {
		return "", fmt.Errorf("failed to decode IV: %w", err)
	}

	encryptedContent, err := base64.StdEncoding.DecodeString(msg.EncryptedContent)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted content: %w", err)
	}

	authTag, err := base64.StdEncoding.DecodeString(msg.AuthTag)
	if err != nil {
		return "", fmt.Errorf("failed to decode auth tag: %w", err)
	}

	// Reconstruct ciphertext (content + auth tag)
	ciphertext := append(encryptedContent, authTag...)

	// Create cipher
	blockCipher, err := aes.NewCipher(sessionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt message: %w", err)
	}

	return string(plaintext), nil
}
