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
		os.Setenv("PORT", "8080")
	}
	if os.Getenv("USER_SERVICE_URL") == "" {
		os.Setenv("USER_SERVICE_URL", "http://user-service:8081")
	}
	if os.Getenv("ADMIN_SERVICE_URL") == "" {
		os.Setenv("ADMIN_SERVICE_URL", "http://admin-service:8082")
	}
	if os.Getenv("REDIS_URL") == "" {
		os.Setenv("REDIS_URL", "redis:6379")
	}
}