package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/repository"
	"github.com/redis/go-redis/v9"
)

type redisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) repository.RedisRepository {
	return &redisRepository{client: client}
}

const (
	KeyOnlineStatus   = "user:online:%s"
	KeyTypingIndicator = "chat:typing:%s:%s"
	KeyUnreadCount    = "chat:unread:%s:%s"
)

func (r *redisRepository) SetUserOnline(ctx context.Context, userID uuid.UUID) error {
	key := fmt.Sprintf(KeyOnlineStatus, userID.String())
	return r.client.Set(ctx, key, "true", 60*time.Second).Err()
}

func (r *redisRepository) SetUserOffline(ctx context.Context, userID uuid.UUID) error {
	key := fmt.Sprintf(KeyOnlineStatus, userID.String())
	return r.client.Del(ctx, key).Err()
}

func (r *redisRepository) IsUserOnline(ctx context.Context, userID uuid.UUID) (bool, error) {
	key := fmt.Sprintf(KeyOnlineStatus, userID.String())
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	return val == "true", err
}

func (r *redisRepository) SetTyping(ctx context.Context, conversationID, userID uuid.UUID, isTyping bool) error {
	key := fmt.Sprintf(KeyTypingIndicator, conversationID.String(), userID.String())
	if isTyping {
		return r.client.Set(ctx, key, "true", 10*time.Second).Err()
	}
	return r.client.Del(ctx, key).Err()
}

func (r *redisRepository) IsTyping(ctx context.Context, conversationID, userID uuid.UUID) (bool, error) {
	key := fmt.Sprintf(KeyTypingIndicator, conversationID.String(), userID.String())
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	return val == "true", err
}

func (r *redisRepository) IncrementUnreadCount(ctx context.Context, userID, conversationID uuid.UUID) (int, error) {
	key := fmt.Sprintf(KeyUnreadCount, userID.String(), conversationID.String())
	val, err := r.client.Incr(ctx, key).Result()
	return int(val), err
}

func (r *redisRepository) ResetUnreadCount(ctx context.Context, userID, conversationID uuid.UUID) error {
	key := fmt.Sprintf(KeyUnreadCount, userID.String(), conversationID.String())
	return r.client.Del(ctx, key).Err()
}

func (r *redisRepository) GetUnreadCount(ctx context.Context, userID, conversationID uuid.UUID) (int, error) {
	key := fmt.Sprintf(KeyUnreadCount, userID.String(), conversationID.String())
	val, err := r.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (r *redisRepository) PublishEvent(ctx context.Context, channel string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, channel, data).Err()
}
