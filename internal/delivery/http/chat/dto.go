package chat

import (
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/usecases"
)

type ConversationResponse struct {
	ID           uuid.UUID            `json:"id"`
	Participants []ParticipantResponse `json:"participants"`
	LastMessage  *MessageResponse     `json:"last_message,omitempty"`
	UnreadCount  int                  `json:"unread_count"`
	CreatedAt    time.Time            `json:"created_at"`

	// User Preview for Chat List
	OtherUser *ParticipantPreviewResponse `json:"other_user,omitempty"`
}

type ParticipantPreviewResponse struct {
	ID             uuid.UUID `json:"id"`
	FullName       string    `json:"full_name"`
	Age            int       `json:"age"`
	ProfilePicture string    `json:"profile_picture"`
	IsOnline       bool      `json:"is_online"`
}

type ParticipantResponse struct {
	UserID   uuid.UUID `json:"user_id"`
	FullName string    `json:"full_name"`
	PhotoURL string    `json:"photo_url"`
	IsOnline bool      `json:"is_online"`
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

func ToConversationResponse(c *entities.Conversation, currentUserID uuid.UUID, unreadCount int, storageUC *usecases.StorageUsecase) ConversationResponse {
	participants := make([]ParticipantResponse, 0, len(c.Participants))
	for _, p := range c.Participants {
		fullName := ""
		photoURL := ""
		isOnline := false

		if p.User != nil {
			fullName = p.User.FullName
			if mainPhoto := p.User.GetMainPhotoProfile(); mainPhoto != nil {
				photoURL = mainPhoto.URL
				if storageUC != nil {
					photoURL = storageUC.GetPublicURL(mainPhoto.URL)
				}
			}
		}

		if p.Presence != nil {
			isOnline = p.Presence.IsOnline
		}

		participants = append(participants, ParticipantResponse{
			UserID:   p.UserID,
			FullName: fullName,
			PhotoURL: photoURL,
			IsOnline: isOnline,
		})
	}

	var lastMsg *MessageResponse
	if len(c.Messages) > 0 {
		m := ToMessageResponse(&c.Messages[0], currentUserID)
		lastMsg = &m
	}

	var otherUser *ParticipantPreviewResponse
	for _, p := range participants {
		if p.UserID != currentUserID {
			// Find original participant to get additional info like age
			var age int
			for _, cp := range c.Participants {
				if cp.UserID == p.UserID && cp.User != nil {
					// Basic age calculation
					birthDate := cp.User.DateOfBirth
					age = time.Now().Year() - birthDate.Year()
					if time.Now().YearDay() < birthDate.YearDay() {
						age--
					}
					break
				}
			}

			otherUser = &ParticipantPreviewResponse{
				ID:             p.UserID,
				FullName:       p.FullName,
				Age:            age,
				ProfilePicture: p.PhotoURL,
				IsOnline:       p.IsOnline,
			}
			break
		}
	}

	return ConversationResponse{
		ID:           c.ID,
		Participants: participants,
		LastMessage:  lastMsg,
		UnreadCount:  unreadCount,
		OtherUser:    otherUser,
		CreatedAt:    c.CreatedAt,
	}
}
