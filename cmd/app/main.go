package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tmozzze/order_checker/internal/config"
	"github.com/tmozzze/order_checker/internal/db"
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

	TestOrderRepoSaveAndGet(ctx, database)

}

func TestOrderRepoSaveAndGet(ctx context.Context, database *db.DB) {
	order := &models.Order{
		OrderUID:    "b563feb7b2b84b6test",
		TrackNumber: "WBILMTESTTRACK",
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
			Transaction:  "b563feb7b2b84b6test",
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
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				RID:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
	}

	OrderRepo := repository.NewOrderRepository(database.Pool)
	err := OrderRepo.SaveOrder(ctx, order)
	if err != nil {
		log.Fatal("Save in db error:", err)
	}
	fmt.Println("Order was saved succesfully!")

	gotOrder, err := OrderRepo.GetOrderById(ctx, order.OrderUID)
	if err != nil {
		log.Fatal("Get from db Error", err)
	}
	fmt.Println("Order got succesfully!")
	fmt.Println(gotOrder)

}
