package dtov1

import (
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type ConversationResponse struct {
	ID             uuid.UUID                `json:"id"`
	Type           string                   `json:"type"` // "direct" or "group"
	Title          string                   `json:"title"`
	AvatarURL      string                   `json:"avatar_url"`
	AvatarURLs     []string                 `json:"avatar_urls,omitempty"`
	SwiperEntityID uuid.UUID                `json:"swiper_entity_id"`
	LastMessage    *MessageResponse         `json:"last_message,omitempty"`
	UnreadCount    int                      `json:"unread_count"`
	IsTyping       bool                     `json:"is_typing"`
	CreatedAt      time.Time                `json:"created_at"`
	Entity         *EntityResponse          `json:"entity,omitempty"`
}

type MessageResponse struct {
	ID             uuid.UUID                `json:"id"`
	ConversationID uuid.UUID                `json:"conversation_id"`
	SenderID       uuid.UUID                `json:"sender_id"`
	SenderName     string                   `json:"sender_name,omitempty"`
	SenderPhotoURL string                   `json:"sender_photo_url,omitempty"`
	Type           entities.MessageType     `json:"type"`
	Content        string                   `json:"content"`
	Metadata       entities.MessageMetadata `json:"metadata"`
	CreatedAt      time.Time                `json:"created_at"`
	IsRead         bool                     `json:"is_read"`
}

type ChatUploadURLResponse struct {
	UploadURL string `json:"upload_url"`
	FileKey   string `json:"file_key"`
}
