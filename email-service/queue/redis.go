package queue

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type EmailJob struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	To           string                 `json:"to"`
	Subject      string                 `json:"subject"`
	TemplateName string                 `json:"template_name"`
	Data         map[string]interface{} `json:"data"`
	Attempts     int                    `json:"attempts"`
	MaxAttempts  int                    `json:"max_attempts"`
	CreatedAt    time.Time              `json:"created_at"`
	ScheduledAt  *time.Time             `json:"scheduled_at,omitempty"`
}

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

func EnqueueEmail(job EmailJob) error {
	ctx := context.Background()
	
	jobData, err := json.Marshal(job)
	if err != nil {
		return err
	}

	if job.ScheduledAt != nil && job.ScheduledAt.After(time.Now()) {
		// Schedule for later
		score := float64(job.ScheduledAt.Unix())
		return RedisClient.ZAdd(ctx, "email_scheduled", &redis.Z{
			Score:  score,
			Member: jobData,
		}).Err()
	}

	// Add to immediate queue
	return RedisClient.LPush(ctx, "email_queue", jobData).Err()
}

func DequeueEmail() (*EmailJob, error) {
	ctx := context.Background()
	
	// Check scheduled emails first
	now := float64(time.Now().Unix())
	scheduled, err := RedisClient.ZRangeByScore(ctx, "email_scheduled", &redis.ZRangeBy{
		Min:   "0",
		Max:   string(rune(int(now))),
		Count: 1,
	}).Result()

	if err == nil && len(scheduled) > 0 {
		// Move scheduled email to immediate queue
		RedisClient.ZRem(ctx, "email_scheduled", scheduled[0])
		RedisClient.LPush(ctx, "email_queue", scheduled[0])
	}

	// Get from immediate queue
	result, err := RedisClient.BRPop(ctx, 5*time.Second, "email_queue").Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No jobs available
		}
		return nil, err
	}

	var job EmailJob
	err = json.Unmarshal([]byte(result[1]), &job)
	return &job, err
}

func RequeueEmail(job EmailJob) error {
	job.Attempts++
	
	if job.Attempts >= job.MaxAttempts {
		// Move to dead letter queue
		ctx := context.Background()
		jobData, _ := json.Marshal(job)
		return RedisClient.LPush(ctx, "email_dead_letter", jobData).Err()
	}

	// Exponential backoff
	delay := time.Duration(job.Attempts*job.Attempts) * time.Minute
	scheduledAt := time.Now().Add(delay)
	job.ScheduledAt = &scheduledAt

	return EnqueueEmail(job)
}

func CheckRateLimit(email string, limit int, window time.Duration) bool {
	ctx := context.Background()
	key := "email_rate_limit:" + email
	
	current, err := RedisClient.Get(ctx, key).Int()
	if err == redis.Nil {
		RedisClient.Set(ctx, key, 1, window)
		return true
	} else if err != nil {
		logrus.Error("Redis error:", err)
		return true
	}

	if current >= limit {
		return false
	}

	RedisClient.Incr(ctx, key)
	return true
}