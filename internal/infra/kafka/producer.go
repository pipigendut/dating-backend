package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/pipigendut/dating-backend/internal/chat/events"
)

type Producer interface {
	Publish(ctx context.Context, topic string, key string, event events.EventEnvelope) error
	Close() error
}

type producer struct {
	writer *kafkago.Writer
}

func NewProducer(brokers []string) Producer {
	w := &kafkago.Writer{
		Addr:                   kafkago.TCP(brokers...),
		Balancer:               &kafkago.LeastBytes{},
		Async:                  true, // Async as requested
		AllowAutoTopicCreation: true,
	}

	return &producer{
		writer: w,
	}
}

func (p *producer) Publish(ctx context.Context, topic string, key string, event events.EventEnvelope) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafkago.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
	})

	if err != nil {
		log.Printf("[KAFKA] Failed to publish message to %s: %v", topic, err)
		return err
	}

	return nil
}

func (p *producer) Close() error {
	return p.writer.Close()
}
