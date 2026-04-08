package message

import "github.com/google/uuid"

const (
	EventSendMessage    EventType = "SEND_MESSAGE"
	EventReceiveMessage EventType = "RECEIVE_MESSAGE"
	EventTypingStart    EventType = "TYPING_START"
	EventTypingStop     EventType = "TYPING_STOP"
	EventMessageRead    EventType = "MESSAGE_READ"
)

// SendMessagePayload is the payload for SEND_MESSAGE events (client → server)
type SendMessagePayload struct {
	Content     string `json:"content"`
	MessageType string `json:"message_type"`
}

// MessagePayload is the payload for RECEIVE_MESSAGE events (server → client)
type MessagePayload struct {
	ID             uuid.UUID `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderID       uuid.UUID `json:"sender_id"`
	Content        string    `json:"content"`
	Type           string    `json:"type"`
	CreatedAt      string    `json:"created_at"`
}

// TypingPayload is the payload for TYPING_START / TYPING_STOP events
type TypingPayload struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	UserID         uuid.UUID `json:"user_id"`
}

// ReadPayload is the payload for MESSAGE_READ events
type ReadPayload struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	MessageID      uuid.UUID `json:"message_id"`
	UserID         uuid.UUID `json:"user_id"`
}
