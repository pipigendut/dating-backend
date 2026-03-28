package entities

import (
	"github.com/google/uuid"
)

type Group struct {
	BaseModel
	EntityID uuid.UUID `gorm:"type:uuid;not null;index" json:"entity_id"`
	Name     string    `gorm:"type:varchar(255);not null" json:"name"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null;index" json:"created_by"`

	// Associations
	Entity  *Entity       `gorm:"foreignKey:EntityID" json:"entity,omitempty"`
	Members []GroupMember `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE" json:"members,omitempty"`
}

type GroupMember struct {
	BaseModel
	GroupID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_group_pair;index" json:"group_id"`
	UserID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_group_pair;index" json:"user_id"`
	IsAdmin bool      `gorm:"default:false" json:"is_admin"`

	// Associations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Group) TableName() string {
	return "groups"
}

func (GroupMember) TableName() string {
	return "group_members"
}
