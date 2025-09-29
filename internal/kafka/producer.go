package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"order-service/internal/models"

	"github.com/segmentio/kafka-go"
)

type Producer interface {
	SendOrderEvent(ctx context.Context, order *models.Order) error
	Close() error
}

type KafkaProducer struct {
	writer  *kafka.Writer
	brokers []string
	topic   string
}

func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	log.Printf("Kafka producer initialized with brokers: %v, topic: %s", brokers, topic)

	return &KafkaProducer{
		writer:  writer,
		brokers: brokers,
		topic:   topic,
	}
}

func (p *KafkaProducer) SendOrderEvent(ctx context.Context, order *models.Order) error {
	log.Printf("Attempting to send Kafka event for order %s to topic %s", order.ID, p.topic)

	event := map[string]interface{}{
		"event_type": "order_created",
		"order_id":   order.ID,
		"product":    order.Product,
		"quantity":   order.Quantity,
		"price":      order.Price,
		"timestamp":  order.CreatedAt,
	}

	message, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal Kafka event: %v", err)
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	log.Printf("Sending Kafka message: %s", string(message))

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Value: message,
	})

	if err != nil {
		log.Printf("FAILED to send Kafka event for order %s: %v", order.ID, err)
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}

	log.Printf("SUCCESS: Kafka event sent for order %s", order.ID)
	return nil
}

func (p *KafkaProducer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}
