package message

import "github.com/google/uuid"

// EventType defines the type of WebSocket event
type EventType string

const (
	EventUserOnline  EventType = "USER_ONLINE"
	EventUserOffline EventType = "USER_OFFLINE"
	EventError       EventType = "ERROR"
)

// WSEvent is the top-level envelope for all WebSocket messages
type WSEvent struct {
	Type           EventType   `json:"type"`
	ConversationID *uuid.UUID  `json:"conversation_id,omitempty"`
	Payload        interface{} `json:"payload"`
}

// ErrorPayload is the payload for ERROR events
type ErrorPayload struct {
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}
