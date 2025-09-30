package main

import (
	"email-service/config"
	"email-service/handlers"
	"email-service/queue"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	config.LoadConfig()
	queue.InitRedis()

	// Start worker pool
	workerPool := queue.NewWorkerPool(3)
	workerPool.Start()

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logrus.Info("Shutting down email service...")
		workerPool.Stop()
		os.Exit(0)
	}()

	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "Email service OK"})
	})

	// Email API routes
	v1 := r.Group("/api/v1")
	{
		email := v1.Group("/email")
		{
			email.POST("/send", handlers.SendEmail)
			email.POST("/otp", handlers.SendOTP)
			email.POST("/welcome", handlers.SendWelcome)
			email.POST("/profile-reminder", handlers.SendProfileReminder)
			email.GET("/stats", handlers.GetQueueStats)
		}
	}

	port := os.Getenv("PORT")
	logrus.Infof("Email service starting on port %s", port)
	r.Run(":" + port)
}