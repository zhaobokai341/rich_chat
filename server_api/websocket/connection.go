package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"rich_chat/server_api/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
)

// sendError sends an error response to the client with non-blocking send
func (c *Connection) sendError(errorMsg string, msgType string) {
	response := ServerMessage{
		Type:      "error",
		SessionID: 0,
		SenderID:  c.UserID,
		Content: map[string]interface{}{
			"error": errorMsg,
			"type":  msgType,
		},
		Timestamp: time.Now(),
	}
	c.sendNonBlocking(response)
}

// sendResponse sends a success response to the client with non-blocking send
func (c *Connection) sendResponse(msgType string, sessionID int, content interface{}) {
	response := ServerMessage{
		Type:      msgType,
		SessionID: sessionID,
		Content:   content,
		Timestamp: time.Now(),
	}
	c.sendNonBlocking(response)
}

// sendNonBlocking sends a message without blocking if the channel is full
func (c *Connection) sendNonBlocking(msg ServerMessage) {
	if atomic.LoadInt32(&c.sendClosed) == 1 {
		return
	}

	messageJSON, err := json.Marshal(msg)
	if err != nil {
		log.Errorf("Failed to marshal message: %v", err)
		return
	}

	select {
	case c.send <- messageJSON:
	default:
		log.Warnf("Send channel full for user %d, dropping message", c.UserID)
	}
}

// Config holds the WebSocket configuration
type Config struct {
	WRITEWAIT             time.Duration // Time allowed to write a message to the peer
	PONGWAIT              time.Duration // Time allowed to read the next pong message from the peer
	PINGPERIOD            time.Duration // Send pings to peer with this period (should be less than PongWait)
	MAXMESSAGESIZE        int64         // Maximum message size allowed from peer
	SEND_CHANNEL_BUFFER   int           // Buffered channel size for outbound messages
	MAX_MESSAGES_PER_SEC  int           // Rate limit: maximum messages per second
	OFFLINE_MESSAGE_DELAY time.Duration // Delay before delivering offline messages
}

// DefaultConfig returns the default WebSocket configuration
func DefaultConfig() Config {
	return Config{
		WRITEWAIT:             10 * time.Second,
		PONGWAIT:              60 * time.Second,
		PINGPERIOD:            (60 * time.Second * 9) / 10, // 54 seconds
		MAXMESSAGESIZE:        5120,                        // 5KB
		SEND_CHANNEL_BUFFER:   256,
		MAX_MESSAGES_PER_SEC:  10,
		OFFLINE_MESSAGE_DELAY: 200 * time.Millisecond,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, you should restrict this to your frontend domains
		// Example: return r.Header.Get("Origin") == "https://yourdomain.com"
		return true
	},
}

// ClientMessage represents a message received from a client
type ClientMessage struct {
	Type      string      `json:"type"`       // message type: 'chat', 'typing', 'ping', etc.
	SessionID int         `json:"session_id"` // chat session ID
	Content   interface{} `json:"content"`    // message content
	Timestamp time.Time   `json:"timestamp"`
}

// ServerMessage represents a message sent to clients
type ServerMessage struct {
	Type      string      `json:"type"`       // message type
	SessionID int         `json:"session_id"` // chat session ID
	SenderID  int         `json:"sender_id"`  // ID of the sender
	Content   interface{} `json:"content"`    // message content
	Timestamp time.Time   `json:"timestamp"`
}

