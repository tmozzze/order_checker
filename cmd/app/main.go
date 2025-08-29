package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/tmozzze/order_checker/internal/cache"
	"github.com/tmozzze/order_checker/internal/config"
	"github.com/tmozzze/order_checker/internal/db"
	"github.com/tmozzze/order_checker/internal/kafka_consumer"
	"github.com/tmozzze/order_checker/internal/models"
	"github.com/tmozzze/order_checker/internal/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Config error:", err)
	}

	ctx := context.Background()

	database, err := db.NewDB(ctx, cfg)
	if err != nil {
		log.Fatal("Postgres init failed:", err)
	}
	defer database.Pool.Close()

	log.Println("Connected to Postgres on port", cfg.DBPort)
	fmt.Println(database)

	repo := repository.NewOrderRepository(database.Pool)
	c := cache.New()

	broker := "localhost:9092"
	topic := "orders"
	if err := kafka_consumer.EnsureTopic(broker, topic, 1, 1); err != nil {
		log.Fatal("failed to ensure topic:", err)
	}

	// Consumer
	consumer := kafka_consumer.NewConsumer(
		[]string{broker},
		topic,
		"test-group",
		repo,
		c,
	)

	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Fatal("consumer failed:", err)
		}
	}()

	// Producer
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "orders",
	})
	defer writer.Close()

	TestOrderSaveAndGet(ctx, repo, c, writer)

}

func TestOrderSaveAndGet(ctx context.Context, repo *repository.OrderRepository, c *cache.Cache, writer *kafka.Writer) {
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

	time.Sleep(10 * time.Second)

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

//Create topics for kafka

//docker exec -it <kafka-container-id> kafka-topics.sh \ --create \ --topic orders \ --partitions 1 \ --replication-factor 1 \ --bootstrap-server localhost:9092

/*
docker exec -it <kafka-container-id> kafka-topics.sh \
  --list \
  --bootstrap-server localhost:9092
*/
