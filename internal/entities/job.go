package entities

import (
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

type Job struct {
	BaseModel
	Type          string     `gorm:"index;not null"`
	Status        JobStatus  `gorm:"type:varchar(20);index;default:'pending'"`
	Payload       []byte     `gorm:"type:jsonb"` // Store arbitrary JSON string/bytes
	ReferenceID   *uuid.UUID `gorm:"type:uuid;index"`
	ReferenceType string     `gorm:"index"`
	Source        string     `gorm:"index"`
	ErrorMessage  *string    `gorm:"type:text"`
	Attempts      int        `gorm:"default:0"`
	MaxAttempts   int        `gorm:"default:3"`
	ProcessedAt   *time.Time `gorm:"index"`
}

// Ensure the table name is `jobs`
func (Job) TableName() string {
	return "jobs"
}
