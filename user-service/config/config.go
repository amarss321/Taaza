package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found")
	}

	// Set default values
	if os.Getenv("PORT") == "" {
		os.Setenv("PORT", "8081")
	}
	if os.Getenv("DB_HOST") == "" {
		os.Setenv("DB_HOST", "user-db")
	}
	if os.Getenv("DB_PORT") == "" {
		os.Setenv("DB_PORT", "5432")
	}
	if os.Getenv("REDIS_URL") == "" {
		os.Setenv("REDIS_URL", "redis:6379")
	}
}