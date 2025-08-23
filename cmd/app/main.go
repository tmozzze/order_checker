package main

import (
	"L0/internal/config"
	"L0/internal/db"
	"context"
	"fmt"
	"log"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	_, err := db.NewDB(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Connected to Postgres on port %s\n", cfg.DBPort)

}
