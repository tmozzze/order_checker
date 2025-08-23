package config

import (
	"errors"
	"fmt"
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

func Load() (*Config, error) {

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
		err := errors.New("DB_USER or DB_PASSWORD is empty")
		return nil, err
	}

	return cfg, nil
}
