package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/pipigendut/dating-backend/internal/repository"
)

const (
	TaskTypeNotificationGroup = "notification:group"
	NotificationDebounceSec   = 10
)

type NotificationTaskPayload struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	LastMessageID  uuid.UUID `json:"last_message_id"`
}

type NotificationService interface {
	TriggerChatNotification(ctx context.Context, conversationID, lastMessageID uuid.UUID) error
}

type notificationService struct {
	asynqClient *asynq.Client
	redisRepo   repository.RedisRepository
}

func NewNotificationService(asynqClient *asynq.Client, redisRepo repository.RedisRepository) NotificationService {
	return &notificationService{
		asynqClient: asynqClient,
		redisRepo:   redisRepo,
	}
}

func (s *notificationService) TriggerChatNotification(ctx context.Context, conversationID, lastMessageID uuid.UUID) error {
	// Debounce check: check if already scheduled in Redis
	debounceKey := fmt.Sprintf("notify_debounce:%s", conversationID)
	
	// Try to set key with TTL. If successful (NX), we are the ones to schedule the task.
	// This uses Redis as a lock/timer.
	isNew, err := s.redisRepo.SetNX(ctx, debounceKey, "1", time.Duration(NotificationDebounceSec)*time.Second)
	if err != nil {
		return err
	}

	if !isNew {
		// Already debouncing, skip enqueuing again
		return nil
	}

	// Not debouncing, schedule a job for NotificationDebounceSec in the future
	payload, _ := json.Marshal(NotificationTaskPayload{
		ConversationID: conversationID,
		LastMessageID:  lastMessageID,
	})

	task := asynq.NewTask(TaskTypeNotificationGroup, payload)
	_, err = s.asynqClient.EnqueueContext(ctx, task, asynq.ProcessIn(time.Duration(NotificationDebounceSec)*time.Second))
	
	return err
}
