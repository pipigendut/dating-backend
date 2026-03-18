package kafka

import (
	"context"
	"encoding/json"
	"log"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/pipigendut/dating-backend/internal/chat/events"
	"github.com/pipigendut/dating-backend/internal/delivery/ws"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
)

type Consumer struct {
	reader  *kafkago.Reader
	repo    repository.ChatRepository
	wsMgr   *ws.Manager
}

func NewConsumer(brokers []string, groupID string, topic string, repo repository.ChatRepository, wsMgr *ws.Manager) *Consumer {
	r := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &Consumer{
		reader: r,
		repo:   repo,
		wsMgr:  wsMgr,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("[KAFKA] Consumer error: %v", err)
			break
		}

		var envelope events.EventEnvelope
		if err := json.Unmarshal(m.Value, &envelope); err != nil {
			log.Printf("[KAFKA] Failed to unmarshal envelope: %v", err)
			continue
		}

		c.handleEvent(ctx, envelope)
	}
}

func (c *Consumer) handleEvent(ctx context.Context, envelope events.EventEnvelope) {
	// Re-marshal data into its specific type based on EventType
	dataBytes, _ := json.Marshal(envelope.Data)

	switch envelope.EventType {
	case events.EventTypeChatMessageSent:
		var event events.ChatMessageEvent
		json.Unmarshal(dataBytes, &event)
		
		// 1. Persist to DB
		msg := &entities.Message{
			ID:             event.MessageID,
			ConversationID: event.ConversationID,
			SenderID:       event.SenderID,
			Type:           entities.MessageType(event.MessageType),
			Content:        event.Content,
			CreatedAt:      envelope.Timestamp,
			Metadata: entities.MessageMetadata{
				GifProvider: event.Metadata.GifProvider,
				ImageWidth:  event.Metadata.ImageWidth,
				ImageHeight: event.Metadata.ImageHeight,
			},
		}
		if err := c.repo.CreateMessage(ctx, msg); err != nil {
			log.Printf("[CHAT] Failed to persist message: %v", err)
			return
		}

		// 2. Update Conversation Last Message for efficient listing/sorting
		if err := c.repo.UpdateConversationLastMessage(ctx, event.ConversationID, msg.ID, msg.CreatedAt); err != nil {
			log.Printf("[CHAT] Failed to update conversation last message: %v", err)
		}

		// 3. Deliver to Receiver via WS if online
		c.wsMgr.SendToUser(event.ReceiverID, dataBytes)

	case events.EventTypeChatTyping:
		var event events.TypingEvent
		json.Unmarshal(dataBytes, &event)
		
		// Typing events are transient, only forward via WS
		// (In a real app, find recipient IDs for the conversation)
		// For simplicity, we assume we know who to broadcast to or broadcast to participants
		participants, _ := c.repo.GetConversationByID(ctx, event.ConversationID)
		if participants != nil {
			for _, p := range participants.Participants {
				if p.UserID != event.UserID {
					c.wsMgr.SendToUser(p.UserID, dataBytes)
				}
			}
		}

	case events.EventTypeChatMessageRead:
		var event events.ReadReceiptEvent
		json.Unmarshal(dataBytes, &event)
		
		// 1. Persist
		c.repo.MarkMessagesAsRead(ctx, event.ConversationID, event.UserID, event.MessageID)

		// 2. Notify sender
		msg, _ := c.repo.GetConversationByID(ctx, event.ConversationID)
		if msg != nil {
			for _, p := range msg.Participants {
				if p.UserID != event.UserID {
					c.wsMgr.SendToUser(p.UserID, dataBytes)
				}
			}
		}

	case events.EventTypeChatPresenceUpdate:
		var event events.PresenceEvent
		json.Unmarshal(dataBytes, &event)
		
		// 1. Persist
		c.repo.UpdatePresence(ctx, &entities.UserPresence{
			UserID:     event.UserID,
			IsOnline:   event.IsOnline,
			LastSeenAt: event.LastSeenAt,
		})
		
		// Presence could be broadcasted to all active friends/matches if needed
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
