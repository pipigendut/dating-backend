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
	TaskTypeNotificationMatch = "notification:match"
	TaskTypeNotificationLike  = "notification:like"
	NotificationDebounceSec   = 10
)

type NotificationTaskPayload struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	LastMessageID  uuid.UUID `json:"last_message_id"`
}

type MatchNotificationPayload struct {
	MatchID        uuid.UUID `json:"match_id"`
	SwiperEntityID uuid.UUID `json:"swiper_entity_id"`
	SwipedEntityID uuid.UUID `json:"swiped_entity_id"`
}

type LikeNotificationPayload struct {
	LikerEntityID  uuid.UUID `json:"liker_entity_id"`
	TargetEntityID uuid.UUID `json:"target_entity_id"`
	IsCrush        bool      `json:"is_crush"`
}

type NotificationService interface {
	TriggerChatNotification(ctx context.Context, conversationID, lastMessageID uuid.UUID) error
	TriggerMatchNotification(ctx context.Context, swiperEntityID, swipedEntityID, matchID uuid.UUID) error
	TriggerLikeNotification(ctx context.Context, likerEntityID, targetEntityID uuid.UUID, isCrush bool) error
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

func (s *notificationService) TriggerMatchNotification(ctx context.Context, swiperEntityID, swipedEntityID, matchID uuid.UUID) error {
	payload, _ := json.Marshal(MatchNotificationPayload{
		MatchID:        matchID,
		SwiperEntityID: swiperEntityID,
		SwipedEntityID: swipedEntityID,
	})

	task := asynq.NewTask(TaskTypeNotificationMatch, payload)
	_, err := s.asynqClient.EnqueueContext(ctx, task)
	return err
}

func (s *notificationService) TriggerLikeNotification(ctx context.Context, likerEntityID, targetEntityID uuid.UUID, isCrush bool) error {
	// Lightweight debounce: prevent duplicate notif in 60s window for same liker→target pair
	debounceKey := fmt.Sprintf("like_notif_debounce:%s:%s", likerEntityID, targetEntityID)
	isNew, err := s.redisRepo.SetNX(ctx, debounceKey, "1", 60*time.Second)
	if err != nil || !isNew {
		return nil
	}

	payload, _ := json.Marshal(LikeNotificationPayload{
		LikerEntityID:  likerEntityID,
		TargetEntityID: targetEntityID,
		IsCrush:        isCrush,
	})

	task := asynq.NewTask(TaskTypeNotificationLike, payload)
	_, err = s.asynqClient.EnqueueContext(ctx, task)
	return err
}
