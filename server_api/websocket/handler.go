package websocket

import (
	"net/http"

	"rich_chat/server_api/service"

	"github.com/gin-gonic/gin"
)

// Handler holds the dependencies for WebSocket handlers
type Handler struct {
	hub            HubInterface
	authService    service.AuthService
	userService    service.UserService
	chatService    service.ChatService
	messageHandler MessageHandler
	jwtSecret      string
	wsConfig       Config
}

// NewHandler creates a new WebSocket handler
func NewHandler(
	hub HubInterface,
	authService service.AuthService,
	userService service.UserService,
	chatService service.ChatService,
	jwtSecret string,
	wsConfig Config,
) *Handler {
	e2eeHandler := NewE2EEMessageHandler(chatService, hub)

	return &Handler{
		hub:            hub,
		authService:    authService,
		userService:    userService,
		chatService:    chatService,
		messageHandler: e2eeHandler,
		jwtSecret:      jwtSecret,
		wsConfig:       wsConfig,
	}
}

// WebSocketEndpoint handles the WebSocket connection upgrade with authentication
func (h *Handler) WebSocketEndpoint(c *gin.Context) {
	UpgradeToWebSocket(c, h.authService, h.userService, h.chatService, h.hub, h.jwtSecret, h.wsConfig)
}

// Implement the WebSocketService interface
func (h *Handler) HandleWebSocket(c *gin.Context) {
	h.WebSocketEndpoint(c)
}

func (h *Handler) BroadcastToSession(sessionID int, msg ServerMessage) {
	h.hub.BroadcastToSession(sessionID, msg)
}

func (h *Handler) SendToUser(userID int, msg ServerMessage) error {
	return h.hub.SendToUser(userID, msg)
}

func (h *Handler) IsUserOnline(userID int) bool {
	return h.hub.IsUserOnline(userID)
}

func (h *Handler) GetConnectionsForSession(sessionID int) []*Connection {
	return h.hub.GetConnectionsForSession(sessionID)
}

func (h *Handler) Close() error {
	h.hub.Close()
	return nil
}

// GetOnlineUsers returns a list of currently online users
func (h *Handler) GetOnlineUsers(c *gin.Context) {
	// This would typically require a method in the hub to return online users
	// For now, we'll return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"message": "Online users endpoint",
	})
}

// GetUserPresence returns the online status of a specific user
func (h *Handler) GetUserPresence(c *gin.Context) {
	userID := c.Param("user_id")

	// Convert string to int and check if user is online
	// This is a simplified version - in practice you'd want proper conversion and validation

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"online":  false, // Placeholder - would check hub in real implementation
	})
}
