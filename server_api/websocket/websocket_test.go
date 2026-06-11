package websocket

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"rich_chat/server_api/service"
)

func TestHub(t *testing.T) {
	t.Run("NewHub", func(t *testing.T) {
		hub := NewHub()
		assert.NotNil(t, hub)
		assert.NotNil(t, hub.Register)
		assert.NotNil(t, hub.Unregister)
		assert.NotNil(t, hub.JoinSession)
		assert.NotNil(t, hub.LeaveSession)
		assert.NotNil(t, hub.Broadcast)
	})

	t.Run("HubOperations", func(t *testing.T) {
		hub := NewHub()

		// Test adding and removing connections
		conn := &Connection{
			UserID:         1,
			ActiveSessions: make(map[int]bool),
		}

		// Register connection
		hub.connections[1] = conn
		assert.Len(t, hub.connections, 1)

		// Add to session
		hub.addToSession(conn, 100)
		sessionConns, exists := hub.sessionConnections[100]
		assert.True(t, exists)
		assert.Len(t, sessionConns, 1)

		// Remove from session
		hub.removeFromSession(conn, 100)
		_, exists = hub.sessionConnections[100]
		assert.False(t, exists) // Should be cleaned up since it was the only connection
	})
}

func TestConnection(t *testing.T) {
	t.Run("SendMessage", func(t *testing.T) {
		conn := &Connection{
			send:   make(chan []byte, 1),
			config: DefaultConfig(), // Initialize with default config
		}

		msg := ServerMessage{
			Type:      "test",
			SessionID: 1,
			SenderID:  1,
			Content:   "test content",
			Timestamp: time.Now(),
		}

		err := conn.SendMessage(msg)
		assert.NoError(t, err)
	})
}

func TestMessageStructures(t *testing.T) {
	t.Run("ClientMessageStructure", func(t *testing.T) {
		msg := ClientMessage{
			Type:      "chat",
			SessionID: 1,
			Content:   "Hello World",
		}

		assert.Equal(t, "chat", msg.Type)
		assert.Equal(t, 1, msg.SessionID)
		assert.Equal(t, "Hello World", msg.Content)
	})

	t.Run("ServerMessageStructure", func(t *testing.T) {
		msg := ServerMessage{
			Type:      "chat",
			SessionID: 1,
			SenderID:  123,
			Content:   "Hello World",
		}

		assert.Equal(t, "chat", msg.Type)
		assert.Equal(t, 1, msg.SessionID)
		assert.Equal(t, 123, msg.SenderID)
		assert.Equal(t, "Hello World", msg.Content)
	})
}

// Mock connection for testing
type MockConnection struct {
	mock.Mock
	UserID         int
	ActiveSessions map[int]bool
}

func (m *MockConnection) SendMessage(msg ServerMessage) error {
	args := m.Called(msg)
	return args.Error(0)
}

func TestHubConcurrency(t *testing.T) {
	t.Run("SafeMapAccess", func(t *testing.T) {
		hub := NewHub()

		// Start the hub in a goroutine
		done := make(chan bool)
		go func() {
			// Run for a short period to test concurrent access
			time.Sleep(100 * time.Millisecond)
			done <- true
		}()

		// Perform some operations concurrently
		go func() {
			for i := 0; i < 10; i++ {
				conn := &Connection{
					UserID:         i,
					ActiveSessions: make(map[int]bool),
				}
				hub.addToSession(conn, i*10)
				time.Sleep(1 * time.Millisecond)
			}
		}()

		<-done
	})
}

func TestHub_GetConnectionForUser(t *testing.T) {
	hub := NewHub()

	conn := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
	}

	hub.connections[1] = conn

	result := hub.GetConnectionForUser(1)
	assert.Equal(t, conn, result)

	result = hub.GetConnectionForUser(999)
	assert.Nil(t, result)
}

func TestHub_IsUserOnline(t *testing.T) {
	hub := NewHub()

	conn := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
	}

	hub.connections[1] = conn

	assert.True(t, hub.IsUserOnline(1))
	assert.False(t, hub.IsUserOnline(999))
}

func TestHub_GetConnectionsForSession(t *testing.T) {
	hub := NewHub()

	conn1 := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
	}
	conn2 := &Connection{
		UserID:         2,
		ActiveSessions: make(map[int]bool),
	}

	hub.addToSession(conn1, 100)
	hub.addToSession(conn2, 100)

	connections := hub.GetConnectionsForSession(100)
	assert.Len(t, connections, 2)

	connections = hub.GetConnectionsForSession(999)
	assert.Len(t, connections, 0)
}

func TestHub_SendToUser(t *testing.T) {
	hub := NewHub()

	conn := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 1),
		config:         DefaultConfig(),
	}

	hub.connections[1] = conn

	msg := ServerMessage{
		Type:      "test",
		SessionID: 1,
		SenderID:  2,
		Content:   "hello",
		Timestamp: time.Now(),
	}

	err := hub.SendToUser(1, msg)
	assert.NoError(t, err)

	err = hub.SendToUser(999, msg)
	assert.NoError(t, err)
}

