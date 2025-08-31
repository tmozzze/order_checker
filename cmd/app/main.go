package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/segmentio/kafka-go"
	"github.com/tmozzze/order_checker/internal/api"
	"github.com/tmozzze/order_checker/internal/cache"
	"github.com/tmozzze/order_checker/internal/config"
	"github.com/tmozzze/order_checker/internal/db"
	"github.com/tmozzze/order_checker/internal/kafka_consumer"
	"github.com/tmozzze/order_checker/internal/models"
	"github.com/tmozzze/order_checker/internal/repository"
	"github.com/tmozzze/order_checker/internal/service"
)

func main() {
	// Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Config error:", err)
	}

	ctx := context.Background()

	// Postgres
	database, err := db.NewDB(ctx, cfg)
	if err != nil {
		log.Fatal("Postgres init failed:", err)
	}
	defer database.Pool.Close()

	log.Println("Connected to Postgres on port", cfg.DBPort)

	// Repositories and cache
	repo := repository.NewOrderRepository(database.Pool)
	c := cache.New()

	// Kafka
	broker := "localhost:9092"
	topic := "orders"

	if err := kafka_consumer.EnsureTopic(broker, topic, 1, 1); err != nil {
		log.Fatal("failed to ensure topic:", err)
	}

	processedC := make(chan string)

	// Producer | Writer
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{broker},
		Topic:   topic,
	})
	defer writer.Close()

	// Consumer
	consumer := kafka_consumer.NewConsumer(
		[]string{broker},
		topic,
		"test-group",
		repo,
		c,
		processedC,
	)
	// Consumer in background
	go func() {
		log.Println("Starting kafka consumer...")
		if err := consumer.Start(ctx); err != nil {
			log.Fatal("consumer failed:", err)
		}
	}()

	// Service + Handlers
	svc := service.NewOrderService(repo, c, writer)
	h := api.NewOrderHandler(svc)

	// Routes
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	// Start HTTP Server on :8080
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Println("HTTP server started on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed:", err)
		}
	}()

	// Logs orders from consumer
	go func() {
		for orderUID := range processedC {
			log.Printf("Order proccesed: %s", orderUID)
		}
	}()

	select {} // Wait forever

}

func TestOrderSaveAndGet(ctx context.Context, repo *repository.OrderRepository, c *cache.Cache, writer *kafka.Writer, processedC chan string) {
	order := &models.Order{
		OrderUID:    "125",
		TrackNumber: "125",
		Entry:       "WBIL",
		Locale:      "en",
		CustomerID:  "test",
		DateCreated: time.Now(),
		Delivery: models.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Email:   "test@gmail.com",
		},
		Payment: models.Payment{
			Transaction:  "125",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
		},
		Items: []models.Item{
			{
				ChrtID:      125,
				TrackNumber: "125",
				Price:       453,
				RID:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				TotalPrice:  317,
				NmID:        125,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
	}

	payload, _ := json.Marshal(order)
	if err := writer.WriteMessages(ctx, kafka.Message{Value: payload}); err != nil {
		log.Fatal("producer failed:", err)
	}
	fmt.Println("Sent test order to Kafka")

	select {
	case <-ctx.Done():
		log.Fatal("context canceled before consumer processed")
	case uid := <-processedC:
		fmt.Println("Consumer processed order:", uid)
	}

	got, err := repo.GetOrderById(ctx, order.OrderUID)
	if err != nil {
		log.Fatal("failed to get order from db:", err)
	}
	fmt.Println("Got oreder from Postgres:", got.OrderUID)

	if cached, ok := c.Get(order.OrderUID); ok {
		fmt.Println("Order found in cache:", cached.OrderUID)
	} else {
		fmt.Println("Order not found in cache:", order.OrderUID)
	}

	fmt.Println("Order:\n", order)

}

/*
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "order_uid": "o-778",
    "track_number": "TN-778",
    "entry": "WBIL",
    "locale": "en",
    "customer_id": "user-124",
    "delivery": {
      "name": "Test User",
      "phone": "+9720000000",
      "city": "Tel Aviv",
      "address": "Some Street 1",
      "email": "user@test.com"
    },
    "payment": {
      "transaction": "trx-778",
      "currency": "USD",
      "provider": "visa",
      "amount": 100,
      "payment_dt": 1637907727,
      "bank": "alpha",
      "delivery_cost": 10,
      "goods_total": 90
    },
    "items": [
      {
        "chrt_id": 2,
        "track_number": "TN-778",
        "price": 100,
        "rid": "some-rid",
        "name": "Book",
        "sale": 0,
        "total_price": 100,
        "nm_id": 11,
        "brand": "NoName",
        "status": 200
      }
    ]
  }'

*/

/*
curl http://localhost:8080/orders/o-778
*/

/*
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "order_uid": "o-999",
    "track_number": "TN-999",
    "entry": "WBIL",
    "locale": "en",
    "customer_id": "user-999",
    "delivery": {
      "name": "Alice Example",
      "phone": "+9721111111",
      "city": "Haifa",
      "address": "Somewhere 42",
      "email": "alice@example.com"
    },
    "payment": {
      "transaction": "trx-999",
      "currency": "USD",
      "provider": "mastercard",
      "amount": 250,
      "payment_dt": 1637908888,
      "bank": "hapoalim",
      "delivery_cost": 15,
      "goods_total": 235,
      "custom_fee": 0
    },
    "items": [
      {
        "chrt_id": 999,
        "track_number": "TN-999",
        "price": 120,
        "rid": "rid-999a",
        "name": "Golang Book",
        "sale": 0,
        "size": "M",
        "total_price": 120,
        "nm_id": 9991,
        "brand": "TechPress",
        "status": 201
      },
      {
        "chrt_id": 1000,
        "track_number": "TN-999",
        "price": 115,
        "rid": "rid-999b",
        "name": "Kafka Guide",
        "sale": 0,
        "size": "L",
        "total_price": 115,
        "nm_id": 9992,
        "brand": "StreamBooks",
        "status": 201
      }
    ]
  }'

*/

/*
curl http://localhost:8080/orders/o-999

*/
