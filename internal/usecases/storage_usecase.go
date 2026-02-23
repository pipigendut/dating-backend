package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/infra/storage"
)

type StorageUsecase struct {
	storage *storage.S3Storage
}

func NewStorageUsecase(storage *storage.S3Storage) *StorageUsecase {
	return &StorageUsecase{storage: storage}
}

func (u *StorageUsecase) GetUploadURL(ctx context.Context, userID uuid.UUID) (string, string, error) {
	fileID := uuid.New().String()
	fileKey := fmt.Sprintf("users/%s/profile/%s.jpg", userID.String(), fileID)

	// Presigned URL for 10 minutes
	url, err := u.storage.GeneratePresignedPutURL(ctx, fileKey, 10*time.Minute)
	if err != nil {
		return "", "", err
	}

	return url, fileKey, nil
}
