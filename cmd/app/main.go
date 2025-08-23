package main

import (
	"L0/internal/config"
	"L0/internal/db"
	"context"
	"fmt"
	"log"
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
