package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/segmentio/kafka-go"
	"github.com/tmozzze/order_checker/internal/api"
	"github.com/tmozzze/order_checker/internal/cache"
	"github.com/tmozzze/order_checker/internal/config"
	"github.com/tmozzze/order_checker/internal/db"
	"github.com/tmozzze/order_checker/internal/kafka_consumer"
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

	// Repositories and cache(capacity: 100)
	repo := repository.NewOrderRepository(database.Pool)

	log.Println("Start preloading data in cache...")
	c := cache.New(100)

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

	// Router
	r := chi.NewRouter()

	// Frontend
	fs := http.FileServer(http.Dir("./web"))
	r.Handle("/*", fs)

	// API
	h.RegisterRoutes(r)

	// Start HTTP Server on :8080
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	log.Println("Starting HTTP server...")
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
