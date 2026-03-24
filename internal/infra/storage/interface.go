package storage

import (
	"context"
	"io"
	"time"
)

type StorageProvider interface {
	GeneratePresignedPutURL(ctx context.Context, key string, expiration time.Duration) (string, error)
	GeneratePresignedGetURL(ctx context.Context, key string, expiration time.Duration) (string, error)
	GetPublicURL(key string) string
	GetFileContent(ctx context.Context, key string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, key string) error
}
