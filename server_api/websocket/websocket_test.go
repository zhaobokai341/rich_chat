package websocket

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, 10*time.Second, config.WRITEWAIT)
	assert.Equal(t, 60*time.Second, config.PONGWAIT)
	assert.Equal(t, 54*time.Second, config.PINGPERIOD)
	assert.Equal(t, int64(5120), config.MAXMESSAGESIZE)
	assert.Equal(t, 256, config.SEND_CHANNEL_BUFFER)
	assert.Equal(t, 10, config.MAX_MESSAGES_PER_SEC)
	assert.Equal(t, 200*time.Millisecond, config.OFFLINE_MESSAGE_DELAY)
}

func TestNewHub(t *testing.T) {
	hub := NewHub()

	assert.NotNil(t, hub)
	assert.NotNil(t, hub.connections)
	assert.NotNil(t, hub.sessionConnections)
	assert.NotNil(t, hub.registerChan)
	assert.NotNil(t, hub.unregisterChan)
	assert.NotNil(t, hub.joinSessionChan)
	assert.NotNil(t, hub.leaveSessionChan)
	assert.NotNil(t, hub.broadcastChan)
}

func TestHubRun(t *testing.T) {
	hub := NewHub()

	go hub.Run()

	// Give the hub some time to start
	time.Sleep(10 * time.Millisecond)

	// Just verify it's running - no cleanup needed for this test
	// The hub will be garbage collected when the test ends
	assert.NotNil(t, hub)
}

func TestHubRegisterConnection(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	conn := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
	}

	hub.Register(conn)

	// Give the hub time to process
	time.Sleep(10 * time.Millisecond)

	// Verify connection was registered
	registeredConn := hub.GetConnectionForUser(1)
	assert.NotNil(t, registeredConn)
	assert.Equal(t, conn, registeredConn)

	// Cleanup - remove from hub
	hub.mutex.Lock()
	delete(hub.connections, 1)
	hub.mutex.Unlock()
}

func TestHubUnregisterConnection(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	conn := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
	}

	// Register first
	hub.Register(conn)
	time.Sleep(10 * time.Millisecond)

	// Verify it's registered
	assert.True(t, hub.IsUserOnline(1))

	// Then unregister
	hub.Unregister(conn)
	time.Sleep(10 * time.Millisecond)

	// Verify connection was unregistered
	assert.False(t, hub.IsUserOnline(1))
}

func TestHubJoinAndLeaveSession(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	conn := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
	}

	// Register first
	hub.Register(conn)
	time.Sleep(50 * time.Millisecond)

	// Join session
	hub.JoinSession(conn, 100)
	time.Sleep(50 * time.Millisecond)

	// Verify connection is in session
	connections := hub.GetConnectionsForSession(100)
	assert.Len(t, connections, 1)
	assert.Contains(t, connections, conn)

	// Leave session
	hub.LeaveSession(conn, 100)
	time.Sleep(50 * time.Millisecond)

	// Verify connection left session
	connections = hub.GetConnectionsForSession(100)
	assert.Empty(t, connections)
}

func TestHubBroadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	config := DefaultConfig()
	conn1 := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
		config:         config,
	}

	conn2 := &Connection{
		UserID:         2,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
		config:         config,
	}

	// Register both connections
	hub.Register(conn1)
	hub.Register(conn2)
	time.Sleep(10 * time.Millisecond)

	// Join session
	hub.JoinSession(conn1, 100)
	hub.JoinSession(conn2, 100)
	time.Sleep(10 * time.Millisecond)

	// Broadcast message
	message := ServerMessage{
		Type:     "message",
		SenderID: 1,
		Content:  "hello",
	}
	hub.Broadcast(message, 100)

	// Give the hub time to process
	time.Sleep(10 * time.Millisecond)

	// Check if message was sent to conn2 (not conn1, since conn1 is sender)
	select {
	case received := <-conn2.send:
		assert.Contains(t, string(received), "hello")
	default:
		t.Fatal("Expected message in conn2 send channel")
	}

	// conn1 should not receive the message (it's the sender)
	select {
	case <-conn1.send:
		t.Fatal("Sender should not receive the broadcast message")
	default:
		// Expected
	}
}

