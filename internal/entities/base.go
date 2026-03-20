package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel provides standard ID and timestamps for all entities.
// GORM will automatically handle CreatedAt and UpdatedAt.
type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;index" json:"updated_at"`
}

// BeforeCreate hook to automatically generate UUID if not provided.
func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// SoftDeleteModel extends BaseModel with soft delete capability.
type SoftDeleteModel struct {
	BaseModel
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty" swaggertype:"string" format:"date-time"`
}
