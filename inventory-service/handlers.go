package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetProducts returns all products
func GetProducts(c *gin.Context) {
	rows, err := db.Query("SELECT id, name, type, price_per_liter, created_at, updated_at FROM products ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.Name, &p.Type, &p.PricePerLiter, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		products = append(products, p)
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

// GetStock returns current stock for all products
func GetStock(c *gin.Context) {
	query := `
		SELECT s.id, s.product_id, s.time_slot, s.total_stock, s.booked_stock, 
		       s.available_stock, s.updated_at, p.name, p.type, p.price_per_liter
		FROM inventory_stock s
		JOIN products p ON s.product_id = p.id
		ORDER BY p.id, s.time_slot
	`
	
	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	stockMap := make(map[string]ProductSummary)
	
	for rows.Next() {
		var s InventoryStock
		var productName, productType string
		var price float64
		
		err := rows.Scan(&s.ID, &s.ProductID, &s.TimeSlot, &s.TotalStock, &s.BookedStock, 
			&s.AvailableStock, &s.UpdatedAt, &productName, &productType, &price)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		key := strconv.Itoa(s.ProductID)
		if _, exists := stockMap[key]; !exists {
			stockMap[key] = ProductSummary{
				ProductID:   s.ProductID,
				ProductName: productName,
				Type:        productType,
				Price:       price,
			}
		}

		summary := stockMap[key]
		stockSummary := StockSummary{
			TotalStock:     s.TotalStock,
			BookedStock:    s.BookedStock,
			AvailableStock: s.AvailableStock,
			Revenue:        s.BookedStock * price,
		}

		if s.TimeSlot == "morning" {
			summary.Morning = stockSummary
		} else {
			summary.Evening = stockSummary
		}
		stockMap[key] = summary
	}

	var products []ProductSummary
	for _, summary := range stockMap {
		products = append(products, summary)
	}

	c.JSON(http.StatusOK, gin.H{"stock": products})
}

// UpdateStock updates total stock for a product and time slot
func UpdateStock(c *gin.Context) {
	productID, _ := strconv.Atoi(c.Param("productId"))
	timeSlot := c.Param("timeSlot")
	
	var req struct {
		TotalStock float64 `json:"total_stock"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current stock and booked stock
	var currentStock, bookedStock float64
	err := db.QueryRow("SELECT total_stock, booked_stock FROM inventory_stock WHERE product_id = $1 AND time_slot = $2", 
		productID, timeSlot).Scan(&currentStock, &bookedStock)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stock not found"})
		return
	}

	// Validate stock changes
	if req.TotalStock < bookedStock {
		availableToDecrease := currentStock - bookedStock
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot decrease stock below booked amount. Current booked: " + strconv.FormatFloat(bookedStock, 'f', 1, 64) + "L. You can only decrease by " + strconv.FormatFloat(availableToDecrease, 'f', 1, 64) + "L"})
		return
	}
	
	// If trying to decrease stock, validate the decrease amount
	if req.TotalStock < currentStock {
		availableToDecrease := currentStock - bookedStock
		maxAllowedStock := bookedStock + availableToDecrease
		if req.TotalStock > maxAllowedStock {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stock decrease. Current: " + strconv.FormatFloat(currentStock, 'f', 1, 64) + "L, Booked: " + strconv.FormatFloat(bookedStock, 'f', 1, 64) + "L. You can only decrease by available stock (" + strconv.FormatFloat(availableToDecrease, 'f', 1, 64) + "L). Minimum allowed: " + strconv.FormatFloat(bookedStock, 'f', 1, 64) + "L"})
			return
		}
	}

	// Prevent no-change updates when stock is being "decreased" to the same value
	if req.TotalStock == currentStock {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No change in stock. Current stock is already " + strconv.FormatFloat(currentStock, 'f', 1, 64) + "L"})
		return
	}
	
	// Update stock
	_, err = db.Exec("UPDATE inventory_stock SET total_stock = $1 WHERE product_id = $2 AND time_slot = $3",
		req.TotalStock, productID, timeSlot)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log history
	_, err = db.Exec(`INSERT INTO stock_history (product_id, time_slot, change_type, quantity, previous_value, new_value, reason, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		productID, timeSlot, "stock_update", req.TotalStock-currentStock, currentStock, req.TotalStock, "Manual update", "admin")
	if err != nil {
		// Log error but don't fail the request
		println("Failed to log stock history:", err.Error())
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stock updated successfully"})
}

// AdjustStock adjusts stock by a specific amount
func AdjustStock(c *gin.Context) {
	productID, _ := strconv.Atoi(c.Param("productId"))
	timeSlot := c.Param("timeSlot")
	
	var req StockAdjustment
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current stock and booked stock
	var currentStock, bookedStock float64
	err := db.QueryRow("SELECT total_stock, booked_stock FROM inventory_stock WHERE product_id = $1 AND time_slot = $2", 
		productID, timeSlot).Scan(&currentStock, &bookedStock)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stock not found"})
		return
	}

	newStock := currentStock + req.Quantity
	if newStock < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stock cannot be negative"})
		return
	}
	
	// Validate that new stock is not less than booked stock
	if newStock < bookedStock {
		availableToDecrease := currentStock - bookedStock
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot decrease stock below booked amount. Current booked: " + strconv.FormatFloat(bookedStock, 'f', 1, 64) + "L. You can only decrease by " + strconv.FormatFloat(availableToDecrease, 'f', 1, 64) + "L"})
		return
	}

	// Update stock
	_, err = db.Exec("UPDATE inventory_stock SET total_stock = $1 WHERE product_id = $2 AND time_slot = $3",
		newStock, productID, timeSlot)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log history
	changeType := "stock_add"
	if req.Quantity < 0 {
		changeType = "stock_remove"
	}
	
	_, err = db.Exec(`INSERT INTO stock_history (product_id, time_slot, change_type, quantity, previous_value, new_value, reason, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		productID, timeSlot, changeType, req.Quantity, currentStock, newStock, req.Reason, "admin")

	c.JSON(http.StatusOK, gin.H{"message": "Stock adjusted successfully", "new_stock": newStock})
}

// UpdateProductPrice updates product price
func UpdateProductPrice(c *gin.Context) {
	productID, _ := strconv.Atoi(c.Param("id"))
	
	var req PriceUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price must be greater than 0"})
		return
	}

	_, err := db.Exec("UPDATE products SET price_per_liter = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
		req.Price, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Price updated successfully"})
}

// AddBooking adds booking to stock
func AddBooking(c *gin.Context) {
	var req BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current stock and booked stock
	var totalStock, currentBooked float64
	err := db.QueryRow("SELECT total_stock, booked_stock FROM inventory_stock WHERE product_id = $1 AND time_slot = $2", 
		req.ProductID, req.TimeSlot).Scan(&totalStock, &currentBooked)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stock not found"})
		return
	}

	newBooked := currentBooked + req.Quantity
	
	// Validate that booking doesn't exceed available stock
	if newBooked > totalStock {
		availableStock := totalStock - currentBooked
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot book more than available stock. Available: " + strconv.FormatFloat(availableStock, 'f', 1, 64) + "L"})
		return
	}

	// Update booked stock
	_, err = db.Exec("UPDATE inventory_stock SET booked_stock = $1 WHERE product_id = $2 AND time_slot = $3",
		newBooked, req.ProductID, req.TimeSlot)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log history
	_, err = db.Exec(`INSERT INTO stock_history (product_id, time_slot, change_type, quantity, previous_value, new_value, reason, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		req.ProductID, req.TimeSlot, "booking_add", req.Quantity, currentBooked, newBooked, req.Reason, "system")

	c.JSON(http.StatusOK, gin.H{"message": "Booking added successfully"})
}

// RemoveBooking removes booking from stock
func RemoveBooking(c *gin.Context) {
	var req BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current booked stock
	var currentBooked float64
	err := db.QueryRow("SELECT booked_stock FROM inventory_stock WHERE product_id = $1 AND time_slot = $2", 
		req.ProductID, req.TimeSlot).Scan(&currentBooked)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stock not found"})
		return
	}

	newBooked := currentBooked - req.Quantity
	if newBooked < 0 {
		newBooked = 0
	}

	// Update booked stock
	_, err = db.Exec("UPDATE inventory_stock SET booked_stock = $1 WHERE product_id = $2 AND time_slot = $3",
		newBooked, req.ProductID, req.TimeSlot)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log history
	_, err = db.Exec(`INSERT INTO stock_history (product_id, time_slot, change_type, quantity, previous_value, new_value, reason, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		req.ProductID, req.TimeSlot, "booking_remove", req.Quantity, currentBooked, newBooked, req.Reason, "system")

	c.JSON(http.StatusOK, gin.H{"message": "Booking removed successfully"})
}

// GetNotifications returns all notification requests
func GetNotifications(c *gin.Context) {
	query := `
		SELECT id, customer_name, phone_number, milk_type, quantity, time_slot, status, created_at, notified_at, notes
		FROM notification_requests
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}
	defer rows.Close()

	var notifications []NotificationRequest
	for rows.Next() {
		var notification NotificationRequest
		err := rows.Scan(
			&notification.ID,
			&notification.CustomerName,
			&notification.PhoneNumber,
			&notification.MilkType,
			&notification.Quantity,
			&notification.TimeSlot,
			&notification.Status,
			&notification.CreatedAt,
			&notification.NotifiedAt,
			&notification.Notes,
		)
		if err != nil {
			continue
		}
		notifications = append(notifications, notification)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Notifications retrieved successfully",
		"notifications": notifications,
	})
}

// CreateNotification creates a new notification request
func CreateNotification(c *gin.Context) {
	var req CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
		INSERT INTO notification_requests (customer_name, phone_number, milk_type, quantity, time_slot, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	var id int
	var createdAt time.Time
	err := db.QueryRow(query, req.CustomerName, req.PhoneNumber, req.MilkType, req.Quantity, req.TimeSlot, req.Notes).Scan(&id, &createdAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification request"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Notification request created successfully",
		"id":      id,
		"created_at": createdAt,
	})
}

// UpdateNotificationStatus updates notification status
func UpdateNotificationStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	
	var req struct {
		Status string `json:"status"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var notifiedAt *time.Time
	if req.Status == "notified" {
		now := time.Now()
		notifiedAt = &now
	}

	_, err := db.Exec("UPDATE notification_requests SET status = $1, notified_at = $2 WHERE id = $3",
		req.Status, notifiedAt, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification status updated successfully"})
}

// DeleteNotification deletes a notification request
func DeleteNotification(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	
	_, err := db.Exec("DELETE FROM notification_requests WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}

// GetAnalyticsSummary returns analytics summary
func GetAnalyticsSummary(c *gin.Context) {
	query := `
		SELECT s.product_id, p.name, p.type, p.price_per_liter, s.time_slot,
		       s.total_stock, s.booked_stock, s.available_stock
		FROM inventory_stock s
		JOIN products p ON s.product_id = p.id
		ORDER BY p.id, s.time_slot
	`
	
	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	productMap := make(map[int]*ProductSummary)
	var totalStock, totalBooked, totalAvailable, dailyRevenue float64

	for rows.Next() {
		var productID int
		var productName, productType, timeSlot string
		var price, stock, booked, available float64
		
		err := rows.Scan(&productID, &productName, &productType, &price, &timeSlot,
			&stock, &booked, &available)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if _, exists := productMap[productID]; !exists {
			productMap[productID] = &ProductSummary{
				ProductID:   productID,
				ProductName: productName,
				Type:        productType,
				Price:       price,
			}
		}

		summary := productMap[productID]
		stockSummary := StockSummary{
			TotalStock:     stock,
			BookedStock:    booked,
			AvailableStock: available,
			Revenue:        booked * price,
		}

		if timeSlot == "morning" {
			summary.Morning = stockSummary
		} else {
			summary.Evening = stockSummary
		}

		totalStock += stock
		totalBooked += booked
		totalAvailable += available
		dailyRevenue += booked * price
	}

	var products []ProductSummary
	for _, summary := range productMap {
		products = append(products, *summary)
	}

	analytics := AnalyticsSummary{
		TotalStock:     totalStock,
		TotalBooked:    totalBooked,
		TotalAvailable: totalAvailable,
		DailyRevenue:   dailyRevenue,
		Products:       products,
	}

	c.JSON(http.StatusOK, analytics)
}