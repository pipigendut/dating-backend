package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type ChatRepository interface {
	// Conversations
	CreateConversation(ctx context.Context, conversation *entities.Conversation) error
	GetConversationByID(ctx context.Context, id uuid.UUID) (*entities.Conversation, error)
	GetUserConversations(ctx context.Context, userID uuid.UUID, limit int, cursor *time.Time) ([]entities.Conversation, error)
	GetConversationBetweenUsers(ctx context.Context, user1ID, user2ID uuid.UUID) (*entities.Conversation, error)
	GetUnreadCount(ctx context.Context, conversationID, userID uuid.UUID) (int, error)

	// Messages
	CreateMessage(ctx context.Context, message *entities.Message) error
	GetConversationMessages(ctx context.Context, conversationID uuid.UUID, limit int, offset int) ([]entities.Message, error)
	MarkMessagesAsRead(ctx context.Context, conversationID, userID uuid.UUID, messageID uuid.UUID) error
	UpdateConversationLastMessage(ctx context.Context, conversationID, messageID uuid.UUID, sentAt time.Time) error

	// Presence
	UpdatePresence(ctx context.Context, presence *entities.UserPresence) error
	GetUserPresence(ctx context.Context, userID uuid.UUID) (*entities.UserPresence, error)
}
