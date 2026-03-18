package events

import (
	"time"

	"github.com/google/uuid"
)

// EventEnvelope is the standard wrapper for all Kafka events in the system.
type EventEnvelope struct {
	EventID   uuid.UUID   `json:"event_id"`
	EventType string      `json:"event_type"`
	Version   string      `json:"version"`
	Timestamp time.Time   `json:"timestamp"`
	Producer  string      `json:"producer"`
	Data      interface{} `json:"data"`
}

// Event Types constants
const (
	EventTypeChatMessageSent    = "chat.message.sent"
	EventTypeChatTyping         = "chat.typing"
	EventTypeChatPresenceUpdate = "chat.presence.updated"
	EventTypeChatMessageRead    = "chat.message.read"
)

// --- Event Data Structs ---

type ChatMessageEvent struct {
	MessageID      uuid.UUID       `json:"message_id"`
	ConversationID uuid.UUID       `json:"conversation_id"`
	SenderID       uuid.UUID       `json:"sender_id"`
	ReceiverID     uuid.UUID       `json:"receiver_id"`
	MessageType    string          `json:"message_type"`
	Content        string          `json:"content,omitempty"`
	MediaURL       string          `json:"media_url,omitempty"`
	Metadata       MessageMetadata `json:"metadata"`
}

type MessageMetadata struct {
	GifProvider string `json:"gif_provider,omitempty"`
	ImageWidth  int    `json:"image_width,omitempty"`
	ImageHeight int    `json:"image_height,omitempty"`
}

type TypingEvent struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	UserID         uuid.UUID `json:"user_id"`
	Status         string    `json:"status"` // "start" | "stop"
}

type PresenceEvent struct {
	UserID     uuid.UUID `json:"user_id"`
	IsOnline   bool      `json:"is_online"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

type ReadReceiptEvent struct {
	MessageID      uuid.UUID `json:"message_id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	UserID         uuid.UUID `json:"user_id"`
	ReadAt         time.Time `json:"read_at"`
}
