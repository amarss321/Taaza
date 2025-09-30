package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"user-service/database"
	"user-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Address struct {
	ID          int     `json:"id"`
	UserID      int     `json:"user_id"`
	Label       string  `json:"label"`
	AddressLine string  `json:"address_line"`
	City        string  `json:"city"`
	State       string  `json:"state"`
	ZipCode     string  `json:"zip_code"`
	Country     string  `json:"country"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	IsDefault   bool    `json:"is_default"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type AddressRequest struct {
	Label       string   `json:"label" validate:"required,min=1,max=50"`
	AddressLine string   `json:"address_line" validate:"required,min=1"`
	City        string   `json:"city" validate:"required,min=1,max=100"`
	State       string   `json:"state" validate:"max=100"`
	ZipCode     string   `json:"zip_code" validate:"required,min=1,max=20"`
	Country     string   `json:"country" validate:"required,min=1,max=100"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	IsDefault   bool     `json:"is_default"`
}

// GetAddresses retrieves all addresses for a user
func GetAddresses(c *gin.Context) {
	userID := c.GetInt("user_id")

	rows, err := database.DB.Query(`
		SELECT id, user_id, label, address_line, city, COALESCE(state, '') as state, 
		       zip_code, country, latitude, longitude, is_default, created_at, updated_at
		FROM user_addresses 
		WHERE user_id = $1 
		ORDER BY is_default DESC, created_at DESC`, userID)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var addresses []Address
	for rows.Next() {
		var addr Address
		err := rows.Scan(&addr.ID, &addr.UserID, &addr.Label, &addr.AddressLine, 
			&addr.City, &addr.State, &addr.ZipCode, &addr.Country, 
			&addr.Latitude, &addr.Longitude, &addr.IsDefault, &addr.CreatedAt, &addr.UpdatedAt)
		if err != nil {
			logrus.Error("Row scan error:", err)
			continue
		}
		addresses = append(addresses, addr)
	}

	utils.LogActivity(userID, "addresses_view", "User viewed addresses", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"addresses": addresses})
}

// CreateAddress creates a new address for a user
func CreateAddress(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req AddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := utils.Validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If this is set as default, unset other defaults
	if req.IsDefault {
		_, err := database.DB.Exec("UPDATE user_addresses SET is_default = FALSE WHERE user_id = $1", userID)
		if err != nil {
			logrus.Error("Database error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
	}

	var addressID int
	err := database.DB.QueryRow(`
		INSERT INTO user_addresses (user_id, label, address_line, city, state, zip_code, country, latitude, longitude, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`, 
		userID, req.Label, req.AddressLine, req.City, req.State, req.ZipCode, req.Country, req.Latitude, req.Longitude, req.IsDefault).Scan(&addressID)
	
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	utils.LogActivity(userID, "address_create", "User created new address: "+req.Label, c.ClientIP())
	c.JSON(http.StatusCreated, gin.H{"message": "Address created successfully", "address_id": addressID})
}

// UpdateAddress updates an existing address
func UpdateAddress(c *gin.Context) {
	userID := c.GetInt("user_id")
	addressID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
		return
	}

	var req AddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := utils.Validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if address belongs to user
	var existingUserID int
	err = database.DB.QueryRow("SELECT user_id FROM user_addresses WHERE id = $1", addressID).Scan(&existingUserID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
		return
	}
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	if existingUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// If this is set as default, unset other defaults
	if req.IsDefault {
		_, err := database.DB.Exec("UPDATE user_addresses SET is_default = FALSE WHERE user_id = $1 AND id != $2", userID, addressID)
		if err != nil {
			logrus.Error("Database error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
	}

	_, err = database.DB.Exec(`
		UPDATE user_addresses 
		SET label = $1, address_line = $2, city = $3, state = $4, zip_code = $5, 
		    country = $6, latitude = $7, longitude = $8, is_default = $9, updated_at = NOW()
		WHERE id = $10 AND user_id = $11`,
		req.Label, req.AddressLine, req.City, req.State, req.ZipCode, req.Country, 
		req.Latitude, req.Longitude, req.IsDefault, addressID, userID)
	
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	utils.LogActivity(userID, "address_update", "User updated address: "+req.Label, c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Address updated successfully"})
}

// DeleteAddress deletes an address
func DeleteAddress(c *gin.Context) {
	userID := c.GetInt("user_id")
	addressID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
		return
	}

	// Check if address belongs to user and get label for logging
	var existingUserID int
	var label string
	err = database.DB.QueryRow("SELECT user_id, label FROM user_addresses WHERE id = $1", addressID).Scan(&existingUserID, &label)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
		return
	}
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	if existingUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	_, err = database.DB.Exec("DELETE FROM user_addresses WHERE id = $1 AND user_id = $2", addressID, userID)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	utils.LogActivity(userID, "address_delete", "User deleted address: "+label, c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Address deleted successfully"})
}

// SetDefaultAddress sets an address as default
func SetDefaultAddress(c *gin.Context) {
	userID := c.GetInt("user_id")
	addressID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
		return
	}

	// Check if address belongs to user
	var existingUserID int
	var label string
	err = database.DB.QueryRow("SELECT user_id, label FROM user_addresses WHERE id = $1", addressID).Scan(&existingUserID, &label)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
		return
	}
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	if existingUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Unset all defaults for user
	_, err = database.DB.Exec("UPDATE user_addresses SET is_default = FALSE WHERE user_id = $1", userID)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Set this address as default
	_, err = database.DB.Exec("UPDATE user_addresses SET is_default = TRUE WHERE id = $1 AND user_id = $2", addressID, userID)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	utils.LogActivity(userID, "address_default", "User set default address: "+label, c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Default address updated successfully"})
}