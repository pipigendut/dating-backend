package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type ChatServiceInterface interface {
	SendMessage(ctx context.Context, senderID, conversationID uuid.UUID, messageType entities.MessageType, content string) error
	SendTypingEvent(ctx context.Context, userID, conversationID uuid.UUID, isTyping bool) error
	SendReadReceipt(ctx context.Context, userID, conversationID, messageID uuid.UUID) error
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // For development
	},
}

type Client struct {
	hub         *Hub
	conn        *websocket.Conn
	send        chan []byte
	userID      uuid.UUID
	chatService ChatServiceInterface
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		
		var event WSEvent
		if err := json.Unmarshal(message, &event); err != nil {
			continue
		}
		
		ctx := context.Background()
		switch event.Type {
		case EventSendMessage:
			var payload SendMessagePayload
			payloadBytes, _ := json.Marshal(event.Payload)
			if err := json.Unmarshal(payloadBytes, &payload); err == nil {
				c.chatService.SendMessage(ctx, c.userID, *event.ConversationID, entities.MessageType(payload.MessageType), payload.Content)
			}
		case EventTypingStart:
			c.chatService.SendTypingEvent(ctx, c.userID, *event.ConversationID, true)
		case EventTypingStop:
			c.chatService.SendTypingEvent(ctx, c.userID, *event.ConversationID, false)
		case EventMessageRead:
			var payload ReadPayload
			payloadBytes, _ := json.Marshal(event.Payload)
			if err := json.Unmarshal(payloadBytes, &payload); err == nil {
				c.chatService.SendReadReceipt(ctx, c.userID, *event.ConversationID, payload.MessageID)
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWs(hub *Hub, chatSvc ChatServiceInterface, w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{
		hub:         hub,
		conn:        conn,
		send:        make(chan []byte, 256),
		userID:      userID,
		chatService: chatSvc,
	}
	log.Printf("[WebSocket] New connection established for userID: %s", userID)
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
