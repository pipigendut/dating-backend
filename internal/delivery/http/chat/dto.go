package chat

import (
	"time"
	"strings"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
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
	Entity         *response.EntityResponse `json:"entity,omitempty"`
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

func ToMessageResponse(m *entities.Message, currentUserID uuid.UUID, storage StorageURLProvider) MessageResponse {
	isRead := m.Status == entities.MessageStatusRead
	resp := MessageResponse{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderID:       m.SenderID,
		Type:           m.Type,
		Content:        m.Content,
		Metadata:       m.Metadata,
		CreatedAt:      m.CreatedAt,
		IsRead:         isRead,
	}

	if m.Sender != nil {
		resp.SenderName = m.Sender.FullName
		if mainPhoto := m.Sender.GetMainPhotoProfile(); mainPhoto != nil {
			url := mainPhoto.URL
			if storage != nil && url != "" && !strings.HasPrefix(url, "http") {
				url = storage.GetPublicURL(url)
			}
			resp.SenderPhotoURL = url
		}
	}

	return resp
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

	var swiperEntityID uuid.UUID
	var otherEntityID uuid.UUID
	var otherEntity *entities.Entity

	// 1. Identify which entity the current user belongs to
	if c.Match != nil {
		// Identify if currentUserID is in Entity1
		isUserInE1 := false
		if c.Match.Entity1 != nil {
			if c.Match.Entity1.Type == entities.EntityTypeUser && c.Match.Entity1.User != nil && c.Match.Entity1.User.ID == currentUserID {
				isUserInE1 = true
			} else if c.Match.Entity1.Type == entities.EntityTypeGroup && c.Match.Entity1.Group != nil {
				for _, m := range c.Match.Entity1.Group.Members {
					if m.UserID == currentUserID {
						isUserInE1 = true
						break
					}
				}
			}
		}

		if isUserInE1 {
			swiperEntityID = c.Match.Entity1ID
			otherEntityID = c.Match.Entity2ID
			otherEntity = c.Match.Entity2
		} else {
			swiperEntityID = c.Match.Entity2ID
			otherEntityID = c.Match.Entity1ID
			otherEntity = c.Match.Entity1
		}
	}
	resp.SwiperEntityID = swiperEntityID

	// 2. Identify participants belonging to the OTHER side
	var targetParticipants []entities.ConversationParticipant
	for _, p := range c.Participants {
		if p.UserID == currentUserID || p.User == nil {
			continue
		}
		
		// In a squad match, we only want members of the other entity
		if otherEntity != nil {
			isTarget := false
			if otherEntity.Type == entities.EntityTypeUser && otherEntity.User != nil && otherEntity.User.ID == p.UserID {
				isTarget = true
			} else if otherEntity.Type == entities.EntityTypeGroup && otherEntity.Group != nil {
				for _, m := range otherEntity.Group.Members {
					if m.UserID == p.UserID {
						isTarget = true
						break
					}
				}
			}
			if isTarget {
				targetParticipants = append(targetParticipants, p)
			}
		} else {
			// Fallback for legacy chats without Match
			targetParticipants = append(targetParticipants, p)
		}
	}

	// 3. Populate metadata (Title, Photos) from targetParticipants and otherEntity
	if c.Type == entities.ConversationTypeDirect && len(targetParticipants) > 0 {
		p := targetParticipants[0]
		resp.Title = p.User.FullName
		if mainPhoto := p.User.GetMainPhotoProfile(); mainPhoto != nil {
			url := mainPhoto.URL
			if storage != nil && url != "" && !strings.HasPrefix(url, "http") {
				url = storage.GetPublicURL(url)
			}
			resp.AvatarURL = url
		}
	} else if c.Type == entities.ConversationTypeGroup {
		// Group chat title prioritizes Group Name
		if otherEntity != nil && otherEntity.Type == entities.EntityTypeGroup && otherEntity.Group != nil {
			resp.Title = otherEntity.Group.Name
		}
		
		var names []string
		for _, p := range targetParticipants {
			names = append(names, strings.Split(p.User.FullName, " ")[0])
			if mainPhoto := p.User.GetMainPhotoProfile(); mainPhoto != nil {
				url := mainPhoto.URL
				if storage != nil && url != "" && !strings.HasPrefix(url, "http") {
					url = storage.GetPublicURL(url)
				}
				resp.AvatarURLs = append(resp.AvatarURLs, url)
			}
		}

		if resp.Title == "" {
			if len(names) > 0 {
				resp.Title = strings.Join(names, ", ")
			} else {
				resp.Title = "Squad Chat"
			}
		}
	}

	// 4. Set Entity object for UI (unmatch/reporting)
	if otherEntityID != uuid.Nil {
		resp.Entity = &response.EntityResponse{
			ID: otherEntityID,
		}
		if otherEntity != nil {
			resp.Entity.Type = string(otherEntity.Type)
			if otherEntity.Type == entities.EntityTypeUser && otherEntity.User != nil {
				ur := response.ToUserLiteResponse(otherEntity.User, storage)
				resp.Entity.User = &ur
			} else if otherEntity.Type == entities.EntityTypeGroup && otherEntity.Group != nil {
				gr := response.ToGroupResponse(otherEntity.Group, storage)
				resp.Entity.Group = &gr
			}
		} else {
			// Fallback type if entity not preloaded
			if c.Type == entities.ConversationTypeGroup {
				resp.Entity.Type = string(entities.EntityTypeGroup)
			} else {
				resp.Entity.Type = string(entities.EntityTypeUser)
			}
		}
	}

	if len(c.Messages) > 0 {
		m := ToMessageResponse(&c.Messages[0], currentUserID, storage)
		resp.LastMessage = &m
	}

	return resp
}