func TestHubGetConnectionsForSession(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	config := DefaultConfig()
	conn1 := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
		config:         config,
	}

	conn2 := &Connection{
		UserID:         2,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
		config:         config,
	}

	// Register both connections
	hub.Register(conn1)
	hub.Register(conn2)
	hub.JoinSession(conn1, 100)
	hub.JoinSession(conn2, 100)
	time.Sleep(10 * time.Millisecond)

	// Get connections for session 100
	connections := hub.GetConnectionsForSession(100)
	assert.Len(t, connections, 2)

	// Get connections for non-existent session
	connections = hub.GetConnectionsForSession(999)
	assert.Empty(t, connections)
}

func TestHubIsUserOnline(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Check user is not online initially
	assert.False(t, hub.IsUserOnline(1))

	conn := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
	}

	// Register connection
	hub.Register(conn)
	time.Sleep(10 * time.Millisecond)

	// Check user is now online
	assert.True(t, hub.IsUserOnline(1))
}

func TestHubSendToUser(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	config := DefaultConfig()
	conn := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
		config:         config,
	}

	// Register connection
	hub.Register(conn)
	time.Sleep(10 * time.Millisecond)

	// Send message to user
	message := ServerMessage{
		Type:     "message",
		SenderID: 2,
		Content:  "hello",
	}
	err := hub.SendToUser(1, message)
	assert.NoError(t, err)

	// Verify message was sent
	select {
	case received := <-conn.send:
		assert.Contains(t, string(received), "hello")
	default:
		t.Fatal("Expected message in send channel")
	}

	// Send to non-existent user (should not error)
	err = hub.SendToUser(999, message)
	assert.NoError(t, err)
}

func TestHubBroadcastToSession(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	config := DefaultConfig()
	conn1 := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
		config:         config,
	}

	conn2 := &Connection{
		UserID:         2,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
		config:         config,
	}

	// Register both connections
	hub.Register(conn1)
	hub.Register(conn2)
	hub.JoinSession(conn1, 100)
	hub.JoinSession(conn2, 100)
	time.Sleep(10 * time.Millisecond)

	// Broadcast to session
	message := ServerMessage{
		Type:     "message",
		SenderID: 1,
		Content:  "hello",
	}
	hub.BroadcastToSession(100, message)

	// Give time to process
	time.Sleep(10 * time.Millisecond)

	// conn2 should receive the message
	select {
	case received := <-conn2.send:
		assert.Contains(t, string(received), "hello")
	default:
		t.Fatal("Expected message in conn2 send channel")
	}
}

func TestConnectionSafeCloseSend(t *testing.T) {
	conn := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
	}

	// Close send channel
	conn.safeCloseSend()

	// Verify sendClosed flag is set
	assert.Equal(t, int32(1), conn.sendClosed)

	// Verify send channel is closed
	_, ok := <-conn.send
	assert.False(t, ok)
}

func TestConnectionSendMessage(t *testing.T) {
	config := DefaultConfig()
	conn := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
		config:         config,
	}

	message := ServerMessage{
		Type:     "message",
		SenderID: 2,
		Content:  "hello",
	}
	err := conn.SendMessage(message)
	assert.NoError(t, err)

	// Verify message was sent
	select {
	case received := <-conn.send:
		assert.Contains(t, string(received), "hello")
	default:
		t.Fatal("Expected message in send channel")
	}
}

func TestConnectionSendMessageAfterClose(t *testing.T) {
	config := DefaultConfig()
	conn := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
		config:         config,
	}

	// Close send channel first
	conn.safeCloseSend()

	// Try to send message after close (should return error)
	message := ServerMessage{
		Type:     "message",
		SenderID: 2,
		Content:  "hello",
	}
	err := conn.SendMessage(message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closing")
}

func TestUpgrader(t *testing.T) {
	// Test that upgrader is properly configured
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// In production, this should check against allowed origins
			return true
		},
	}

	assert.NotNil(t, upgrader)
	assert.Equal(t, 1024, upgrader.ReadBufferSize)
	assert.Equal(t, 1024, upgrader.WriteBufferSize)
}

func TestServerMessageJSONMarshal(t *testing.T) {
	msg := ServerMessage{
		Type:      "message",
		SenderID:  1,
		SessionID: 100,
		Content:   "hello world",
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, string(data), "hello world")
	assert.Contains(t, string(data), "message")
}

