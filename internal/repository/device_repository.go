package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type DeviceRepository interface {
	UpsertDevice(ctx context.Context, device *entities.Device) error
	UpdateFCMToken(ctx context.Context, userID uuid.UUID, deviceID string, token string) error
	GetUserDevices(ctx context.Context, userID uuid.UUID) ([]entities.Device, error)
	DeactivateDevice(ctx context.Context, userID uuid.UUID, deviceID string) error
}
