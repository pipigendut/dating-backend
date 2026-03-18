package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/chat/events"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
)

type Producer interface {
	Publish(ctx context.Context, topic string, key string, event events.EventEnvelope) error
}

type ChatService interface {
	// Actions (Kafka Producers)
	SendMessage(ctx context.Context, senderID, conversationID uuid.UUID, messageType entities.MessageType, content string, metadata events.MessageMetadata) error
	SendTypingEvent(ctx context.Context, userID, conversationID uuid.UUID, status string) error
	SendReadReceipt(ctx context.Context, userID, conversationID, messageID uuid.UUID) error
	UpdatePresence(ctx context.Context, userID uuid.UUID, isOnline bool) error

	// Queries
	GetConversations(ctx context.Context, userID uuid.UUID) ([]entities.Conversation, error)
	GetUnreadCount(ctx context.Context, conversationID, userID uuid.UUID) (int, error)
	GetMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]entities.Message, error)
	GetOrCreateConversation(ctx context.Context, user1ID, user2ID uuid.UUID, visibleAt time.Time) (*entities.Conversation, error)
}

type chatService struct {
	repo     repository.ChatRepository
	producer Producer
}

func NewChatService(repo repository.ChatRepository, producer Producer) ChatService {
	return &chatService{
		repo:     repo,
		producer: producer,
	}
}

func (s *chatService) SendMessage(ctx context.Context, senderID, conversationID uuid.UUID, messageType entities.MessageType, content string, metadata events.MessageMetadata) error {
	// 1. Get conversation to find receiver (simplified for 1:1)
	conv, err := s.repo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return err
	}

	var receiverID uuid.UUID
	for _, p := range conv.Participants {
		if p.UserID != senderID {
			receiverID = p.UserID
			break
		}
	}

	// 2. Wrap in Envelope
	event := events.EventEnvelope{
		EventID:   uuid.New(),
		EventType: events.EventTypeChatMessageSent,
		Version:   "v1",
		Timestamp: time.Now(),
		Producer:  "chat-service",
		Data: events.ChatMessageEvent{
			MessageID:      uuid.New(),
			ConversationID: conversationID,
			SenderID:       senderID,
			ReceiverID:     receiverID,
			MessageType:    string(messageType),
			Content:        content,
			Metadata:       metadata,
		},
	}

	// 3. Publish to Kafka (Ordering by conversation_id)
	return s.producer.Publish(ctx, "chat.messages", conversationID.String(), event)
}

func (s *chatService) SendTypingEvent(ctx context.Context, userID, conversationID uuid.UUID, status string) error {
	event := events.EventEnvelope{
		EventID:   uuid.New(),
		EventType: events.EventTypeChatTyping,
		Version:   "v1",
		Timestamp: time.Now(),
		Producer:  "chat-service",
		Data: events.TypingEvent{
			ConversationID: conversationID,
			UserID:         userID,
			Status:         status,
		},
	}
	return s.producer.Publish(ctx, "chat.typing", conversationID.String(), event)
}

func (s *chatService) SendReadReceipt(ctx context.Context, userID, conversationID, messageID uuid.UUID) error {
	event := events.EventEnvelope{
		EventID:   uuid.New(),
		EventType: events.EventTypeChatMessageRead,
		Version:   "v1",
		Timestamp: time.Now(),
		Producer:  "chat-service",
		Data: events.ReadReceiptEvent{
			MessageID:      messageID,
			ConversationID: conversationID,
			UserID:         userID,
			ReadAt:         time.Now(),
		},
	}
	return s.producer.Publish(ctx, "chat.read_receipts", conversationID.String(), event)
}

func (s *chatService) UpdatePresence(ctx context.Context, userID uuid.UUID, isOnline bool) error {
	event := events.EventEnvelope{
		EventID:   uuid.New(),
		EventType: events.EventTypeChatPresenceUpdate,
		Version:   "v1",
		Timestamp: time.Now(),
		Producer:  "chat-service",
		Data: events.PresenceEvent{
			UserID:     userID,
			IsOnline:   isOnline,
			LastSeenAt: time.Now(),
		},
	}
	return s.producer.Publish(ctx, "chat.presence", userID.String(), event)
}

func (s *chatService) GetConversations(ctx context.Context, userID uuid.UUID) ([]entities.Conversation, error) {
	return s.repo.GetUserConversations(ctx, userID)
}

func (s *chatService) GetUnreadCount(ctx context.Context, conversationID, userID uuid.UUID) (int, error) {
	return s.repo.GetUnreadCount(ctx, conversationID, userID)
}

func (s *chatService) GetMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]entities.Message, error) {
	return s.repo.GetConversationMessages(ctx, conversationID, limit, offset)
}

func (s *chatService) GetOrCreateConversation(ctx context.Context, user1ID, user2ID uuid.UUID, visibleAt time.Time) (*entities.Conversation, error) {
	conv, err := s.repo.GetConversationBetweenUsers(ctx, user1ID, user2ID)
	if err == nil {
		// If it exists but is not yet visible, maybe update its visibility if the new match is sooner?
		// For now, let's just return existing.
		return conv, nil
	}

	// Create new 1:1 conversation
	newConv := &entities.Conversation{
		ID:        uuid.New(),
		VisibleAt: visibleAt,
		Participants: []entities.ConversationParticipant{
			{ID: uuid.New(), UserID: user1ID},
			{ID: uuid.New(), UserID: user2ID},
		},
	}

	if err := s.repo.CreateConversation(ctx, newConv); err != nil {
		return nil, err
	}

	return newConv, nil
}
