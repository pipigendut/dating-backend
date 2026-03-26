package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type JobRepository interface {
	CreateJob(ctx context.Context, job *entities.Job) error
	GetJobByID(ctx context.Context, id uuid.UUID) (*entities.Job, error)
	UpdateJobStatus(ctx context.Context, id uuid.UUID, status entities.JobStatus, errMessage *string) error
	IncrementJobAttempt(ctx context.Context, id uuid.UUID) error
}
