package service

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

type OrderService struct {
	repo   *repository.OrderRepository
	cache  *cache.Cache
	writer *kafka.Writer
}

func NewOrderService(repo *repository.OrderRepository, cache *cache.Cache, writer *kafka.Writer) *OrderService {
	s := &OrderService{repo: repo, cache: cache, writer: writer}

	// preload orders from Postgres
	orders, err := repo.GetAllOrders(context.Background())
	if err != nil {
		log.Printf("failed to preload cache from DB: %v", err)
	} else {
		log.Printf("cache preloaded with %d orders", len(orders)) // Cache capacity = 100 now
	}

	return s
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*models.Order, error) {
	start := time.Now()
	// Check cache
	if val, ok := s.cache.Get(id); ok {
		order, _ := val.(*models.Order)
		log.Printf("[CACHE HIT] id=%s dur=%s", id, time.Since(start))
		return order, nil
	}

	// Go to Postgres
	order, err := s.repo.GetOrderById(ctx, id)
	if err != nil {
		return nil, err
	}
	// Set cache
	s.cache.Set(order.OrderUID, order)
	log.Printf("[DB FETCH] id=%s dur=%s (cached)", id, time.Since(start))
	return order, nil

}

func (s *OrderService) SaveOrder(ctx context.Context, order *models.Order) error {
	// Date
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Printf("failed to load location for Date: %v", err)
		order.DateCreated = time.Now().UTC()
	} else {
		order.DateCreated = time.Now().In(loc)
	}

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
