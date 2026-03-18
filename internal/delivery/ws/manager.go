package ws

import (
	"sync"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client represents a single WebSocket connection
type Client struct {
	UserID uuid.UUID
	Conn   *websocket.Conn
	Send   chan []byte
}

// Manager manages all active WebSocket connections
type Manager struct {
	// Map of UserID to their active clients (supports multiple devices per user)
	clients map[uuid.UUID][]*Client
	sync.RWMutex

	// Channels for registration and broadcast
	Register   chan *Client
	Unregister chan *Client
}

func NewManager() *Manager {
	return &Manager{
		clients:    make(map[uuid.UUID][]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (m *Manager) Run() {
	for {
		select {
		case client := <-m.Register:
			m.Lock()
			m.clients[client.UserID] = append(m.clients[client.UserID], client)
			m.Unlock()
		case client := <-m.Unregister:
			m.Lock()
			if clients, ok := m.clients[client.UserID]; ok {
				for i, c := range clients {
					if c == client {
						m.clients[client.UserID] = append(clients[:i], clients[i+1:]...)
						close(client.Send)
						break
					}
				}
				if len(m.clients[client.UserID]) == 0 {
					delete(m.clients, client.UserID)
				}
			}
			m.Unlock()
		}
	}
}

// SendToUser dispatches a message to all active connections of a specific user
func (m *Manager) SendToUser(userID uuid.UUID, message []byte) {
	m.RLock()
	defer m.RUnlock()

	if clients, ok := m.clients[userID]; ok {
		for _, client := range clients {
			select {
			case client.Send <- message:
			default:
				// If send buffer is full, we could drop or handle otherwise
			}
		}
	}
}

// IsUserOnline checks if a user has at least one active connection
func (m *Manager) IsUserOnline(userID uuid.UUID) bool {
	m.RLock()
	defer m.RUnlock()
	_, online := m.clients[userID]
	return online
}
