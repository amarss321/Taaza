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
		os.Setenv("PORT", "8084")
	}
	if os.Getenv("REDIS_URL") == "" {
		os.Setenv("REDIS_URL", "redis:6379")
	}
	if os.Getenv("SMTP_HOST") == "" {
		os.Setenv("SMTP_HOST", "smtp.gmail.com")
	}
	if os.Getenv("SMTP_PORT") == "" {
		os.Setenv("SMTP_PORT", "587")
	}
	if os.Getenv("APP_URL") == "" {
		os.Setenv("APP_URL", "http://localhost:3000")
	}
	if os.Getenv("SMTP_FROM") == "" {
		os.Setenv("SMTP_FROM", os.Getenv("SMTP_USER"))
	}
}