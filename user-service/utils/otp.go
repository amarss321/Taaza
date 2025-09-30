package utils

import (
	"context"
	"fmt"
	"math/rand"
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

func GenerateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func StoreOTP(email, otp string) error {
	ctx := context.Background()
	key := fmt.Sprintf("otp:%s", email)
	return RedisClient.Set(ctx, key, otp, 10*time.Minute).Err()
}

func VerifyOTP(email, otp string) bool {
	ctx := context.Background()
	key := fmt.Sprintf("otp:%s", email)
	
	storedOTP, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return false
	}

	if storedOTP == otp {
		RedisClient.Del(ctx, key)
		return true
	}
	return false
}