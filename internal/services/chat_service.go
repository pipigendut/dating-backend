package services

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/chat/ws"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
)

type ChatService interface {
	SendMessage(ctx context.Context, senderID, conversationID uuid.UUID, messageType entities.MessageType, content string) error
	SendTypingEvent(ctx context.Context, userID, conversationID uuid.UUID, isTyping bool) error
	SendReadReceipt(ctx context.Context, userID, conversationID, messageID uuid.UUID) error

	// Queries
	GetConversations(ctx context.Context, userID uuid.UUID, limit int, cursor *time.Time) ([]entities.Conversation, error)
	GetUnreadCount(ctx context.Context, userID, conversationID uuid.UUID) (int, error)
	IsTyping(ctx context.Context, conversationID, userID uuid.UUID) (bool, error)
	GetMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]entities.Message, error)
	GetOrCreateConversation(ctx context.Context, user1ID, user2ID uuid.UUID, visibleAt time.Time) (*entities.Conversation, error)
}

type chatService struct {
	repo      repository.ChatRepository
	redisRepo repository.RedisRepository
	hub       *ws.Hub
}

func NewChatService(repo repository.ChatRepository, redisRepo repository.RedisRepository, hub *ws.Hub) ChatService {
	return &chatService{
		repo:      repo,
		redisRepo: redisRepo,
		hub:       hub,
	}
}

func (s *chatService) SendMessage(ctx context.Context, senderID, conversationID uuid.UUID, messageType entities.MessageType, content string) error {
	// 1. Save to DB
	msg := &entities.Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		Type:           messageType,
		Content:        content,
		Status:         entities.MessageStatusSent,
	}
	msg.ID = uuid.New()
	log.Printf("[ChatService] Saving message from %s to conversation %s", senderID, conversationID)
	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return err
	}

	// 2. Update conversation last message
	s.repo.UpdateConversationLastMessage(ctx, conversationID, msg.ID, time.Now())

	// 3. Find receiver
	conv, _ := s.repo.GetConversationByID(ctx, conversationID)
	var receiverID uuid.UUID
	for _, p := range conv.Participants {
		if p.UserID != senderID {
			receiverID = p.UserID
			break
		}
	}

	// 4. Redis: Increment unread count for receiver
	s.redisRepo.IncrementUnreadCount(ctx, receiverID, conversationID)

	// 5. Broadcast to Hub (Real-time)
	event := ws.WSEvent{
		Type:           ws.EventReceiveMessage,
		ConversationID: &conversationID,
		Payload: ws.MessagePayload{
			ID:             msg.ID,
			ConversationID: conversationID,
			SenderID:       senderID,
			Content:        content,
			Type:           string(messageType),
			CreatedAt:      msg.CreatedAt.Format(time.RFC3339),
		},
	}

	// Publish to Redis Pub/Sub for cross-instance scaling
	// We broadcast to both sender and receiver to ensure UI synchronization
	for _, targetID := range []uuid.UUID{senderID, receiverID} {
		s.redisRepo.PublishEvent(ctx, "chat:events", struct {
			TargetUserID uuid.UUID  `json:"target_user_id"`
			Event        ws.WSEvent `json:"event"`
		}{
			TargetUserID: targetID,
			Event:        event,
		})
		s.hub.BroadcastEvent(event, targetID)
	}

	return nil
}

func (s *chatService) SendTypingEvent(ctx context.Context, userID, conversationID uuid.UUID, isTyping bool) error {
	// 1. Update Redis
	s.redisRepo.SetTyping(ctx, conversationID, userID, isTyping)

	// 2. Find receiver to notify
	conv, _ := s.repo.GetConversationByID(ctx, conversationID)
	var receiverID uuid.UUID
	for _, p := range conv.Participants {
		if p.UserID != userID {
			receiverID = p.UserID
			break
		}
	}

	eventType := ws.EventTypingStart
	if !isTyping {
		eventType = ws.EventTypingStop
	}

	event := ws.WSEvent{
		Type:           eventType,
		ConversationID: &conversationID,
		Payload: ws.TypingPayload{
			ConversationID: conversationID,
			UserID:         userID,
		},
	}

	s.redisRepo.PublishEvent(ctx, "chat:events", struct {
		TargetUserID uuid.UUID  `json:"target_user_id"`
		Event        ws.WSEvent `json:"event"`
	}{
		TargetUserID: receiverID,
		Event:        event,
	})
	s.hub.BroadcastEvent(event, receiverID)

	return nil
}

func (s *chatService) SendReadReceipt(ctx context.Context, userID, conversationID, messageID uuid.UUID) error {
	// 1. Update DB
	if err := s.repo.MarkMessagesAsRead(ctx, conversationID, userID, messageID); err != nil {
		return err
	}

	// 2. Reset Redis unread count
	s.redisRepo.ResetUnreadCount(ctx, userID, conversationID)

	// 3. Find other participant and broadcast READ event
	conv, _ := s.repo.GetConversationByID(ctx, conversationID)
	var otherID uuid.UUID
	for _, p := range conv.Participants {
		if p.UserID != userID {
			otherID = p.UserID
			break
		}
	}

	event := ws.WSEvent{
		Type:           ws.EventMessageRead,
		ConversationID: &conversationID,
		Payload: ws.ReadPayload{
			ConversationID: conversationID,
			UserID:         userID,
			MessageID:      messageID,
		},
	}
	s.redisRepo.PublishEvent(ctx, "chat:events", struct {
		TargetUserID uuid.UUID  `json:"target_user_id"`
		Event        ws.WSEvent `json:"event"`
	}{
		TargetUserID: otherID,
		Event:        event,
	})
	s.hub.BroadcastEvent(event, otherID)

	return nil
}

func (s *chatService) GetConversations(ctx context.Context, userID uuid.UUID, limit int, cursor *time.Time) ([]entities.Conversation, error) {
	return s.repo.GetUserConversations(ctx, userID, limit, cursor)
}

func (s *chatService) GetUnreadCount(ctx context.Context, userID, conversationID uuid.UUID) (int, error) {
	// Try Redis first
	count, err := s.redisRepo.GetUnreadCount(ctx, userID, conversationID)
	if err == nil && count > 0 {
		return count, nil
	}
	// Fallback to DB
	return s.repo.GetUnreadCount(ctx, conversationID, userID)
}

func (s *chatService) IsTyping(ctx context.Context, conversationID, userID uuid.UUID) (bool, error) {
	return s.redisRepo.IsTyping(ctx, conversationID, userID)
}

func (s *chatService) GetMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]entities.Message, error) {
	return s.repo.GetConversationMessages(ctx, conversationID, limit, offset)
}

func (s *chatService) GetOrCreateConversation(ctx context.Context, user1ID, user2ID uuid.UUID, visibleAt time.Time) (*entities.Conversation, error) {
	conv, err := s.repo.GetConversationBetweenUsers(ctx, user1ID, user2ID)
	if err == nil {
		return conv, nil
	}

	newConv := entities.Conversation{
		VisibleAt: visibleAt,
		Participants: []entities.ConversationParticipant{
			{UserID: user1ID},
			{UserID: user2ID},
		},
	}
	newConv.ID = uuid.New()

	if err := s.repo.CreateConversation(ctx, &newConv); err != nil {
		return nil, err
	}

	return &newConv, nil
}
