package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmozzze/order_checker/internal/config"
	"github.com/tmozzze/order_checker/internal/db"
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

}
