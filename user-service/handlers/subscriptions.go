package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"user-service/database"
	"user-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Subscription struct {
	ID              int                    `json:"id"`
	UserID          int                    `json:"user_id"`
	SubscriptionType string                `json:"subscription_type"`
	MorningEnabled  bool                   `json:"morning_enabled"`
	MorningMilkType string                 `json:"morning_milk_type"`
	MorningQuantity float64                `json:"morning_quantity"`
	MorningFrequency string               `json:"morning_frequency"`
	MorningTimeSlot string                `json:"morning_time_slot"`
	MorningDays     map[string]interface{} `json:"morning_days"`
	EveningEnabled  bool                   `json:"evening_enabled"`
	EveningMilkType string                 `json:"evening_milk_type"`
	EveningQuantity float64                `json:"evening_quantity"`
	EveningFrequency string               `json:"evening_frequency"`
	EveningTimeSlot string                `json:"evening_time_slot"`
	EveningDays     map[string]interface{} `json:"evening_days"`
	AddressData     map[string]interface{} `json:"address_data"`
	Status          string                 `json:"status"`
}

func GetUserSubscriptions(c *gin.Context) {
	userID := c.GetInt("user_id")

	query := `SELECT id, user_id, subscription_type, morning_enabled, morning_milk_type, 
			  morning_quantity, morning_frequency, morning_time_slot, morning_days,
			  evening_enabled, evening_milk_type, evening_quantity, evening_frequency, 
			  evening_time_slot, evening_days, address_data, status 
			  FROM user_subscriptions WHERE user_id = $1 AND status = 'active'`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var subscriptions []Subscription
	for rows.Next() {
		var s Subscription
		var morningDaysJSON, eveningDaysJSON, addressDataJSON []byte

		err := rows.Scan(&s.ID, &s.UserID, &s.SubscriptionType, &s.MorningEnabled,
			&s.MorningMilkType, &s.MorningQuantity, &s.MorningFrequency, &s.MorningTimeSlot,
			&morningDaysJSON, &s.EveningEnabled, &s.EveningMilkType, &s.EveningQuantity,
			&s.EveningFrequency, &s.EveningTimeSlot, &eveningDaysJSON, &addressDataJSON, &s.Status)
		if err != nil {
			continue
		}

		if morningDaysJSON != nil {
			json.Unmarshal(morningDaysJSON, &s.MorningDays)
		}
		if eveningDaysJSON != nil {
			json.Unmarshal(eveningDaysJSON, &s.EveningDays)
		}
		if addressDataJSON != nil {
			json.Unmarshal(addressDataJSON, &s.AddressData)
		}

		subscriptions = append(subscriptions, s)
	}

	utils.LogActivity(userID, "subscriptions_view", "User viewed subscriptions", c.ClientIP())
	c.JSON(http.StatusOK, subscriptions)
}

