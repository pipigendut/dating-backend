package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/infra/fcm"
	"github.com/pipigendut/dating-backend/internal/repository"
	"github.com/pipigendut/dating-backend/internal/services"
)

type NotificationWorker struct {
	chatRepo   repository.ChatRepository
	userRepo   repository.UserRepository
	redisRepo  repository.RedisRepository
	deviceRepo repository.DeviceRepository
	groupRepo  repository.GroupRepository
	notifRepo  repository.NotificationRepository
	fcmClient  *fcm.Client
}

func NewNotificationWorker(chatRepo repository.ChatRepository, userRepo repository.UserRepository, redisRepo repository.RedisRepository, deviceRepo repository.DeviceRepository, groupRepo repository.GroupRepository, notifRepo repository.NotificationRepository, fcmClient *fcm.Client) *NotificationWorker {
	return &NotificationWorker{
		chatRepo:   chatRepo,
		userRepo:   userRepo,
		redisRepo:  redisRepo,
		deviceRepo: deviceRepo,
		groupRepo:  groupRepo,
		notifRepo:  notifRepo,
		fcmClient:  fcmClient,
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
		
		// 4. Fetch all active devices for this user
		devices, err := w.deviceRepo.GetUserDevices(ctx, part.UserID)
		if err != nil {
			log.Printf("[Worker] Failed to fetch devices for user %s: %v", part.UserID, err)
			continue
		}

		var tokens []string
		for _, d := range devices {
			if d.FCMToken != nil && *d.FCMToken != "" {
				tokens = append(tokens, *d.FCMToken)
			}
		}

		if len(tokens) == 0 {
			continue
		}

		// 5. Send FCM
		title := "New Message"
		if conv.Type == entities.ConversationTypeGroup && conv.EntityID != nil {
			g, _ := w.groupRepo.GetGroupByEntityID(ctx, *conv.EntityID)
			if g != nil {
				title = fmt.Sprintf("%s (Group)", g.Name)
			} else {
				title = "Group Message"
			}
		} else {
			// Find the sender participant in this conversation
			sender, _ := w.userRepo.GetByID(lastMsg.SenderID)
			if sender != nil {
				title = sender.FullName
			}
		}

		data := map[string]string{
			"notification_type": "new_message",
			"conversation_id":   p.ConversationID.String(),
			"type":              string(conv.Type),
		}

		log.Printf("[FCM] Sending Grouped Notification to User %s: '%s' from Conv %s", 
			part.UserID, lastMsg.Content, p.ConversationID)
		
		if w.fcmClient != nil && w.canSendNotification(ctx, part.UserID, "new_message") {
			_ = w.fcmClient.SendMulticast(ctx, tokens, title, lastMsg.Content, data)
		}
	}

	return nil
}

func (w *NotificationWorker) canSendNotification(ctx context.Context, recipientID uuid.UUID, notifType string) bool {
	// 1. Check Global Setting
	globalSetting, err := w.notifRepo.GetGlobalSettingByType(ctx, notifType)
	if err != nil || globalSetting == nil || !globalSetting.IsEnable {
		return false
	}

	// 2. Check user's preference
	userSetting, err := w.notifRepo.GetUserSettingByType(ctx, recipientID, notifType)
	if err != nil {
		// If no record found, user asked to treat as 'false' (disabled) by default
		// in the GetUserSettings list, but for sending, we should respect that.
		// So if no record, is_user_enable is false.
		return false
	}

	return userSetting.IsEnable
}


// HandleMatchNotificationTask sends a "It's a Match!" notification to both matched users.
func (w *NotificationWorker) HandleMatchNotificationTask(ctx context.Context, t *asynq.Task) error {
	var p services.MatchNotificationPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("[Worker] Processing match notification for Match %s", p.MatchID)

	if w.fcmClient == nil {
		return nil
	}

	allUserIDs := w.resolveUserIDsFromEntity(ctx, p.SwiperEntityID)
	allUserIDs = append(allUserIDs, w.resolveUserIDsFromEntity(ctx, p.SwipedEntityID)...)

	seen := make(map[uuid.UUID]bool)
	for _, userID := range allUserIDs {
		if seen[userID] {
			continue
		}
		seen[userID] = true

		tokens := w.getTokensForUser(ctx, userID)
		if len(tokens) == 0 {
			continue
		}

		data := map[string]string{
			"notification_type": "new_match",
			"match_id":          p.MatchID.String(),
		}

		log.Printf("[FCM] Sending match notification to user %s", userID)
		if w.canSendNotification(ctx, userID, "new_match") {
			_ = w.fcmClient.SendMulticast(ctx, tokens, "💘 It's a Match!", "You have a new match! Start the conversation.", data)
		}
	}

	return nil
}

// HandleLikeNotificationTask sends a like or crush notification to the target user.
func (w *NotificationWorker) HandleLikeNotificationTask(ctx context.Context, t *asynq.Task) error {
	var p services.LikeNotificationPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("[Worker] Processing like notification: liker=%s target=%s isCrush=%v",
		p.LikerEntityID, p.TargetEntityID, p.IsCrush)

	if w.fcmClient == nil {
		return nil
	}

	targetUserIDs := w.resolveUserIDsFromEntity(ctx, p.TargetEntityID)
	if len(targetUserIDs) == 0 {
		return nil
	}

	title := "❤️ Someone liked you!"
	body := "Someone liked your profile. Go check them out!"
	notifType := "new_like"
	if p.IsCrush {
		title = "💘 You got a Crush!"
		body = "Someone sent you a Crush! They really like you."
		notifType = "new_crush"
	}

	for _, userID := range targetUserIDs {
		tokens := w.getTokensForUser(ctx, userID)
		if len(tokens) == 0 {
			continue
		}

		data := map[string]string{
			"notification_type": notifType,
		}

		log.Printf("[FCM] Sending %s notification to user %s", notifType, userID)
		if w.canSendNotification(ctx, userID, notifType) {
			_ = w.fcmClient.SendMulticast(ctx, tokens, title, body, data)
		}
	}

	return nil
}

// resolveUserIDsFromEntity returns the user IDs behind an entity ID.
// For user entities: returns the single user's ID.
// For group entities: returns all member user IDs.
func (w *NotificationWorker) resolveUserIDsFromEntity(ctx context.Context, entityID uuid.UUID) []uuid.UUID {
	// Try as user first
	user, err := w.userRepo.GetByEntityID(entityID)
	if err == nil && user != nil {
		return []uuid.UUID{user.ID}
	}

	// Try as group
	g, err := w.groupRepo.GetGroupByEntityID(ctx, entityID)
	if err != nil || g == nil {
		return nil
	}

	var userIDs []uuid.UUID
	for _, m := range g.Members {
		userIDs = append(userIDs, m.UserID)
	}
	return userIDs
}

// getTokensForUser fetches active FCM tokens for a given user.
func (w *NotificationWorker) getTokensForUser(ctx context.Context, userID uuid.UUID) []string {
	devices, err := w.deviceRepo.GetUserDevices(ctx, userID)
	if err != nil {
		return nil
	}
	var tokens []string
	for _, d := range devices {
		if d.FCMToken != nil && *d.FCMToken != "" {
			tokens = append(tokens, *d.FCMToken)
		}
	}
	return tokens
}