// Connection manages a WebSocket connection to a client
type Connection struct {
	ws             *websocket.Conn // The actual WebSocket connection
	send           chan []byte     // Buffered channel of outbound messages
	hub            HubInterface    // The Hub that handles this connection
	UserID         int             // The authenticated user ID
	ActiveSessions map[int]bool    // The user's active chat sessions
	config         Config          // WebSocket configuration
	messageHandler MessageHandler  // Message handler (E2EE or other)

	// Rate limiting
	maxMessagesPerSecond int        // Maximum messages per second
	lastMessageTime      time.Time  // Timestamp of the last message sent
	rateLimiterMutex     sync.Mutex // Mutex to protect rate limiter

	// Channel close protection
	sendClosed   int32        // Atomic flag to prevent double close
	sendClosedMu sync.RWMutex // Mutex to protect send channel access
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Connection) WritePump() {
	ticker := time.NewTicker(c.config.PINGPERIOD)
	defer func() {
		ticker.Stop()
		// Don't try to write close message if connection is already closed
		_ = c.ws.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			time.Now().Add(time.Second),
		)
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The hub closed the channel.
				return
			}

			_ = c.ws.SetWriteDeadline(time.Now().Add(c.config.WRITEWAIT))
			w, err := c.ws.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			// Write the message
			if _, err := w.Write(message); err != nil {
				return
			}

			// Add queued messages to the current websocket message
			n := len(c.send)
		batchLoop:
			for i := 0; i < n; i++ {
				select {
				case msg, ok := <-c.send:
					if !ok {
						// Channel was closed, stop writing
						_ = w.Close()
						return
					}
					_, _ = w.Write([]byte("\n"))
					_, _ = w.Write(msg)
				default:
					// No more messages, exit batch loop
					break batchLoop
				}
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			// Send ping to keep connection alive using WriteControl (safe for concurrent use)
			if err := c.ws.WriteControl(
				websocket.PingMessage,
				nil,
				time.Now().Add(c.config.WRITEWAIT),
			); err != nil {
				return
			}
		}
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Connection) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.ws.Close()
	}()

	c.ws.SetReadLimit(c.config.MAXMESSAGESIZE)
	_ = c.ws.SetReadDeadline(time.Now().Add(c.config.PONGWAIT))
	c.ws.SetPongHandler(
		func(string) error {
			_ = c.ws.SetReadDeadline(time.Now().Add(c.config.PONGWAIT))
			return nil
		},
	)

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
				websocket.CloseNormalClosure,
			) {
				log.Errorf("WebSocket error: %v", err)
			} else {
				log.Debugf("WebSocket connection closed for user %d: %v", c.UserID, err)
			}
			break
		}

		// Parse the incoming message
		var clientMsg ClientMessage
		if err := json.Unmarshal(message, &clientMsg); err != nil {
			log.Errorf("Error parsing message: %v", err)
			continue
		}

		// Check if this is an E2EE message
		if c.messageHandler != nil && c.messageHandler.IsE2EE(clientMsg.Type) {
			// Validate the E2EE message
			if err := c.messageHandler.ValidateMessage(clientMsg); err != nil {
				log.Errorf("Invalid E2EE message: %v", err)
				c.sendError(err.Error(), clientMsg.Type)
				continue
			}

			// Handle the E2EE message
			if err := c.messageHandler.HandleMessage(c, clientMsg); err != nil {
				log.Errorf("Error handling E2EE message: %v", err)
				c.sendError(err.Error(), clientMsg.Type)
			}
			continue
		}

		// Process the message based on its type
		switch clientMsg.Type {
		case "join_session":
			// User wants to join a chat session
			sessionID, ok := clientMsg.Content.(float64) // JSON numbers are float64
			if !ok {
				log.Error("Invalid session ID in join_session message")
				continue
			}

			sessionIDInt := int(sessionID)
			c.ActiveSessions[sessionIDInt] = true

			// Register the connection with the hub for this session
			c.hub.JoinSession(c, sessionIDInt)

			// Send confirmation back to the user
			c.sendResponse("session_joined", sessionIDInt, "Successfully joined session")

		case "leave_session":
			// User wants to leave a chat session
			sessionID, ok := clientMsg.Content.(float64)
			if !ok {
				log.Error("Invalid session ID in leave_session message")
				continue
			}

			sessionIDInt := int(sessionID)
			delete(c.ActiveSessions, sessionIDInt)

			// Unregister the connection from the hub for this session
			c.hub.LeaveSession(c, sessionIDInt)

			// Send confirmation back to the user
			c.sendResponse("session_left", sessionIDInt, "Successfully left session")

		case "chat":
			// Handle chat message
			chatData, ok := clientMsg.Content.(map[string]interface{})
			if !ok {
				log.Error("Invalid chat message content")
				continue
			}

			sessionID, ok := chatData["session_id"].(float64)
			if !ok {
				log.Error("Missing or invalid session_id in chat message")
				continue
			}

			content, ok := chatData["content"].(string)
			if !ok {
				log.Error("Missing or invalid content in chat message")
				continue
			}

			// Create server message to broadcast
			serverMsg := ServerMessage{
				Type:      "chat",
				SessionID: int(sessionID),
				SenderID:  c.UserID,
				Content: map[string]interface{}{
					"text": content,
				},
				Timestamp: time.Now(),
			}

			// Broadcast to all clients in the session
			c.hub.Broadcast(serverMsg, int(sessionID))

		default:
			// Unknown message type
			c.sendError("Unknown message type: "+clientMsg.Type, clientMsg.Type)
		}
	}
}

