package repository

import (
	"context"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type ConfigRepository interface {
	GetAll(ctx context.Context) ([]entities.AppConfig, error)
	Get(ctx context.Context, key string) (*entities.AppConfig, error)
	Set(ctx context.Context, key, value string) error
}
