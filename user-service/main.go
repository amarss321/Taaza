package main

import (
	"os"
	"user-service/config"
	"user-service/database"
	"user-service/handlers"
	"user-service/middleware"
	"user-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	config.LoadConfig()

	if err := database.Connect(); err != nil {
		logrus.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	utils.InitRedis()

	// Start session cleanup goroutine
	go utils.StartSessionCleanup()

	r := gin.Default()
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.Compression())
	r.Use(middleware.CacheControl())
	r.Use(middleware.RateLimit())
	r.Use(middleware.ValidateInput())
	r.Use(middleware.SanitizeInput())

	v1 := r.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("/start-registration", handlers.StartRegistration)
			users.POST("/verify-otp", handlers.VerifyOTP)
			users.POST("/login-otp", handlers.LoginWithOTP)
			users.POST("/login-verify-otp", handlers.LoginVerifyOTP)
			users.POST("/complete-profile", handlers.CompleteUserProfile)
			users.POST("/login", handlers.Login)
			users.POST("/logout", handlers.Logout)
			users.POST("/logout-all", middleware.AuthMiddleware(), handlers.LogoutAll)
			users.GET("/profile", middleware.AuthMiddleware(), handlers.GetProfileSecure)
			users.PUT("/update-profile", middleware.AuthMiddleware(), handlers.UpdateUserProfileSecure)
			users.PUT("/personal-info", middleware.AuthMiddleware(), handlers.UpdatePersonalInfo)
			users.POST("/upload-picture", middleware.AuthMiddleware(), handlers.UploadProfilePicture)
			
			// Address routes
			users.GET("/addresses", middleware.AuthMiddleware(), handlers.GetAddresses)
			users.POST("/addresses", middleware.AuthMiddleware(), handlers.CreateAddress)
			users.PUT("/addresses/:id", middleware.AuthMiddleware(), handlers.UpdateAddress)
			users.DELETE("/addresses/:id", middleware.AuthMiddleware(), handlers.DeleteAddress)
			users.PUT("/addresses/:id/default", middleware.AuthMiddleware(), handlers.SetDefaultAddress)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "user-service ok"})
	})

	// Serve uploaded files
	r.Static("/uploads", "./uploads")

	port := os.Getenv("PORT")
	logrus.Infof("Server starting on port %s", port)
	r.Run(":" + port)
}