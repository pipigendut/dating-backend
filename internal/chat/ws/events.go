package ws

import (
	"github.com/google/uuid"
)

type EventType string

const (
	EventSendMessage   EventType = "SEND_MESSAGE"
	EventReceiveMessage EventType = "RECEIVE_MESSAGE"
	EventTypingStart   EventType = "TYPING_START"
	EventTypingStop    EventType = "TYPING_STOP"
	EventMessageRead   EventType = "MESSAGE_READ"
	EventUserOnline    EventType = "USER_ONLINE"
	EventUserOffline   EventType = "USER_OFFLINE"
	EventError         EventType = "ERROR"
)

type WSEvent struct {
	Type           EventType   `json:"type"`
	ConversationID *uuid.UUID  `json:"conversation_id,omitempty"`
	Payload        interface{} `json:"payload"`
}

type SendMessagePayload struct {
	Content     string `json:"content"`
	MessageType string `json:"message_type"`
}

type MessagePayload struct {
	ID             uuid.UUID   `json:"id"`
	ConversationID uuid.UUID   `json:"conversation_id"`
	SenderID       uuid.UUID   `json:"sender_id"`
	Content        string      `json:"content"`
	Type           string      `json:"type"`
	CreatedAt      string      `json:"created_at"`
}

type TypingPayload struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	UserID         uuid.UUID `json:"user_id"`
}

type ReadPayload struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	MessageID      uuid.UUID `json:"message_id"`
	UserID         uuid.UUID `json:"user_id"`
}
