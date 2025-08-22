package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
}

func Load() *Config {

	if err := godotenv.Load(); err != nil {
		fmt.Println(".env not found. Skip")
	}

	cfg := &Config{
		DBUser:     os.Getenv("POSTGRES_USER"),
		DBPassword: os.Getenv("POSTGRES_PASSWORD"),
		DBHost:     os.Getenv("POSTGRES_HOST"),
		DBPort:     os.Getenv("POSTGRES_PORT"),
		DBName:     os.Getenv("POSTGRES_DB"),
	}

	if cfg.DBUser == "" || cfg.DBPassword == "" {
		log.Fatal("Config error: DB_USER or DB_PASSWORD is empty")
	}

	return cfg
}
