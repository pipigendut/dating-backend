package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/pipigendut/dating-backend/internal/repository"
"github.com/pipigendut/dating-backend/internal/services"
)

type NotificationWorker struct {
	chatRepo  repository.ChatRepository
	userRepo  repository.UserRepository
	redisRepo repository.RedisRepository
}

func NewNotificationWorker(chatRepo repository.ChatRepository, userRepo repository.UserRepository, redisRepo repository.RedisRepository) *NotificationWorker {
	return &NotificationWorker{
		chatRepo:  chatRepo,
		userRepo:  userRepo,
		redisRepo: redisRepo,
	}
}

func (w *NotificationWorker) HandleNotificationGroupTask(ctx context.Context, t *asynq.Task) error {
	var p services.NotificationTaskPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("[Worker] Processing debounced notification for Conversation %s", p.ConversationID)

	// 1. Fetch conversation and participants
	conv, err := w.chatRepo.GetConversationByID(ctx, p.ConversationID)
	if err != nil {
		return err
	}

	// 2. Fetch the last message to show in preview
	messages, err := w.chatRepo.GetConversationMessages(ctx, p.ConversationID, 1, 0)
	if err != nil || len(messages) == 0 {
		return nil
	}
	lastMsg := messages[0]

	// 3. For each participant (except sender), check if they should get a push
	for _, part := range conv.Participants {
		if part.UserID == lastMsg.SenderID {
			continue
		}

		// check if user is currently online/active in this chat (optional skip if online)
		// For dating apps, we usually send push if they've been inactive for > 10s
		
		// 4. Send FCM
		// In a real implementation, we would fetch tokens from `devices` table
		// and use firebase-admin-sdk.
		log.Printf("[FCM] Sending Grouped Notification to User %s: '%s' from Conv %s", 
			part.UserID, lastMsg.Content, p.ConversationID)
		
		// Real FCM logic would go here:
		// fcm.Send(Token, Content, ThreadID: p.ConversationID)
	}

	return nil
}
