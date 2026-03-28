package services

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/infra/storage"
)

type StorageService struct {
	storage storage.StorageProvider
}

func NewStorageService(storage storage.StorageProvider) *StorageService {
	return &StorageService{storage: storage}
}

func (u *StorageService) GetUploadURL(ctx context.Context, userID uuid.UUID) (string, string, error) {
	fileID := uuid.New().String()
	fileKey := fmt.Sprintf("users/%s/profile/%s.jpg", userID.String(), fileID)

	// Presigned URL for 10 minutes
	url, err := u.storage.GeneratePresignedPutURL(ctx, fileKey, 10*time.Minute)
	if err != nil {
		return "", "", err
	}

	return url, fileKey, nil
}

func (u *StorageService) GetChatUploadURL(ctx context.Context, conversationID uuid.UUID) (string, string, error) {
	fileID := uuid.New().String()
	fileKey := fmt.Sprintf("chat/conversations/%s/%s.jpg", conversationID.String(), fileID)

	// Presigned URL for 10 minutes
	url, err := u.storage.GeneratePresignedPutURL(ctx, fileKey, 10*time.Minute)
	if err != nil {
		return "", "", err
	}

	return url, fileKey, nil
}

func (u *StorageService) GetUploadURLPublic(ctx context.Context, clientID string) (string, string, error) {
	fileID := uuid.New().String()
	fileKey := fmt.Sprintf("users/%s/profile/%s.jpg", clientID, fileID)

	// Presigned URL for 10 minutes
	url, err := u.storage.GeneratePresignedPutURL(ctx, fileKey, 10*time.Minute)
	if err != nil {
		return "", "", err
	}

	return url, fileKey, nil
}

func (u *StorageService) GetPublicURL(key string) string {
	return u.storage.GetPublicURL(key)
}

func (u *StorageService) DeleteFile(ctx context.Context, key string) error {
	return u.storage.DeleteFile(ctx, key)
}

func (u *StorageService) GetFileContent(ctx context.Context, key string) (io.ReadCloser, error) {
	return u.storage.GetFileContent(ctx, key)
}
