package message

const (
	EventNotification EventType = "NOTIFICATION"
)

// NotificationPayload is the payload for NOTIFICATION events
type NotificationPayload struct {
	Title   string      `json:"title"`
	Body    string      `json:"body"`
	Type    string      `json:"type"`
	Data    interface{} `json:"data,omitempty"`
}
