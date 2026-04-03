package impl

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) repository.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) GetGlobalSettings(ctx context.Context) ([]entities.NotificationSetting, error) {
	var settings []entities.NotificationSetting
	err := r.db.WithContext(ctx).Find(&settings).Error
	return settings, err
}

func (r *notificationRepository) GetGlobalSettingByType(ctx context.Context, notifType string) (*entities.NotificationSetting, error) {
	var setting entities.NotificationSetting
	err := r.db.WithContext(ctx).Where("type = ?", notifType).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *notificationRepository) GetUserSettings(ctx context.Context, userID uuid.UUID) ([]entities.UserNotificationSetting, error) {
	var settings []entities.UserNotificationSetting
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&settings).Error
	return settings, err
}

func (r *notificationRepository) GetUserSetting(ctx context.Context, userID, settingID uuid.UUID) (*entities.UserNotificationSetting, error) {
	var setting entities.UserNotificationSetting
	err := r.db.WithContext(ctx).Where("user_id = ? AND notification_setting_id = ?", userID, settingID).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *notificationRepository) GetUserSettingByType(ctx context.Context, userID uuid.UUID, notifType string) (*entities.UserNotificationSetting, error) {
	var setting entities.UserNotificationSetting
	err := r.db.WithContext(ctx).
		Joins("JOIN notification_settings ON notification_settings.id = user_notification_settings.notification_setting_id").
		Where("user_notification_settings.user_id = ? AND notification_settings.type = ?", userID, notifType).
		First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *notificationRepository) UpdateUserSetting(ctx context.Context, setting *entities.UserNotificationSetting) error {
	var existing entities.UserNotificationSetting
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND notification_setting_id = ?", setting.UserID, setting.NotificationSettingID).
		First(&existing).Error

	if err == nil {
		// Record exists, update it
		return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
			"is_enable":  setting.IsEnable,
			"updated_at": time.Now(),
		}).Error
	}

	// Record doesn't exist, create it
	return r.db.WithContext(ctx).Create(setting).Error
}

func (r *notificationRepository) GetUserSettingsWithMetadata(ctx context.Context, userID uuid.UUID) ([]entities.UserNotificationSetting, error) {
	var masterSettings []entities.NotificationSetting
	if err := r.db.WithContext(ctx).Find(&masterSettings).Error; err != nil {
		return nil, err
	}

	var userSettings []entities.UserNotificationSetting
	if err := r.db.WithContext(ctx).Preload("NotificationSetting").Where("user_id = ?", userID).Find(&userSettings).Error; err != nil {
		return nil, err
	}

	userSettingMap := make(map[uuid.UUID]entities.UserNotificationSetting)
	for _, us := range userSettings {
		userSettingMap[us.NotificationSettingID] = us
	}

	var result []entities.UserNotificationSetting
	for _, ms := range masterSettings {
		if us, exists := userSettingMap[ms.ID]; exists {
			// Record exists, use its value (either true or false)
			result = append(result, us)
		} else {
			// No record exists, user asked for it to be 'false' if not found
			result = append(result, entities.UserNotificationSetting{
				UserID:                userID,
				NotificationSettingID: ms.ID,
				IsEnable:              false,
				NotificationSetting:   &ms,
			})
		}
	}

	return result, nil
}

func (r *notificationRepository) DeactivateAllUserSettings(ctx context.Context, userID uuid.UUID) error {
	var masterSettings []entities.NotificationSetting
	if err := r.db.WithContext(ctx).Find(&masterSettings).Error; err != nil {
		return err
	}

	for _, ms := range masterSettings {
		setting := &entities.UserNotificationSetting{
			UserID:                userID,
			NotificationSettingID: ms.ID,
			IsEnable:              false,
		}
		if err := r.UpdateUserSetting(ctx, setting); err != nil {
			return err
		}
	}
	return nil
}

func (r *notificationRepository) ActivateAllUserSettings(ctx context.Context, userID uuid.UUID) error {
	var masterSettings []entities.NotificationSetting
	if err := r.db.WithContext(ctx).Find(&masterSettings).Error; err != nil {
		return err
	}

	for _, ms := range masterSettings {
		setting := &entities.UserNotificationSetting{
			UserID:                userID,
			NotificationSettingID: ms.ID,
			IsEnable:              true,
		}
		if err := r.UpdateUserSetting(ctx, setting); err != nil {
			return err
		}
	}
	return nil
}
