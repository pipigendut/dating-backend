package entities

import (
	"github.com/google/uuid"
)

type NotificationSetting struct {
	BaseModel
	Type        string `gorm:"uniqueIndex;not null" json:"type"`
	Title       string `gorm:"not null" json:"title"`
	Description string `json:"description"`
	IsEnable    bool   `gorm:"default:true" json:"is_enable"`
}

type UserNotificationSetting struct {
	BaseModel
	UserID                uuid.UUID `gorm:"type:uuid;index;uniqueIndex:idx_user_notif_setting;not null" json:"user_id"`
	NotificationSettingID uuid.UUID `gorm:"type:uuid;index;uniqueIndex:idx_user_notif_setting;not null" json:"notification_setting_id"`
	IsEnable              bool      `gorm:"default:true" json:"is_enable"`

	// Associations
	User                *User                `gorm:"foreignKey:UserID" json:"-"`
	NotificationSetting *NotificationSetting `gorm:"foreignKey:NotificationSettingID" json:"setting,omitempty"`
}
