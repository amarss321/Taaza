package main

import (
	"admin-service/database"
	"admin-service/handlers"
	"admin-service/middleware"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	if err := database.Connect(); err != nil {
		logrus.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

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

	v1 := r.Group("/api/v1")
	{
		admin := v1.Group("/admin")
		admin.Use(middleware.AdminAuthMiddleware())
		{
			users := admin.Group("/users")
			{
				users.GET("", handlers.GetUsers)
				users.GET("/stats", handlers.GetUserStats)
				users.GET("/:id", handlers.GetUser)
				users.PUT("/:id", handlers.UpdateUser)
				users.GET("/:id/activity", handlers.GetUserActivity)
				users.GET("/:id/addresses", handlers.GetUserAddresses)
				users.GET("/:id/sessions", handlers.GetUserSessions)
			}
			
			admin.GET("/activity/recent", handlers.GetRecentActivity)
			admin.POST("/sessions/:id/revoke", handlers.RevokeSession)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "admin-service ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	
	logrus.Infof("Admin service starting on port %s", port)
	r.Run(":" + port)
}