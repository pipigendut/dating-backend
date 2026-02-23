package storage

import (
	"context"
	"time"
)

type StorageProvider interface {
	GeneratePresignedPutURL(ctx context.Context, key string, expiration time.Duration) (string, error)
	GeneratePresignedGetURL(ctx context.Context, key string, expiration time.Duration) (string, error)
	DeleteFile(ctx context.Context, key string) error
}
