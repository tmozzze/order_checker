package kafka_consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/tmozzze/order_checker/internal/cache"
	"github.com/tmozzze/order_checker/internal/models"
	"github.com/tmozzze/order_checker/internal/repository"
)

type Consumer struct {
	reader *kafka.Reader
	repo   *repository.OrderRepository
	cache  *cache.Cache
}

func NewConsumer(brokers []string, topic, groupID string, repo *repository.OrderRepository, c *cache.Cache) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           brokers,
		Topic:             topic,
		GroupID:           groupID,
		MinBytes:          1,
		MaxBytes:          10e6,
		HeartbeatInterval: 3_000_000_000,  //3s
		SessionTimeout:    30_000_000_000, //30s
		CommitInterval:    1_000_000_000,  //1s

	})

	return &Consumer{reader: r, repo: repo, cache: c}
}

func (c *Consumer) Start(ctx context.Context) error {
	log.Println("Kafka consumer started")

	// Exit if ctx cancel
	for {
		select {
		case <-ctx.Done():
			log.Println("Kafka consumer stopping...")
			if err := c.reader.Close(); err != nil {
				log.Printf("failed to close kafka reader: %v", err)
			}
			return nil
		default:
		}

		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return err
		}

		log.Printf("got message")

		// Parsing order
		var order models.Order
		if err := json.Unmarshal(m.Value, &order); err != nil {
			log.Printf("invalid message: %v", err)
			continue
		}

		// Saving to postgres
		if err := c.repo.SaveOrder(ctx, &order); err != nil {
			log.Printf("invalid message: %v", err)
			continue
		}

		// Cache
		c.cache.Set(order.OrderUID, &order)
		log.Printf("order cached: %s", order.OrderUID)

	}
}
