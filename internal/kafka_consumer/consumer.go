package kafka_consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/tmozzze/order_checker/internal/cache"
	"github.com/tmozzze/order_checker/internal/models"
	"github.com/tmozzze/order_checker/internal/repository"
)

type Consumer struct {
	reader     *kafka.Reader
	repo       *repository.OrderRepository
	cache      *cache.Cache
	processedC chan string
}

func NewConsumer(brokers []string, topic, groupID string,
	repo *repository.OrderRepository, c *cache.Cache, processedC chan string) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           brokers,
		Topic:             topic,
		GroupID:           groupID,
		MinBytes:          1,
		MaxBytes:          10e6,
		HeartbeatInterval: 3 * time.Second,  //3s
		SessionTimeout:    30 * time.Second, //30s
		CommitInterval:    time.Second,      //1s
		StartOffset:       kafka.FirstOffset,
	})

	return &Consumer{reader: r, repo: repo, cache: c, processedC: processedC}
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

		// Read message
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			log.Printf("failed to fetch message: %v", err)
			time.Sleep(time.Second)
			continue
		}

		log.Printf("got message at topic/partition/offset %v/%v/%v", m.Topic, m.Partition, m.Offset)

		// Parsing order
		var order models.Order
		if err := json.Unmarshal(m.Value, &order); err != nil {
			log.Printf("failed to unmarshal order: %v", err)
			continue
		}

		// Saving to postgres
		if err := c.repo.SaveOrder(ctx, &order); err != nil {
			log.Printf("failed to save order: %v", err)
			continue
		}

		// Cache
		c.cache.Set(order.OrderUID, &order)
		c.processedC <- order.OrderUID
		log.Printf("order cached: %s", order.OrderUID)

		c.commitOffset(ctx, m)

	}
}

func (c *Consumer) commitOffset(ctx context.Context, m kafka.Message) {
	if err := c.reader.CommitMessages(ctx, m); err != nil {
		log.Printf("failed to commit message offset: %v", err)
	}
}