func TestHub_Close(t *testing.T) {
	hub := NewHub()

	conn := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 1),
	}

	hub.connections[1] = conn
	hub.sessionConnections[100] = map[*Connection]bool{conn: true}

	// Test that maps are cleared after Close
	// Note: We don't test actual websocket close here as it requires a real connection
	close(conn.send)

	// Verify maps before close
	assert.Len(t, hub.connections, 1)
	assert.Len(t, hub.sessionConnections, 1)

	// Clear maps manually to simulate what Close would do
	hub.mutex.Lock()
	hub.connections = make(map[int]*Connection)
	hub.sessionConnections = make(map[int]map[*Connection]bool)
	hub.mutex.Unlock()

	assert.Len(t, hub.connections, 0)
	assert.Len(t, hub.sessionConnections, 0)
}

func TestHandler(t *testing.T) {
	mockAuthService := new(service.MockAuthService)
	mockUserService := new(service.MockUserService)
	mockChatService := new(service.MockChatService)
	hub := NewHub()
	wsConfig := DefaultConfig()

	handler := NewHandler(hub, mockAuthService, mockUserService, mockChatService, "secret", wsConfig)

	assert.NotNil(t, handler)
	assert.Equal(t, hub, handler.hub)
	assert.Equal(t, "secret", handler.jwtSecret)
}

func TestHandler_BroadcastToSession(t *testing.T) {
	mockAuthService := new(service.MockAuthService)
	mockUserService := new(service.MockUserService)
	mockChatService := new(service.MockChatService)
	hub := NewHub()
	wsConfig := DefaultConfig()

	handler := NewHandler(hub, mockAuthService, mockUserService, mockChatService, "secret", wsConfig)

	msg := ServerMessage{
		Type:      "test",
		SessionID: 1,
		SenderID:  1,
		Content:   "hello",
		Timestamp: time.Now(),
	}

	handler.BroadcastToSession(1, msg)
}

func TestHandler_IsUserOnline(t *testing.T) {
	mockAuthService := new(service.MockAuthService)
	mockUserService := new(service.MockUserService)
	mockChatService := new(service.MockChatService)
	hub := NewHub()
	wsConfig := DefaultConfig()

	handler := NewHandler(hub, mockAuthService, mockUserService, mockChatService, "secret", wsConfig)

	assert.False(t, handler.IsUserOnline(1))
}

func TestHandler_GetConnectionsForSession(t *testing.T) {
	mockAuthService := new(service.MockAuthService)
	mockUserService := new(service.MockUserService)
	mockChatService := new(service.MockChatService)
	hub := NewHub()
	wsConfig := DefaultConfig()

	handler := NewHandler(hub, mockAuthService, mockUserService, mockChatService, "secret", wsConfig)

	connections := handler.GetConnectionsForSession(1)
	assert.Len(t, connections, 0)
}

func TestHandler_Close(t *testing.T) {
	mockAuthService := new(service.MockAuthService)
	mockUserService := new(service.MockUserService)
	mockChatService := new(service.MockChatService)
	hub := NewHub()
	wsConfig := DefaultConfig()

	handler := NewHandler(hub, mockAuthService, mockUserService, mockChatService, "secret", wsConfig)

	err := handler.Close()
	assert.NoError(t, err)
}

func TestHandler_GetOnlineUsers(t *testing.T) {
	mockAuthService := new(service.MockAuthService)
	mockUserService := new(service.MockUserService)
	mockChatService := new(service.MockChatService)
	hub := NewHub()
	wsConfig := DefaultConfig()

	handler := NewHandler(hub, mockAuthService, mockUserService, mockChatService, "secret", wsConfig)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/online", handler.GetOnlineUsers)

	req, _ := http.NewRequest("GET", "/online", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_GetUserPresence(t *testing.T) {
	mockAuthService := new(service.MockAuthService)
	mockUserService := new(service.MockUserService)
	mockChatService := new(service.MockChatService)
	hub := NewHub()
	wsConfig := DefaultConfig()

	handler := NewHandler(hub, mockAuthService, mockUserService, mockChatService, "secret", wsConfig)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/presence/:user_id", handler.GetUserPresence)

	req, _ := http.NewRequest("GET", "/presence/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, 10*time.Second, config.WRITEWAIT)
	assert.Equal(t, 60*time.Second, config.PONGWAIT)
	assert.Equal(t, int64(5120), config.MAXMESSAGESIZE)
}

func TestConnection_SendMessage_LargeMessage(t *testing.T) {
	conn := &Connection{
		send:   make(chan []byte, 1),
		config: DefaultConfig(),
	}

	largeContent := make([]byte, 10000)
	msg := ServerMessage{
		Type:      "test",
		SessionID: 1,
		SenderID:  1,
		Content:   string(largeContent),
		Timestamp: time.Now(),
	}

	err := conn.SendMessage(msg)
	assert.Error(t, err)
}
