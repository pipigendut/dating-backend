package hub

import (
	"context"
	"sync"

	"github.com/pipigendut/dating-backend/internal/repository"
	"github.com/redis/go-redis/v9"
)

// Manager manages one or more WebSocket hubs.
// It provides a central place to initialize and access the WS infrastructure.
type Manager struct {
	chatHub *Hub
	mu      sync.RWMutex
}

// NewManager creates a new WebSocket manager.
func NewManager(redisRepo repository.RedisRepository, rdb *redis.Client) *Manager {
	m := &Manager{
		chatHub: NewHub(redisRepo),
	}
	
	// Pre-start the chat hub
	ctx := context.Background()
	go m.chatHub.Run(ctx)
	go m.chatHub.ListenToRedisPubSub(ctx, rdb)
	
	return m
}

// GetChatHub returns the main chat hub.
func (m *Manager) GetChatHub() *Hub {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.chatHub
}
