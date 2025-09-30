package utils

import (
	"context"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

var RedisClient *redis.Client

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})

	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		logrus.Fatal("Failed to connect to Redis:", err)
	}
	logrus.Info("Redis connected successfully")
}

func CheckRateLimit(key string, limit int, window time.Duration) bool {
	ctx := context.Background()
	
	current, err := RedisClient.Get(ctx, key).Int()
	if err == redis.Nil {
		// First request
		RedisClient.Set(ctx, key, 1, window)
		return true
	} else if err != nil {
		logrus.Error("Redis error:", err)
		return true // Allow on error
	}

	if current >= limit {
		return false
	}

	RedisClient.Incr(ctx, key)
	return true
}