// UpgradeToWebSocket upgrades an HTTP connection to a WebSocket connection
// with JWT authentication
func UpgradeToWebSocket(
	c *gin.Context,
	userService service.UserService,
	chatService service.ChatService,
	hub HubInterface,
	jwtSecret string,
	wsConfig Config,
) {
	// Authenticate the user using JWT from Authorization header
	userID, err := authenticateUser(c, userService, jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("WebSocket upgrade error: %v", err)
		return
	}

	// Create E2EE handler
	e2eeHandler := NewE2EEMessageHandler(chatService, hub)

	connection := &Connection{
		ws:                   conn,
		send:                 make(chan []byte, wsConfig.SEND_CHANNEL_BUFFER), // Buffered channel
		hub:                  hub,
		UserID:               userID,
		ActiveSessions:       make(map[int]bool),
		config:               wsConfig,
		messageHandler:       e2eeHandler,
		maxMessagesPerSecond: wsConfig.MAX_MESSAGES_PER_SEC, // Rate limit from config
	}

	// Register the connection with the hub
	hub.Register(connection)

	// Start the write and read pumps
	go connection.WritePump()
	go connection.ReadPump()

	// Deliver offline messages after pumps are started
	go deliverOfflineMessages(connection, chatService, e2eeHandler)
}

