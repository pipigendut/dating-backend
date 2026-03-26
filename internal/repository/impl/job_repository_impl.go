package impl

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
	"gorm.io/gorm"
)

type jobRepository struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) repository.JobRepository {
	return &jobRepository{
		db: db,
	}
}

func (r *jobRepository) CreateJob(ctx context.Context, job *entities.Job) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *jobRepository) GetJobByID(ctx context.Context, id uuid.UUID) (*entities.Job, error) {
	var job entities.Job
	if err := r.db.WithContext(ctx).First(&job, id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *jobRepository) UpdateJobStatus(ctx context.Context, id uuid.UUID, status entities.JobStatus, errMessage *string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if errMessage != nil {
		updates["error_message"] = *errMessage
	}

	if status == entities.JobStatusCompleted {
		now := time.Now()
		updates["processed_at"] = &now
	}

	return r.db.WithContext(ctx).Model(&entities.Job{}).Where("id = ?", id).Updates(updates).Error
}

func (r *jobRepository) IncrementJobAttempt(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&entities.Job{}).Where("id = ?", id).
		Update("attempts", gorm.Expr("attempts + 1")).Error
}
