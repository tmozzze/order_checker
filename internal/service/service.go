package service

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/tmozzze/order_checker/internal/cache"
	"github.com/tmozzze/order_checker/internal/models"
	"github.com/tmozzze/order_checker/internal/repository"
)

type OrderService struct {
	repo   *repository.OrderRepository
	cache  *cache.Cache
	writer *kafka.Writer
}

func NewOrderService(repo *repository.OrderRepository, cache *cache.Cache, writer *kafka.Writer) *OrderService {
	return &OrderService{repo: repo, cache: cache, writer: writer}
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*models.Order, error) {
	// Check cache
	if cached, ok := s.cache.Get(id); ok {
		return cached, nil
	}

	// Go to Postgres
	order, err := s.repo.GetOrderById(ctx, id)
	if err != nil {
		return nil, err
	}
	// Set cache
	s.cache.Set(order.OrderUID, order)
	return order, nil

}

func (s *OrderService) SaveOrder(ctx context.Context, order *models.Order) error {
	payload, err := json.Marshal(order)
	if err != nil {
		return err
	}

	err = s.writer.WriteMessages(ctx, kafka.Message{
		Value: payload,
	})
	if err != nil {
		log.Println("failde to write message:", err)
		return err
	}

	return nil
}
