package entities

import (
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	SoftDeleteModel
	LastMessageID *uuid.UUID     `gorm:"type:uuid;index"`
	LastMessageAt time.Time      `gorm:"index"`
	VisibleAt     time.Time      `gorm:"index"` // When this conversation becomes visible

	Participants []ConversationParticipant `gorm:"foreignKey:ConversationID"`
	Messages     []Message                 `gorm:"foreignKey:ConversationID"`
}

type ConversationParticipant struct {
	BaseModel
	ConversationID    uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_conv_user;index;not null"`
	UserID            uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_conv_user;index;not null"`
	LastReadMessageID *uuid.UUID `gorm:"type:uuid;index"` // Optimization for read receipts

	// Associations
	User     *User         `gorm:"foreignKey:UserID"`
	Presence *UserPresence `gorm:"foreignKey:UserID"`
}

type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeGif   MessageType = "gif"
)

type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
)

type Message struct {
	SoftDeleteModel
	ConversationID uuid.UUID      `gorm:"type:uuid;uniqueIndex:idx_conv_created;index;not null"`
	SenderID       uuid.UUID      `gorm:"type:uuid;index;not null"`
	Type           MessageType    `gorm:"type:varchar(20);not null"`
	Status         MessageStatus  `gorm:"type:varchar(20);default:'sent';index"`
	Content        string         `gorm:"type:text"` // For text messages or URLs
	ReplyToID      *uuid.UUID     `gorm:"type:uuid;index"`

	Metadata MessageMetadata `gorm:"type:jsonb"`
}

type MessageMetadata struct {
	GifProvider string `json:"gif_provider,omitempty"`
	ImageWidth  int    `json:"image_width,omitempty"`
	ImageHeight int    `json:"image_height,omitempty"`
}

type MessageRead struct {
	BaseModel
	MessageID      uuid.UUID `gorm:"type:uuid;index;not null"`
	UserID         uuid.UUID `gorm:"type:uuid;index;not null"`
	ConversationID uuid.UUID `gorm:"type:uuid;index;not null"`
}

type UserPresence struct {
	BaseModel
	UserID     uuid.UUID `gorm:"primaryKey;type:uuid"`
	IsOnline   bool      `gorm:"index"`
	LastSeenAt time.Time `gorm:"index"`
}
