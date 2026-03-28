package entities

import (
	"time"

	"github.com/google/uuid"
)

type GroupInvite struct {
	BaseModel
	GroupID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"group_id"`
	InviterID uuid.UUID  `gorm:"type:uuid;not null;index" json:"inviter_id"`
	Token     string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"token"`
	ExpiresAt time.Time  `gorm:"index" json:"expires_at"`
	UsedAt    *time.Time `gorm:"index" json:"used_at,omitempty"`

	// Associations
	Group   *Group `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	Inviter *User  `gorm:"foreignKey:InviterID" json:"inviter,omitempty"`
}

func (GroupInvite) TableName() string {
	return "group_invites"
}