func TestServerMessageToJSONError(t *testing.T) {
	// Create a message with content that can't be marshaled (circular reference)
	type circular struct {
		Self *circular
	}
	c := &circular{}
	c.Self = c

	msg := ServerMessage{
		Type:    "message",
		Content: c,
	}

	_, err := json.Marshal(msg)
	assert.Error(t, err)
}

func TestHubGetConnectionForUser(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Get non-existent connection
	conn := hub.GetConnectionForUser(1)
	assert.Nil(t, conn)

	// Register a connection
	config := DefaultConfig()
	newConn := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
		config:         config,
	}
	hub.Register(newConn)
	time.Sleep(50 * time.Millisecond)

	// Get the connection
	retrievedConn := hub.GetConnectionForUser(1)
	assert.NotNil(t, retrievedConn)
	assert.Equal(t, newConn, retrievedConn)
}

func TestHubBroadcastToNonExistentSession(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Broadcast to non-existent session should not panic
	message := ServerMessage{
		Type:     "message",
		SenderID: 1,
		Content:  "hello",
	}
	hub.BroadcastToSession(999, message)
	time.Sleep(10 * time.Millisecond)

	// No error should occur
	assert.True(t, true)
}

func TestHubSendToOfflineUser(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Send to offline user should not error
	message := ServerMessage{
		Type:     "message",
		SenderID: 1,
		Content:  "hello",
	}
	err := hub.SendToUser(999, message)
	assert.NoError(t, err)
}

func TestConnectionAllowMessage(t *testing.T) {
	config := DefaultConfig()
	conn := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
		config:         config,
	}

	// First message should be allowed
	allowed := conn.allowMessage()
	assert.True(t, allowed)
}

func TestConnectionSendMessageWithRateLimitDisabled(t *testing.T) {
	config := DefaultConfig()
	conn := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
		config:         config,
	}

	message := ServerMessage{
		Type:     "message",
		SenderID: 2,
		Content:  "hello",
	}

	// Send message without rate limit
	err := conn.SendMessageWithRateLimit(message, false)
	assert.NoError(t, err)

	// Verify message was sent
	select {
	case received := <-conn.send:
		assert.Contains(t, string(received), "hello")
	default:
		t.Fatal("Expected message in send channel")
	}
}

func TestServerMessageStruct(t *testing.T) {
	ts := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	msg := ServerMessage{
		Type:      "typing",
		SenderID:  1,
		SessionID: 5,
		Content:   map[string]interface{}{"key": "value"},
		Timestamp: ts,
	}

	assert.Equal(t, "typing", msg.Type)
	assert.Equal(t, 1, msg.SenderID)
	assert.Equal(t, 5, msg.SessionID)
	assert.Equal(t, ts, msg.Timestamp)
}

func TestClientMessageStruct(t *testing.T) {
	msg := ClientMessage{
		Type:    "chat",
		Content: "hello",
	}

	assert.Equal(t, "chat", msg.Type)
	assert.Equal(t, "hello", msg.Content)
}

func TestConnectionSendResponse(t *testing.T) {
	config := DefaultConfig()
	conn := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
		config:         config,
	}

	conn.sendResponse("test_type", 100, "test message")

	select {
	case received := <-conn.send:
		assert.Contains(t, string(received), "test_type")
		assert.Contains(t, string(received), "test message")
	default:
		t.Fatal("Expected response in send channel")
	}
}

func TestConnectionSendError(t *testing.T) {
	config := DefaultConfig()
	conn := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
		config:         config,
	}

	conn.sendError("test error", "test_type")

	select {
	case received := <-conn.send:
		assert.Contains(t, string(received), "error")
		assert.Contains(t, string(received), "test error")
	default:
		t.Fatal("Expected error message in send channel")
	}
}

func TestConnectionWritePumpWithClosedChannel(t *testing.T) {
	config := DefaultConfig()
	conn := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
		config:         config,
	}

	// Close the send channel
	close(conn.send)

	// WritePump should handle closed channel gracefully
	// We can't test WritePump directly without a real WebSocket connection,
	// but we can verify the send channel behavior
	_, ok := <-conn.send
	assert.False(t, ok)
}

