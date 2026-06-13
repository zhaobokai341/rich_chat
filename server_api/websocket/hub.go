package websocket

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered connections by user ID.
	connections map[int]*Connection

	// Registered connections by session ID.
	sessionConnections map[int]map[*Connection]bool

	// Inbound requests from the connections.
	registerChan   chan *Connection
	unregisterChan chan *Connection

	// Inbound requests for session management
	joinSessionChan  chan JoinSessionRequest
	leaveSessionChan chan LeaveSessionRequest
	broadcastChan    chan BroadcastRequest

	// Mutex to protect concurrent access to maps
	mutex sync.RWMutex
}

// JoinSessionRequest represents a request to join a chat session
type JoinSessionRequest struct {
	Connection *Connection
	SessionID  int
}

// LeaveSessionRequest represents a request to leave a chat session
type LeaveSessionRequest struct {
	Connection *Connection
	SessionID  int
}

// BroadcastRequest represents a request to broadcast a message to a session
type BroadcastRequest struct {
	Message   ServerMessage
	SessionID int
}

// NewHub creates a new hub instance
func NewHub() *Hub {
	return &Hub{
		connections:        make(map[int]*Connection),
		sessionConnections: make(map[int]map[*Connection]bool),
		registerChan:       make(chan *Connection),
		unregisterChan:     make(chan *Connection),
		joinSessionChan:    make(chan JoinSessionRequest),
		leaveSessionChan:   make(chan LeaveSessionRequest),
		broadcastChan:      make(chan BroadcastRequest),
	}
}

// Register adds a connection to the hub (direct method call for interface)
func (h *Hub) Register(conn *Connection) {
	h.registerChan <- conn
}

// Unregister removes a connection from the hub (direct method call for interface)
func (h *Hub) Unregister(conn *Connection) {
	h.unregisterChan <- conn
}

// JoinSession adds a connection to a session (direct method call for interface)
func (h *Hub) JoinSession(conn *Connection, sessionID int) {
	h.joinSessionChan <- JoinSessionRequest{
		Connection: conn,
		SessionID:  sessionID,
	}
}

// LeaveSession removes a connection from a session (direct method call for interface)
func (h *Hub) LeaveSession(conn *Connection, sessionID int) {
	h.leaveSessionChan <- LeaveSessionRequest{
		Connection: conn,
		SessionID:  sessionID,
	}
}

// Broadcast sends a message to all connections in a session (direct method call for interface)
func (h *Hub) Broadcast(msg ServerMessage, sessionID int) {
	h.broadcastChan <- BroadcastRequest{
		Message:   msg,
		SessionID: sessionID,
	}
}

// Run starts the hub and handles all events
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.registerChan:
			h.mutex.Lock()
			h.connections[conn.UserID] = conn
			h.mutex.Unlock()

		case conn := <-h.unregisterChan:
			h.mutex.Lock()
			if _, ok := h.connections[conn.UserID]; ok {
				// Remove connection from all sessions
				for sessionID := range conn.ActiveSessions {
					h.removeFromSession(conn, sessionID)
				}

				delete(h.connections, conn.UserID)
				conn.safeCloseSend()
			}
			h.mutex.Unlock()

		case req := <-h.joinSessionChan:
			h.mutex.Lock()
			h.addToSession(req.Connection, req.SessionID)
			h.mutex.Unlock()

		case req := <-h.leaveSessionChan:
			h.mutex.Lock()
			h.removeFromSession(req.Connection, req.SessionID)
			h.mutex.Unlock()

		case req := <-h.broadcastChan:
			h.mutex.RLock()
			connections, ok := h.sessionConnections[req.SessionID]
			if !ok {
				h.mutex.RUnlock()
				continue
			}

			connsCopy := make([]*Connection, 0, len(connections))
			for conn := range connections {
				if conn.UserID != req.Message.SenderID {
					connsCopy = append(connsCopy, conn)
				}
			}
			h.mutex.RUnlock()

			for _, conn := range connsCopy {
				if err := conn.SendMessage(req.Message); err != nil {
					log.Warnf("Failed to broadcast message to user %d: %v", conn.UserID, err)
				}
			}
		}
	}
}

// addToSession adds a connection to a session
func (h *Hub) addToSession(conn *Connection, sessionID int) {
	// Ensure the session map exists
	if _, ok := h.sessionConnections[sessionID]; !ok {
		h.sessionConnections[sessionID] = make(map[*Connection]bool)
	}

	// Add the connection to the session
	h.sessionConnections[sessionID][conn] = true
}

// removeFromSession removes a connection from a session
func (h *Hub) removeFromSession(conn *Connection, sessionID int) {
	if sessionConns, ok := h.sessionConnections[sessionID]; ok {
		// Remove the connection from the session
		delete(sessionConns, conn)

		// Clean up the session map if it's empty
		if len(sessionConns) == 0 {
			delete(h.sessionConnections, sessionID)
		}
	}

	// Also remove from connection's active sessions
	delete(conn.ActiveSessions, sessionID)
}

// GetConnectionsForSession returns all connections in a specific session
func (h *Hub) GetConnectionsForSession(sessionID int) []*Connection {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var connections []*Connection
	if sessionConns, ok := h.sessionConnections[sessionID]; ok {
		for conn := range sessionConns {
			connections = append(connections, conn)
		}
	}

	return connections
}

// GetConnectionForUser returns the connection for a specific user if connected
func (h *Hub) GetConnectionForUser(userID int) *Connection {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.connections[userID]
}

// IsUserOnline checks if a user is currently connected via WebSocket
func (h *Hub) IsUserOnline(userID int) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	_, exists := h.connections[userID]
	return exists
}

// BroadcastToSession sends a message to all connections in a specific session
func (h *Hub) BroadcastToSession(sessionID int, msg ServerMessage) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if sessionConns, ok := h.sessionConnections[sessionID]; ok {
		for conn := range sessionConns {
			// Don't send the message back to the sender
			if conn.UserID != msg.SenderID {
				_ = conn.SendMessage(msg)
			}
		}
	}
}

// SendToUser sends a message to a specific user if they are connected
func (h *Hub) SendToUser(userID int, msg ServerMessage) error {
	h.mutex.RLock()
	conn, exists := h.connections[userID]
	h.mutex.RUnlock()

	if !exists {
		return nil // User not connected, ignore
	}

	return conn.SendMessage(msg)
}

// Close closes the hub and all connections
func (h *Hub) Close() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Close all connections
	for _, conn := range h.connections {
		conn.safeCloseSend()
		conn.ws.Close()
	}

	// Clear maps
	h.connections = make(map[int]*Connection)
	h.sessionConnections = make(map[int]map[*Connection]bool)
}
