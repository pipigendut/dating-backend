package repository

import (
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type SessionRepository interface {
	// Device operations
	CreateDevice(device *entities.Device) error
	GetDeviceByDeviceIDAndUserID(deviceID string, userID uuid.UUID) (*entities.Device, error)
	UpdateDevice(device *entities.Device) error
	DeactivateDevice(deviceID string, userID uuid.UUID) error

	// RefreshToken operations
	CreateRefreshToken(token *entities.RefreshToken) error
	GetRefreshTokenByHash(hash string) (*entities.RefreshToken, error)
	RevokeRefreshToken(id uuid.UUID) error
	RevokeAllUserTokens(userID uuid.UUID) error
}
