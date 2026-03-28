package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type EntityRepository interface {
	Create(ctx context.Context, entity *entities.Entity) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Entity, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
