package main

import (
	"api-gateway/config"
	"api-gateway/handlers"
	"api-gateway/middleware"
	"api-gateway/utils"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	config.LoadConfig()
	utils.InitRedis()

	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "API Gateway OK"})
	})

	v1 := r.Group("/api/v1")
	{
		// User routes
		users := v1.Group("/users")
		{
			// Public routes (no auth required)
			users.POST("/start-registration", middleware.OTPRateLimitMiddleware(), handlers.ProxyToUserService)
			users.POST("/verify-otp", handlers.ProxyToUserService)
			users.POST("/login-otp", middleware.OTPRateLimitMiddleware(), handlers.ProxyToUserService)
			users.POST("/login-verify-otp", handlers.ProxyToUserService)
			users.POST("/complete-profile", handlers.ProxyToUserService)
			users.POST("/login", handlers.ProxyToUserService)

			// Protected routes (auth required)
			protected := users.Group("")
			protected.Use(middleware.AuthMiddleware())
			{
				protected.GET("/profile", handlers.ProxyToUserService)
				protected.PUT("/profile", handlers.ProxyToUserService)
				protected.PUT("/update-profile", handlers.ProxyToUserService)
				protected.PUT("/personal-info", handlers.ProxyToUserService)
				protected.POST("/upload-picture", handlers.ProxyToUserService)
			}
		}

		// Admin routes
		admin := v1.Group("/admin")
		{
			admin.Any("/users/*path", handlers.ProxyToAdminService)
			admin.Any("/users", handlers.ProxyToAdminService)
		}
	}

	// Catch-all for other routes
	r.NoRoute(handlers.RouteBasedOnJWT)

	port := os.Getenv("PORT")
	logrus.Infof("API Gateway starting on port %s", port)
	r.Run(":" + port)
}