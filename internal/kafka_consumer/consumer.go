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
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})

	return &Consumer{reader: r, repo: repo, cache: c}
}

func (c *Consumer) Start(ctx context.Context) error {
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return err
		}

		var order models.Order
		if err := json.Unmarshal(m.Value, &order); err != nil {
			log.Printf("invalid message: %v", err)
			continue
		}

		if err := c.repo.SaveOrder(ctx, &order); err != nil {
			log.Printf("invalid message: %v", err)
			continue
		}

		c.cache.Set(order.OrderUID, &order)
		log.Printf("order %s processed", order.OrderUID)

	}
}