// deliverOfflineMessages delivers pending offline messages when a user connects
func deliverOfflineMessages(conn *Connection, chatService service.ChatService, e2eeHandler MessageHandler) {
	// Wait for WritePump to be ready by checking connection state
	// This delay ensures the WritePump goroutine has started and is ready to receive messages
	time.Sleep(conn.config.OFFLINE_MESSAGE_DELAY)

	// Check if connection is still active before delivering offline messages
	if atomic.LoadInt32(&conn.sendClosed) == 1 {
		return
	}

	// Get undelivered message count
	count, err := chatService.GetUndeliveredMessageCount(context.Background(), conn.UserID)
	if err != nil {
		log.Errorf("Failed to get offline message count for user %d: %v", conn.UserID, err)
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
		if err := conn.SendMessageWithRateLimit(notification, false); err != nil {
			log.Debugf("Failed to send offline notification to user %d: %v", conn.UserID, err)
		}
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
	if err := conn.SendMessageWithRateLimit(notification, false); err != nil {
		log.Debugf("Failed to send offline notification to user %d: %v", conn.UserID, err)
		return
	}

	// Trigger offline sync
	syncMsg := ClientMessage{
		Type:      "e2ee_offline_sync",
		SessionID: 0,
		Content:   nil,
		Timestamp: time.Now(),
	}
	_ = e2eeHandler.HandleMessage(conn, syncMsg)
}

// Claims represents JWT claims structure (same as in main.go)
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// authenticateUser validates the JWT token and returns the user ID
func authenticateUser(
	c *gin.Context,
	userService service.UserService,
	jwtSecret string,
) (int, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		log.Warningf("[WebSocket Auth] No authorization header provided")
		return 0, fmt.Errorf("no authorization header")
	}

	log.Debugf("[WebSocket Auth] Received Authorization header: %s...", authHeader[:min(30, len(authHeader))])

	// Expect "Bearer <token>" format
	tokenString := ""
	if len(authHeader) >= 7 && strings.HasPrefix(authHeader, "Bearer ") {
		tokenString = strings.TrimSpace(authHeader[7:])
	} else {
		log.Warningf("[WebSocket Auth] Invalid authorization header format: %s", authHeader)
		return 0, fmt.Errorf("invalid authorization header format")
	}

	log.Debugf("[WebSocket Auth] Token extracted (length: %d): %s...", len(tokenString), tokenString[:min(30, len(tokenString))])

	// Parse and validate the JWT token (similar to safe_policy.go)
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method to prevent "alg: none" attack
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

	if err != nil {
		log.Warningf("[WebSocket Auth] Token parse error: %v", err)
		return 0, fmt.Errorf("token parse error: %w", err)
	}

	if !token.Valid {
		log.Warningf("[WebSocket Auth] Token is invalid (expired or malformed)")
		return 0, fmt.Errorf("invalid token")
	}

	log.Debugf("[WebSocket Auth] Token validated successfully, claims UserID: %d", claims.UserID)

	// Check if user exists using UserService
	exists, err := userService.CheckUserExists(claims.UserID)
	if err != nil {
		log.Warningf("[WebSocket Auth] Failed to check user existence for user %d: %v", claims.UserID, err)
		return 0, fmt.Errorf("failed to verify user")
	}
	if !exists {
		log.Warningf("[WebSocket Auth] User %d does not exist in database", claims.UserID)
		return 0, fmt.Errorf("user does not exist")
	}

	log.Debugf("[WebSocket Auth] Authentication successful for user %d", claims.UserID)
	return claims.UserID, nil
}

// safeCloseSend safely closes the send channel only once
func (c *Connection) safeCloseSend() {
	c.sendClosedMu.Lock()
	defer c.sendClosedMu.Unlock()

	if atomic.CompareAndSwapInt32(&c.sendClosed, 0, 1) {
		close(c.send)
	}
}

// SendMessage sends a message to the client
func (c *Connection) SendMessage(msg ServerMessage) error {
	return c.SendMessageWithRateLimit(msg, true)
}

// SendMessageWithRateLimit sends a message to the client with optional rate limiting
func (c *Connection) SendMessageWithRateLimit(msg ServerMessage, applyRateLimit bool) error {
	// Check if send channel is already closed
	if atomic.LoadInt32(&c.sendClosed) == 1 {
		return fmt.Errorf("connection is closing")
	}

	// Check rate limit (skip for system messages like offline sync)
	if applyRateLimit && !c.allowMessage() {
		log.Errorf("Rate limit exceeded for user %d, dropping message", c.UserID)
		return fmt.Errorf("rate limit exceeded")
	}

	messageJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Don't send messages larger than c.config.MaxMessageSize
	if len(messageJSON) > int(c.config.MAXMESSAGESIZE) {
		return bytes.ErrTooLarge
	}

	select {
	case c.send <- messageJSON:
		return nil
	default:
		// If the send channel is full, close the connection safely
		c.safeCloseSend()
		c.hub.Unregister(c)
		return fmt.Errorf("send channel full, connection closing")
	}
}

// allowMessage checks if a message is allowed to be sent based on rate limiting
func (c *Connection) allowMessage() bool {
	c.rateLimiterMutex.Lock()
	defer c.rateLimiterMutex.Unlock()

	if c.maxMessagesPerSecond <= 0 {
		// No rate limit configured, allow all messages
		return true
	}

	now := time.Now()
	minInterval := time.Second / time.Duration(c.maxMessagesPerSecond)

	// If this is the first message or enough time has passed since the last message
	if c.lastMessageTime.IsZero() || now.Sub(c.lastMessageTime) >= minInterval {
		c.lastMessageTime = now
		return true
	}

	return false
}
