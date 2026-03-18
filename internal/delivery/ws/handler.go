package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/services"
	"github.com/pipigendut/dating-backend/internal/chat/events"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512KB for images/gifs
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, restrict this
	},
}

type Handler struct {
	manager     *Manager
	chatService services.ChatService
}

func NewHandler(manager *Manager, chatService services.ChatService) *Handler {
	return &Handler{
		manager:     manager,
		chatService: chatService,
	}
}

func (h *Handler) HandleWebSocket(c *gin.Context) {
	userIDStr := c.Query("user_id") // In reality, get from JWT middleware
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to websocket: %v", err)
		return
	}

	client := &Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	h.manager.Register <- client

	// Start pumps
	go h.writePump(client)
	go h.readPump(client)

	// Update presence (Online)
	h.chatService.UpdatePresence(c.Request.Context(), userID, true)
}

func (h *Handler) readPump(c *Client) {
	defer func() {
		h.manager.Unregister <- c
		c.Conn.Close()
		// Update presence (Offline)
		h.chatService.UpdatePresence(context.Background(), c.UserID, false)
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var incoming struct {
			Type           string      `json:"type"`
			ConversationID uuid.UUID   `json:"conversation_id"`
			Content        string      `json:"content"`
			MessageType    string      `json:"message_type"`
			MessageID      uuid.UUID   `json:"message_id"`
			Status         string      `json:"status"`
			Metadata       interface{} `json:"metadata"`
		}

		if err := json.Unmarshal(message, &incoming); err != nil {
			continue
		}

		ctx := context.Background()

		switch incoming.Type {
		case "chat_message":
			var meta events.MessageMetadata
			if incoming.Metadata != nil {
				metaBytes, _ := json.Marshal(incoming.Metadata)
				json.Unmarshal(metaBytes, &meta)
			}
			h.chatService.SendMessage(ctx, c.UserID, incoming.ConversationID, entities.MessageType(incoming.MessageType), incoming.Content, meta)

		case "typing":
			h.chatService.SendTypingEvent(ctx, c.UserID, incoming.ConversationID, incoming.Status)

		case "read_receipt":
			h.chatService.SendReadReceipt(ctx, c.UserID, incoming.ConversationID, incoming.MessageID)
		}
	}
}

func (h *Handler) writePump(c *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
