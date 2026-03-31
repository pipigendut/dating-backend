package impl

import (
	"context"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type deviceRepo struct {
	db *gorm.DB
}

func NewDeviceRepository(db *gorm.DB) repository.DeviceRepository {
	return &deviceRepo{db: db}
}

func (r *deviceRepo) UpsertDevice(ctx context.Context, device *entities.Device) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "device_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"device_name", "device_model", "os_version", "app_version", "fcm_token", "is_active", "last_ip", "updated_at",
		}),
	}).Create(device).Error
}

func (r *deviceRepo) UpdateFCMToken(ctx context.Context, userID uuid.UUID, deviceID string, token string) error {
	return r.db.WithContext(ctx).Model(&entities.Device{}).
		Where("user_id = ? AND device_id = ?", userID, deviceID).
		Updates(map[string]interface{}{
			"fcm_token": token,
			"is_active": true,
		}).Error
}

func (r *deviceRepo) GetUserDevices(ctx context.Context, userID uuid.UUID) ([]entities.Device, error) {
	var devices []entities.Device
	err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = ?", userID, true).Find(&devices).Error
	return devices, err
}

func (r *deviceRepo) DeactivateDevice(ctx context.Context, userID uuid.UUID, deviceID string) error {
	return r.db.WithContext(ctx).Model(&entities.Device{}).
		Where("user_id = ? AND device_id = ?", userID, deviceID).
		Update("is_active", false).Error
}
