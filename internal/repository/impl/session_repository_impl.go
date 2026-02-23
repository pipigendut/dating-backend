package impl

import (
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
	"gorm.io/gorm"
)

type sessionRepo struct {
	db *gorm.DB
}

func NewSessionRepo(db *gorm.DB) repository.SessionRepository {
	return &sessionRepo{db: db}
}

// Device operations
func (r *sessionRepo) CreateDevice(device *entities.Device) error {
	return r.db.Create(device).Error
}

func (r *sessionRepo) GetDeviceByDeviceIDAndUserID(deviceID string, userID uuid.UUID) (*entities.Device, error) {
	var device entities.Device
	err := r.db.Where("device_id = ? AND user_id = ?", deviceID, userID).First(&device).Error
	return &device, err
}

func (r *sessionRepo) UpdateDevice(device *entities.Device) error {
	return r.db.Save(device).Error
}

func (r *sessionRepo) DeactivateDevice(deviceID string, userID uuid.UUID) error {
	return r.db.Model(&entities.Device{}).
		Where("device_id = ? AND user_id = ?", deviceID, userID).
		Update("is_active", false).Error
}

// RefreshToken operations
func (r *sessionRepo) CreateRefreshToken(token *entities.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *sessionRepo) GetRefreshTokenByHash(hash string) (*entities.RefreshToken, error) {
	var token entities.RefreshToken
	err := r.db.Where("token_hash = ?", hash).First(&token).Error
	return &token, err
}

func (r *sessionRepo) RevokeRefreshToken(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&entities.RefreshToken{}).
		Where("id = ?", id).
		Update("revoked_at", &now).Error
}

func (r *sessionRepo) RevokeAllUserTokens(userID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&entities.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", &now).Error
}
