package websocket

import (
	"rich_chat/server_api/service"

	"github.com/gin-gonic/gin"
)

// WebSocketService defines the interface for WebSocket operations
type WebSocketService interface {
	// HandleWebSocket handles the WebSocket connection upgrade with authentication
	HandleWebSocket(c *gin.Context)

	// BroadcastToSession sends a message to all connections in a specific session
	BroadcastToSession(sessionID int, msg ServerMessage)

	// SendToUser sends a message to a specific user if they are connected
	SendToUser(userID int, msg ServerMessage) error

	// IsUserOnline checks if a user is currently connected via WebSocket
	IsUserOnline(userID int) bool

	// GetConnectionsForSession returns all connections in a specific session
	GetConnectionsForSession(sessionID int) []*Connection

	// Close terminates the WebSocket service and all connections
	Close() error
}

// MessageHandler defines the interface for handling different types of WebSocket messages
type MessageHandler interface {
	// HandleMessage processes an incoming message from a WebSocket connection
	HandleMessage(conn *Connection, msg ClientMessage) error

	// ValidateMessage checks if a message is valid
	ValidateMessage(msg ClientMessage) error
}

// ConnectionManager defines the interface for managing WebSocket connections
type ConnectionManager interface {
	// GetConnectionsForSession returns all connections in a specific session
	GetConnectionsForSession(sessionID int) []*Connection

	// IsUserOnline checks if a user is currently connected via WebSocket
	IsUserOnline(userID int) bool

	// BroadcastToSession sends a message to all connections in a specific session
	BroadcastToSession(sessionID int, msg ServerMessage) error

	// SendToUser sends a message to a specific user if they are connected
	SendToUser(userID int, msg ServerMessage) error
}

// Authenticator defines the interface for WebSocket authentication
type Authenticator interface {
	// Authenticate validates a user's credentials and returns their user ID
	Authenticate(c *gin.Context, authService service.AuthService, userService service.UserService, jwtSecret string) (int, error)

	// ValidateToken checks if a token is valid
	ValidateToken(tokenString, jwtSecret string) (int, error)
}

// MessageValidator defines the interface for validating messages
type MessageValidator interface {
	// ValidateChatMessage checks if a chat message is valid
	ValidateChatMessage(content map[string]interface{}) error

	// ValidateJoinSessionMessage checks if a join session message is valid
	ValidateJoinSessionMessage(content interface{}) error

	// ValidateLeaveSessionMessage checks if a leave session message is valid
	ValidateLeaveSessionMessage(content interface{}) error
}
