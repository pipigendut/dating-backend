package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type SwipeRepository interface {
	CreateSwipe(ctx context.Context, swipe *entities.Swipe) error
	GetMatch(ctx context.Context, userID, targetUserID uuid.UUID) (*entities.Match, error)
	GetUnmatch(ctx context.Context, userID, targetUserID uuid.UUID) (*entities.Unmatch, error)
	UnmatchUser(ctx context.Context, userID, targetUserID uuid.UUID, matchID uuid.UUID, conversationID uuid.UUID) error
	GetLikesSent(ctx context.Context, userID uuid.UUID) ([]entities.Swipe, error)
	UnlikeUser(ctx context.Context, swiperID, swipedID uuid.UUID) error
}