func CreateSubscription(c *gin.Context) {
	userID := c.GetInt("user_id")

	var s Subscription
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	morningDaysJSON, _ := json.Marshal(s.MorningDays)
	eveningDaysJSON, _ := json.Marshal(s.EveningDays)
	addressDataJSON, _ := json.Marshal(s.AddressData)

	// Check if user already has a subscription
	var existingID int
	checkQuery := `SELECT id FROM user_subscriptions WHERE user_id = $1 LIMIT 1`
	err := database.DB.QueryRow(checkQuery, userID).Scan(&existingID)
	
	if err == nil {
		// Update existing subscription
		updateQuery := `UPDATE user_subscriptions SET subscription_type = $1, morning_enabled = $2,
					  morning_milk_type = $3, morning_quantity = $4, morning_frequency = $5,
					  morning_time_slot = $6, morning_days = $7, evening_enabled = $8,
					  evening_milk_type = $9, evening_quantity = $10, evening_frequency = $11,
					  evening_time_slot = $12, evening_days = $13, address_data = $14,
					  status = $15, updated_at = CURRENT_TIMESTAMP
					  WHERE user_id = $16 RETURNING id`
		
		err = database.DB.QueryRow(updateQuery, s.SubscriptionType, s.MorningEnabled,
			s.MorningMilkType, s.MorningQuantity, s.MorningFrequency, s.MorningTimeSlot,
			morningDaysJSON, s.EveningEnabled, s.EveningMilkType, s.EveningQuantity,
			s.EveningFrequency, s.EveningTimeSlot, eveningDaysJSON, addressDataJSON,
			"active", userID).Scan(&s.ID)
		
		if err != nil {
			logrus.Error("Database update error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
			return
		}
		
		utils.LogActivity(userID, "subscription_update", "User updated subscription", c.ClientIP())
		c.JSON(http.StatusOK, s)
		return
	}

	// Create new subscription if none exists
	insertQuery := `INSERT INTO user_subscriptions (user_id, subscription_type, morning_enabled, 
			  morning_milk_type, morning_quantity, morning_frequency, morning_time_slot, 
			  morning_days, evening_enabled, evening_milk_type, evening_quantity, 
			  evening_frequency, evening_time_slot, evening_days, address_data, status)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
			  RETURNING id`

	err = database.DB.QueryRow(insertQuery, userID, s.SubscriptionType, s.MorningEnabled,
		s.MorningMilkType, s.MorningQuantity, s.MorningFrequency, s.MorningTimeSlot,
		morningDaysJSON, s.EveningEnabled, s.EveningMilkType, s.EveningQuantity,
		s.EveningFrequency, s.EveningTimeSlot, eveningDaysJSON, addressDataJSON, "active").Scan(&s.ID)

	if err != nil {
		logrus.Error("Database insert error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription"})
		return
	}

	utils.LogActivity(userID, "subscription_create", "User created subscription", c.ClientIP())
	c.JSON(http.StatusCreated, s)
}

func UpdateSubscription(c *gin.Context) {
	userID := c.GetInt("user_id")
	subscriptionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	var s Subscription
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	morningDaysJSON, _ := json.Marshal(s.MorningDays)
	eveningDaysJSON, _ := json.Marshal(s.EveningDays)
	addressDataJSON, _ := json.Marshal(s.AddressData)

	query := `UPDATE user_subscriptions SET morning_enabled = $1, morning_milk_type = $2,
			  morning_quantity = $3, morning_frequency = $4, morning_time_slot = $5,
			  morning_days = $6, evening_enabled = $7, evening_milk_type = $8,
			  evening_quantity = $9, evening_frequency = $10, evening_time_slot = $11,
			  evening_days = $12, address_data = $13, updated_at = CURRENT_TIMESTAMP
			  WHERE id = $14 AND user_id = $15`

	_, err = database.DB.Exec(query, s.MorningEnabled, s.MorningMilkType, s.MorningQuantity,
		s.MorningFrequency, s.MorningTimeSlot, morningDaysJSON, s.EveningEnabled,
		s.EveningMilkType, s.EveningQuantity, s.EveningFrequency, s.EveningTimeSlot,
		eveningDaysJSON, addressDataJSON, subscriptionID, userID)

	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
		return
	}

	utils.LogActivity(userID, "subscription_update", "User updated subscription", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Subscription updated successfully"})
}

func UpdateSubscriptionAddress(c *gin.Context) {
	userID := c.GetInt("user_id")
	subscriptionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	var addressData map[string]interface{}
	if err := c.ShouldBindJSON(&addressData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	addressDataJSON, _ := json.Marshal(addressData)

	query := `UPDATE user_subscriptions SET address_data = $1, updated_at = CURRENT_TIMESTAMP
			  WHERE id = $2 AND user_id = $3`

	_, err = database.DB.Exec(query, addressDataJSON, subscriptionID, userID)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update address"})
		return
	}

	utils.LogActivity(userID, "subscription_address_update", "User updated subscription address", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Address updated successfully"})
}

func DeleteSubscription(c *gin.Context) {
	userID := c.GetInt("user_id")
	subscriptionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	query := `UPDATE user_subscriptions SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP 
			  WHERE id = $1 AND user_id = $2`

	_, err = database.DB.Exec(query, subscriptionID, userID)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel subscription"})
		return
	}

	utils.LogActivity(userID, "subscription_cancel", "User cancelled subscription", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Subscription cancelled successfully"})
}