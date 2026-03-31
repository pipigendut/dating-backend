package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
)

type NotificationConfigService interface {
	GetUserSettings(ctx context.Context, userID uuid.UUID) ([]entities.UserNotificationSetting, error)
	UpdateUserSetting(ctx context.Context, userID, settingID uuid.UUID, isEnable bool) error
}

type notificationConfigService struct {
	notifRepo repository.NotificationRepository
}

func NewNotificationConfigService(notifRepo repository.NotificationRepository) NotificationConfigService {
	return &notificationConfigService{
		notifRepo: notifRepo,
	}
}

func (s *notificationConfigService) GetUserSettings(ctx context.Context, userID uuid.UUID) ([]entities.UserNotificationSetting, error) {
	return s.notifRepo.GetUserSettingsWithMetadata(ctx, userID)
}

func (s *notificationConfigService) UpdateUserSetting(ctx context.Context, userID, settingID uuid.UUID, isEnable bool) error {
	setting := &entities.UserNotificationSetting{
		UserID:                userID,
		NotificationSettingID: settingID,
		IsEnable:              isEnable,
	}
	return s.notifRepo.UpdateUserSetting(ctx, setting)
}

