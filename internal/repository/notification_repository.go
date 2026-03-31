package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type NotificationRepository interface {
	GetGlobalSettings(ctx context.Context) ([]entities.NotificationSetting, error)
	GetGlobalSettingByType(ctx context.Context, notifType string) (*entities.NotificationSetting, error)
	
	GetUserSettings(ctx context.Context, userID uuid.UUID) ([]entities.UserNotificationSetting, error)
	GetUserSetting(ctx context.Context, userID, settingID uuid.UUID) (*entities.UserNotificationSetting, error)
	GetUserSettingByType(ctx context.Context, userID uuid.UUID, notifType string) (*entities.UserNotificationSetting, error)
	
	UpdateUserSetting(ctx context.Context, setting *entities.UserNotificationSetting) error
	
	// Combined view for the API
	GetUserSettingsWithMetadata(ctx context.Context, userID uuid.UUID) ([]entities.UserNotificationSetting, error)
}
