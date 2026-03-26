package repository

import (
	"context"

	"time"
	"github.com/google/uuid"
)

type RedisRepository interface {
	// Online Status
	SetUserOnline(ctx context.Context, userID uuid.UUID) error
	SetUserOffline(ctx context.Context, userID uuid.UUID) error
	IsUserOnline(ctx context.Context, userID uuid.UUID) (bool, error)

	// Typing Indicator
	SetTyping(ctx context.Context, conversationID, userID uuid.UUID, isTyping bool) error
	IsTyping(ctx context.Context, conversationID, userID uuid.UUID) (bool, error)

	// Unread Count
	IncrementUnreadCount(ctx context.Context, userID, conversationID uuid.UUID) (int, error)
	ResetUnreadCount(ctx context.Context, userID, conversationID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID, conversationID uuid.UUID) (int, error)

	// Pub/Sub
	PublishEvent(ctx context.Context, channel string, payload interface{}) error

	// Boost
	SetUserBoost(ctx context.Context, userID uuid.UUID, expiresAt time.Time) error
	GetBoostExpiration(ctx context.Context, userID uuid.UUID) (*time.Time, error)
	DeleteUserBoost(ctx context.Context, userID uuid.UUID) error
}
