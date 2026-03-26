package background

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
)

// Base payload that all our jobs must include to track the job ID.
type BaseJobPayload struct {
	JobID string `json:"job_id"`
}

type JobRouter struct {
	mux    *asynq.ServeMux
	jobRepo repository.JobRepository
}

func NewJobRouter(jobRepo repository.JobRepository) *JobRouter {
	mux := asynq.NewServeMux()
	router := &JobRouter{
		mux:     mux,
		jobRepo: jobRepo,
	}

	// Register our middleware to handle DB statuses automatically
	mux.Use(router.trackingMiddleware)

	return router
}

func (r *JobRouter) Mux() *asynq.ServeMux {
	return r.mux
}

// trackingMiddleware wraps every job execution to update the jobs table in PostgreSQL.
func (r *JobRouter) trackingMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		// 1. Extract internal JobID from payload
		var basePayload BaseJobPayload
		if err := json.Unmarshal(t.Payload(), &basePayload); err != nil {
			log.Printf("Failed to unmarshal base payload for task %s", t.Type())
			return h.ProcessTask(ctx, t) // Not tracked by DB
		}

		jobID, err := uuid.Parse(basePayload.JobID)
		if err != nil {
			return h.ProcessTask(ctx, t) // Not tracked by DB
		}

		// 2. Mark as processing and increment attempt
		r.jobRepo.IncrementJobAttempt(context.Background(), jobID)
		r.jobRepo.UpdateJobStatus(context.Background(), jobID, entities.JobStatusProcessing, nil)

		// 3. Execute the actual handler module
		handlerErr := h.ProcessTask(ctx, t)

		// 4. Handle Result
		if handlerErr != nil {
			errMsg := handlerErr.Error()
			r.jobRepo.UpdateJobStatus(context.Background(), jobID, entities.JobStatusFailed, &errMsg)
			return handlerErr // Return error to Asynq so it retries
		}

		// 5. Success
		r.jobRepo.UpdateJobStatus(context.Background(), jobID, entities.JobStatusCompleted, nil)
		return nil
	})
}

// RegisterHandler binds a task type to a specific handler function.
func (r *JobRouter) RegisterHandler(taskType string, handler asynq.HandlerFunc) {
	r.mux.HandleFunc(taskType, handler)
}
