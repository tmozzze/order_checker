package main

import (
	"L0/internal/config"
	"fmt"
)

func main() {
	cfg := config.Load()

	fmt.Println(cfg)
}
