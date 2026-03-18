package impl

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) repository.ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) CreateConversation(ctx context.Context, conversation *entities.Conversation) error {
	return r.db.WithContext(ctx).Create(conversation).Error
}

func (r *chatRepository) GetConversationByID(ctx context.Context, id uuid.UUID) (*entities.Conversation, error) {
	var conv entities.Conversation
	err := r.db.WithContext(ctx).
		Preload("Participants.User.Photos").
		Preload("Participants.Presence").
		// Preload only the last message for the summary
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("messages.created_at DESC").Limit(1)
		}).
		Where("id = ?", id).
		First(&conv).Error
	return &conv, err
}

func (r *chatRepository) GetUserConversations(ctx context.Context, userID uuid.UUID) ([]entities.Conversation, error) {
	var convs []entities.Conversation
	err := r.db.WithContext(ctx).
		Preload("Participants.User.Photos").
		Preload("Participants.Presence").
		// Preload only the last message for the summary
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("messages.created_at DESC").Limit(1)
		}).
		Joins("JOIN conversation_participants cp ON cp.conversation_id = conversations.id").
		Where("cp.user_id = ? AND conversations.visible_at <= ?", userID, time.Now()).
		Order("conversations.last_message_at DESC").
		Find(&convs).Error
	return convs, err
}

func (r *chatRepository) GetConversationBetweenUsers(ctx context.Context, user1ID, user2ID uuid.UUID) (*entities.Conversation, error) {
	var conv entities.Conversation
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT c.* FROM conversations c
			JOIN conversation_participants cp1 ON c.id = cp1.conversation_id
			JOIN conversation_participants cp2 ON c.id = cp2.conversation_id
			WHERE cp1.user_id = ? AND cp2.user_id = ?
		`, user1ID, user2ID).
		First(&conv).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *chatRepository) GetUnreadCount(ctx context.Context, conversationID, userID uuid.UUID) (int, error) {
	var count int64
	// Simple query to count messages sent by others after last read
	err := r.db.WithContext(ctx).Table("messages m").
		Joins("JOIN conversation_participants cp ON cp.conversation_id = m.conversation_id").
		Where("cp.conversation_id = ? AND cp.user_id = ?", conversationID, userID).
		Where("m.sender_id != ?", userID).
		Where("m.created_at > (SELECT COALESCE(ms.created_at, '1970-01-01') FROM messages ms WHERE ms.id = cp.last_read_message_id)").
		Count(&count).Error
	return int(count), err
}

func (r *chatRepository) CreateMessage(ctx context.Context, message *entities.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *chatRepository) GetConversationMessages(ctx context.Context, conversationID uuid.UUID, limit int, offset int) ([]entities.Message, error) {
	var msgs []entities.Message
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&msgs).Error
	return msgs, err
}

func (r *chatRepository) MarkMessagesAsRead(ctx context.Context, conversationID, userID uuid.UUID, messageID uuid.UUID) error {
	// 1. Update Participant's LastReadMessageID (Scalable Approach)
	err := r.db.WithContext(ctx).Model(&entities.ConversationParticipant{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("last_read_message_id", messageID).Error
	if err != nil {
		return err
	}

	// 2. Also keep MessageRead for audit if needed, but primary check is above
	read := entities.MessageRead{
		ID:             uuid.New(),
		MessageID:      messageID,
		UserID:         userID,
		ConversationID: conversationID,
	}
	return r.db.WithContext(ctx).Create(&read).Error
}

func (r *chatRepository) UpdateConversationLastMessage(ctx context.Context, conversationID, messageID uuid.UUID, sentAt time.Time) error {
	return r.db.WithContext(ctx).Model(&entities.Conversation{}).
		Where("id = ?", conversationID).
		Updates(map[string]interface{}{
			"last_message_id": messageID,
			"last_message_at": sentAt,
		}).Error
}

func (r *chatRepository) UpdatePresence(ctx context.Context, presence *entities.UserPresence) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"is_online", "last_seen_at", "updated_at"}),
	}).Create(presence).Error
}

func (r *chatRepository) GetUserPresence(ctx context.Context, userID uuid.UUID) (*entities.UserPresence, error) {
	var presence entities.UserPresence
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&presence).Error
	return &presence, err
}
