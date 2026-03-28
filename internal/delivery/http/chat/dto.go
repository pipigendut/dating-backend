package chat

import (
	"time"
	"strings"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
)

type ConversationResponse struct {
	ID          uuid.UUID                `json:"id"`
	Type        string                   `json:"type"` // "direct" or "group"
	Title       string                   `json:"title"`
	AvatarURL   string                   `json:"avatar_url"`
	LastMessage *MessageResponse         `json:"last_message,omitempty"`
	UnreadCount int                      `json:"unread_count"`
	IsTyping    bool                     `json:"is_typing"`
	CreatedAt   time.Time                `json:"created_at"`
	Entity      *response.EntityResponse `json:"entity,omitempty"`
}

type MessageResponse struct {
	ID             uuid.UUID                `json:"id"`
	ConversationID uuid.UUID                `json:"conversation_id"`
	SenderID       uuid.UUID                `json:"sender_id"`
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


func ToMessageResponse(m *entities.Message, currentUserID uuid.UUID) MessageResponse {
	isRead := m.Status == entities.MessageStatusRead
	return MessageResponse{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderID:       m.SenderID,
		Type:           m.Type,
		Content:        m.Content,
		Metadata:       m.Metadata,
		CreatedAt:      m.CreatedAt,
		IsRead:         isRead,
	}
}

type StorageURLProvider interface {
	GetPublicURL(key string) string
}

func ToConversationResponse(c *entities.Conversation, currentUserID uuid.UUID, unreadCount int, isTyping bool, storage StorageURLProvider) ConversationResponse {
	resp := ConversationResponse{
		ID:          c.ID,
		Type:        string(c.Type),
		UnreadCount: unreadCount,
		IsTyping:    isTyping,
		CreatedAt:   c.CreatedAt,
	}

	// For dating apps, the "Title" and "Avatar" are usually derived from the "Other" party
	// In group chats, it could be the Group Name.
	
	// If it's a direct chat, we find the other participant
	if c.Type == entities.ConversationTypeDirect {
		for _, p := range c.Participants {
			if p.UserID != currentUserID && p.User != nil {
				resp.Title = p.User.FullName
				if mainPhoto := p.User.GetMainPhotoProfile(); mainPhoto != nil {
					url := mainPhoto.URL
					if storage != nil && url != "" && !strings.HasPrefix(url, "http") {
						url = storage.GetPublicURL(url)
					}
					resp.AvatarURL = url
				}
				break
			}
		}
	} else if c.Type == entities.ConversationTypeGroup {
		// Group chat title comes from Group Name (would need Preload or separate fetch)
		resp.Title = "Group Chat" 
		// Extra: If matched to a specific Group entity, use its name.
	}

	if len(c.Messages) > 0 {
		m := ToMessageResponse(&c.Messages[0], currentUserID)
		resp.LastMessage = &m
	}

	// Populate Entity for UI convenience (e.g. unmatching)
	if c.Type == entities.ConversationTypeDirect {
		for _, p := range c.Participants {
			if p.UserID != currentUserID && p.User != nil {
				ur := response.ToUserLiteResponse(p.User, storage)
				resp.Entity = &response.EntityResponse{
					ID:   p.User.EntityID,
					Type: string(entities.EntityTypeUser),
					User: &ur,
				}
				break
			}
		}
	}

	return resp
}
