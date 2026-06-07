package websocket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"rich_chat/server_api/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
)

// Config holds the WebSocket configuration
type Config struct {
	WRITEWAIT      time.Duration // Time allowed to write a message to the peer
	PONGWAIT       time.Duration // Time allowed to read the next pong message from the peer
	PINGPERIOD     time.Duration // Send pings to peer with this period (should be less than PongWait)
	MAXMESSAGESIZE int64         // Maximum message size allowed from peer
}

// DefaultConfig returns the default WebSocket configuration
func DefaultConfig() Config {
	return Config{
		WRITEWAIT:      10 * time.Second,
		PONGWAIT:       60 * time.Second,
		PINGPERIOD:     (60 * time.Second * 9) / 10, // 54 seconds
		MAXMESSAGESIZE: 5120,                        // 5KB
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: In production, you should restrict this to your frontend domains
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
	hub            *Hub            // The Hub that handles this connection
	UserID         int             // The authenticated user ID
	ActiveSessions map[int]bool    // The user's active chat sessions
	config         Config          // WebSocket configuration
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Connection) WritePump() {
	ticker := time.NewTicker(c.config.PINGPERIOD)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The hub closed the channel.
				err := c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					log.Printf("Error writing close message: %v", err)
				}
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
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte("\n"))
				_, _ = w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			// Send ping to keep connection alive
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Connection) ReadPump(
	authService service.AuthService,
	userService service.UserService,
) {
	defer func() {
		c.hub.Unregister <- c
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
			) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse the incoming message
		var clientMsg ClientMessage
		if err := json.Unmarshal(message, &clientMsg); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Process the message based on its type
		switch clientMsg.Type {
		case "join_session":
			// User wants to join a chat session
			sessionID, ok := clientMsg.Content.(float64) // JSON numbers are float64
			if !ok {
				log.Println("Invalid session ID in join_session message")
				continue
			}

			sessionIDInt := int(sessionID)
			c.ActiveSessions[sessionIDInt] = true

			// Register the connection with the hub for this session
			c.hub.JoinSession <- JoinSessionRequest{
				Connection: c,
				SessionID:  sessionIDInt,
			}

			// Send confirmation back to the user
			response := ServerMessage{
				Type:      "session_joined",
				SessionID: sessionIDInt,
				Content:   "Successfully joined session",
				Timestamp: time.Now(),
			}
			responseBytes, _ := json.Marshal(response)
			c.send <- responseBytes

		case "leave_session":
			// User wants to leave a chat session
			sessionID, ok := clientMsg.Content.(float64)
			if !ok {
				log.Println("Invalid session ID in leave_session message")
				continue
			}

			sessionIDInt := int(sessionID)
			delete(c.ActiveSessions, sessionIDInt)

			// Unregister the connection from the hub for this session
			c.hub.LeaveSession <- LeaveSessionRequest{
				Connection: c,
				SessionID:  sessionIDInt,
			}

			// Send confirmation back to the user
			response := ServerMessage{
				Type:      "session_left",
				SessionID: sessionIDInt,
				Content:   "Successfully left session",
				Timestamp: time.Now(),
			}
			responseBytes, _ := json.Marshal(response)
			c.send <- responseBytes

		case "chat":
			// Handle chat message
			chatData, ok := clientMsg.Content.(map[string]interface{})
			if !ok {
				log.Println("Invalid chat message content")
				continue
			}

			sessionID, ok := chatData["session_id"].(float64)
			if !ok {
				log.Println("Missing or invalid session_id in chat message")
				continue
			}

			content, ok := chatData["content"].(string)
			if !ok {
				log.Println("Missing or invalid content in chat message")
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
			c.hub.Broadcast <- BroadcastRequest{
				Message:   serverMsg,
				SessionID: int(sessionID),
			}

		default:
			// Unknown message type
			response := ServerMessage{
				Type:      "error",
				SessionID: 0,
				Content:   "Unknown message type: " + clientMsg.Type,
				Timestamp: time.Now(),
			}
			responseBytes, _ := json.Marshal(response)
			c.send <- responseBytes
		}
	}
}

// UpgradeToWebSocket upgrades an HTTP connection to a WebSocket connection
// with JWT authentication
func UpgradeToWebSocket(
	c *gin.Context,
	authService service.AuthService,
	userService service.UserService,
	hub *Hub,
	jwtSecret string,
	wsConfig Config,
) {
	// Authenticate the user using JWT from Authorization header
	userID, err := authenticateUser(c, authService, userService, jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	connection := &Connection{
		ws:             conn,
		send:           make(chan []byte, 256), // Buffered channel
		hub:            hub,
		UserID:         userID,
		ActiveSessions: make(map[int]bool),
		config:         wsConfig,
	}

	// Register the connection with the hub
	hub.Register <- connection

	// Start the write and read pumps
	go connection.WritePump()
	go connection.ReadPump(authService, userService)
}

// Claims represents JWT claims structure (same as in main.go)
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// authenticateUser validates the JWT token and returns the user ID
func authenticateUser(
	c *gin.Context,
	authService service.AuthService,
	userService service.UserService,
	jwtSecret string,
) (int, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return 0, fmt.Errorf("no authorization header")
	}

	// Expect "Bearer <token>" format
	tokenString := ""
	if len(authHeader) >= 7 && strings.HasPrefix(authHeader, "Bearer ") {
		tokenString = strings.TrimSpace(authHeader[7:])
	} else {
		return 0, fmt.Errorf("invalid authorization header format")
	}

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
		return 0, fmt.Errorf("token parse error: %w", err)
	}

	if !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	// Check if user exists using UserService
	exists, _ := userService.CheckUserExists(claims.UserID)
	if !exists {
		return 0, fmt.Errorf("user does not exist")
	}

	return claims.UserID, nil
}

// SendMessage sends a message to the client
func (c *Connection) SendMessage(msg ServerMessage) error {
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
	default:
		// If the send channel is full, close the connection
		close(c.send)
		c.hub.Unregister <- c
	}

	return nil
}