func TestHubMultipleUsersInSession(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	config := DefaultConfig()
	conn1 := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
		config:         config,
	}

	conn2 := &Connection{
		UserID:         2,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
		config:         config,
	}

	conn3 := &Connection{
		UserID:         3,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
		config:         config,
	}

	// Register all connections
	hub.Register(conn1)
	hub.Register(conn2)
	hub.Register(conn3)
	time.Sleep(50 * time.Millisecond)

	// All join the same session
	hub.JoinSession(conn1, 100)
	hub.JoinSession(conn2, 100)
	hub.JoinSession(conn3, 100)
	time.Sleep(50 * time.Millisecond)

	// Verify all are in the session
	connections := hub.GetConnectionsForSession(100)
	assert.Len(t, connections, 3)

	// Broadcast from conn1 - should go to conn2 and conn3
	message := ServerMessage{
		Type:     "message",
		SenderID: 1,
		Content:  "hello from conn1",
	}
	hub.Broadcast(message, 100)
	time.Sleep(50 * time.Millisecond)

	// conn2 and conn3 should receive the message
	select {
	case <-conn2.send:
		// Expected
	default:
		t.Fatal("Expected message in conn2 send channel")
	}

	select {
	case <-conn3.send:
		// Expected
	default:
		t.Fatal("Expected message in conn3 send channel")
	}

	// conn1 should not receive (it's the sender)
	select {
	case <-conn1.send:
		t.Fatal("Sender should not receive the broadcast message")
	default:
		// Expected
	}
}

func TestConnectionRateLimiting(t *testing.T) {
	config := DefaultConfig()
	conn := &Connection{
		UserID:               1,
		ActiveSessions:       make(map[int]bool),
		send:                 make(chan []byte, 256),
		config:               config,
		maxMessagesPerSecond: 1, // 1 message per second
	}

	message := ServerMessage{
		Type:     "message",
		SenderID: 2,
		Content:  "hello",
	}

	// First message should succeed
	err := conn.SendMessageWithRateLimit(message, true)
	assert.NoError(t, err)

	// Consume the message from the channel
	select {
	case <-conn.send:
		// Expected
	default:
		t.Fatal("Expected message in send channel")
	}

	// Second message immediately should fail due to rate limit
	err = conn.SendMessageWithRateLimit(message, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit")
}

func TestHubUnregisterRemovesFromSessions(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	config := DefaultConfig()
	conn := &Connection{
		UserID:         1,
		ActiveSessions: map[int]bool{100: true}, // Pre-populate active sessions
		send:           make(chan []byte, 256),
		config:         config,
	}

	// Register
	hub.Register(conn)
	time.Sleep(50 * time.Millisecond)

	// Manually add to session (simulating JoinSession processing)
	hub.JoinSession(conn, 100)
	time.Sleep(50 * time.Millisecond)

	// Verify in session
	connections := hub.GetConnectionsForSession(100)
	assert.Len(t, connections, 1)

	// Unregister
	hub.Unregister(conn)
	time.Sleep(50 * time.Millisecond)

	// Verify removed from session
	connections = hub.GetConnectionsForSession(100)
	assert.Empty(t, connections)

	// Verify user is offline
	assert.False(t, hub.IsUserOnline(1))
}

func TestServerMessageWithDifferentContentTypes(t *testing.T) {
	config := DefaultConfig()
	conn := &Connection{
		UserID:         1,
		ActiveSessions: make(map[int]bool),
		send:           make(chan []byte, 256),
		config:         config,
	}

	// Test with string content
	msg1 := ServerMessage{
		Type:      "chat",
		SenderID:  1,
		SessionID: 100,
		Content:   "text message",
		Timestamp: time.Now(),
	}
	err := conn.SendMessage(msg1)
	assert.NoError(t, err)

	// Test with map content
	msg2 := ServerMessage{
		Type:      "chat",
		SenderID:  1,
		SessionID: 100,
		Content:   map[string]interface{}{"text": "hello", "user": "test"},
		Timestamp: time.Now(),
	}
	err = conn.SendMessage(msg2)
	assert.NoError(t, err)

	// Test with numeric content
	msg3 := ServerMessage{
		Type:      "typing",
		SenderID:  1,
		SessionID: 100,
		Content:   12345,
		Timestamp: time.Now(),
	}
	err = conn.SendMessage(msg3)
	assert.NoError(t, err)

	// Verify all messages were sent
	for i := 0; i < 3; i++ {
		select {
		case <-conn.send:
			// Expected
		default:
			t.Fatal("Expected message in send channel")
		}
	}
}
