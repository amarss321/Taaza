package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize database
	if err := InitDB(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// API routes
	api := r.Group("/api/v1/inventory")
	{
		// Products
		api.GET("/products", GetProducts)
		api.PUT("/products/:id/price", UpdateProductPrice)

		// Stock management
		api.GET("/stock", GetStock)
		api.PUT("/stock/:productId/:timeSlot", UpdateStock)
		api.POST("/stock/:productId/:timeSlot/adjust", AdjustStock)

		// Bookings
		api.POST("/bookings", AddBooking)
		api.DELETE("/bookings", RemoveBooking)

		// Notifications
		api.GET("/notifications", GetNotifications)
		api.PUT("/notifications/:id/status", UpdateNotificationStatus)
		api.DELETE("/notifications/:id", DeleteNotification)
		api.POST("/notifications", CreateNotification)

		// Analytics
		api.GET("/analytics/summary", GetAnalyticsSummary)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	log.Printf("Inventory service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}