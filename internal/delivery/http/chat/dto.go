package chat

import (
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/usecases"
)

type ConversationResponse struct {
	ID          uuid.UUID                  `json:"id"`
	User        ParticipantPreviewResponse `json:"user"`
	LastMessage *MessageResponse           `json:"last_message,omitempty"`
	UnreadCount int                        `json:"unread_count"`
	IsTyping    bool                       `json:"is_typing"`
	CreatedAt   time.Time                  `json:"created_at"`
}

type ParticipantPreviewResponse struct {
	ID               uuid.UUID  `json:"id"`
	FullName         string     `json:"full_name"`
	Age              int        `json:"age"`
	ProfilePicture   string     `json:"profile_picture"`
	OfficialVerified bool       `json:"official_verified"`
	VerifiedAt       *time.Time `json:"verified_at,omitempty"`
	IsOnline         bool       `json:"is_online"`
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

func ToConversationResponse(c *entities.Conversation, currentUserID uuid.UUID, unreadCount int, isTyping bool, storageUC *usecases.StorageUsecase) ConversationResponse {
	var otherUser *entities.ConversationParticipant
	for _, p := range c.Participants {
		if p.UserID != currentUserID {
			otherUser = &p
			break
		}
	}

	if otherUser == nil || otherUser.User == nil {
		return ConversationResponse{ID: c.ID, CreatedAt: c.CreatedAt}
	}

	fullName := otherUser.User.FullName
	photoURL := ""
	if mainPhoto := otherUser.User.GetMainPhotoProfile(); mainPhoto != nil {
		photoURL = mainPhoto.URL
		if storageUC != nil {
			photoURL = storageUC.GetPublicURL(mainPhoto.URL)
		}
	}

	isOnline := false
	if otherUser.Presence != nil {
		isOnline = otherUser.Presence.IsOnline
	}

	// Calculate age
	birthDate := otherUser.User.DateOfBirth
	age := time.Now().Year() - birthDate.Year()
	if time.Now().YearDay() < birthDate.YearDay() {
		age--
	}

	var lastMsg *MessageResponse
	if len(c.Messages) > 0 {
		m := ToMessageResponse(&c.Messages[0], currentUserID)
		lastMsg = &m
	}

	return ConversationResponse{
		ID: c.ID,
		User: ParticipantPreviewResponse{
			ID:               otherUser.UserID,
			FullName:         fullName,
			Age:              age,
			ProfilePicture:   photoURL,
			OfficialVerified: otherUser.User.VerifiedAt != nil,
			VerifiedAt:       otherUser.User.VerifiedAt,
			IsOnline:         isOnline,
		},
		LastMessage: lastMsg,
		UnreadCount: unreadCount,
		IsTyping:    isTyping,
		CreatedAt:   c.CreatedAt,
	}
}
