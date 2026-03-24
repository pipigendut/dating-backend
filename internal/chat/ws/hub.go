package ws

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/repository"
	"github.com/redis/go-redis/v9"
)

type Hub struct {
	clients    map[uuid.UUID]*Client
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	
	redisRepo  repository.RedisRepository
}

func NewHub(redisRepo repository.RedisRepository) *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		redisRepo:  redisRepo,
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.userID] = client
			h.mu.Unlock()
			// Notify online status via Redis
			h.redisRepo.SetUserOnline(ctx, client.userID)
			
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
			}
			h.mu.Unlock()
			// Notify offline status
			h.redisRepo.SetUserOffline(ctx, client.userID)

		case <-ctx.Done():
			return
		}
	}
}

func (h *Hub) BroadcastEvent(event WSEvent, targetUserID uuid.UUID) {
	h.mu.RLock()
	client, ok := h.clients[targetUserID]
	h.mu.RUnlock()

	if ok {
		data, _ := json.Marshal(event)
		client.send <- data
	}
}

// ListenToRedisPubSub listens to the global chat:events channel and broadcasts to local clients
func (h *Hub) ListenToRedisPubSub(ctx context.Context, rdb *redis.Client) {
	pubsub := rdb.Subscribe(ctx, "chat:events")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var payload struct {
			TargetUserID uuid.UUID `json:"target_user_id"`
			Event        WSEvent   `json:"event"`
		}
		if err := json.Unmarshal([]byte(msg.Payload), &payload); err == nil {
			h.BroadcastEvent(payload.Event, payload.TargetUserID)
		}
	}
}
