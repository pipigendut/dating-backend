package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type SwipeRepository interface {
	CreateSwipe(ctx context.Context, swipe *entities.Swipe) error
	GetMatch(ctx context.Context, entity1ID, entity2ID uuid.UUID) (*entities.Match, error)
	GetLikesSent(ctx context.Context, entityIDs []uuid.UUID, limit, offset, expiryHours int) ([]entities.Swipe, error)
	GetLikesYou(ctx context.Context, entityIDs []uuid.UUID, limit, offset, expiryHours int) ([]entities.Swipe, error)
	CountLikesYou(ctx context.Context, entityIDs []uuid.UUID, expiryHours int) (int64, error)
	DeleteMatch(ctx context.Context, entity1ID, entity2ID uuid.UUID) error
	DeleteSwipe(ctx context.Context, swiperEntityID, targetEntityID uuid.UUID) error
}
