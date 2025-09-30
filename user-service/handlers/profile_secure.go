package handlers

import (
	"database/sql"
	"net/http"
	"regexp"
	"strings"
	"time"
	"user-service/database"
	"user-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Input sanitization and validation
var (
	nameRegex = regexp.MustCompile(`^[a-zA-Z\s]{2,50}$`)
	dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

func sanitizeString(input string) string {
	return strings.TrimSpace(strings.ReplaceAll(input, "\n", " "))
}

func validateName(name string) bool {
	return nameRegex.MatchString(name)
}

func validateDate(dateStr string) bool {
	if !dateRegex.MatchString(dateStr) {
		return false
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return false
	}
	now := time.Now()
	minAge := now.AddDate(-120, 0, 0)
	maxAge := now.AddDate(-13, 0, 0)
	return date.After(minAge) && date.Before(maxAge)
}

func validateGender(gender string) bool {
	validGenders := map[string]bool{"male": true, "female": true, "other": true, "": true}
	return validGenders[strings.ToLower(gender)]
}

func GetProfileSecure(c *gin.Context) {
	userID := c.GetInt("user_id")
	
	// Additional security check
	if userID <= 0 {
		logrus.Warn("Invalid user ID in token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	
	logrus.WithField("user_id", userID).Info("GetProfile called")

	// Use sql.NullString for nullable fields
	var (
		id                 int
		name               string
		email              string
		mobile             string
		gender             sql.NullString
		dateOfBirth        sql.NullString
		registrationStatus string
	)

	// Query without COALESCE for nullable fields
	query := `SELECT id, name, email, COALESCE(mobile, '') as mobile, 
			 gender, date_of_birth, registration_status FROM users WHERE id = $1`
	
	err := database.DB.QueryRow(query, userID).Scan(
		&id, &name, &email, &mobile, 
		&gender, &dateOfBirth, &registrationStatus)

	if err == sql.ErrNoRows {
		logrus.WithField("user_id", userID).Warn("User not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		logrus.WithFields(logrus.Fields{"user_id": userID}).WithError(err).Error("Database error in GetProfile")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Build response with proper null handling
	var user struct {
		ID                 int    `json:"id"`
		Name               string `json:"name"`
		Email              string `json:"email"`
		Mobile             string `json:"mobile"`
		Gender             string `json:"gender"`
		DateOfBirth        string `json:"date_of_birth"`
		RegistrationStatus string `json:"registration_status"`
	}

	user.ID = id
	user.Name = name
	user.Email = email
	user.Mobile = mobile
	user.RegistrationStatus = registrationStatus
	
	// Handle nullable fields
	if gender.Valid {
		user.Gender = gender.String
	} else {
		user.Gender = ""
	}
	
	if dateOfBirth.Valid {
		user.DateOfBirth = dateOfBirth.String
	} else {
		user.DateOfBirth = ""
	}

	// Set no-cache headers to ensure fresh data
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	
	logrus.WithFields(logrus.Fields{"user_id": user.ID, "name_length": len(user.Name), "gender": user.Gender, "dob": user.DateOfBirth}).Info("Successfully retrieved profile")
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func UpdateUserProfileSecure(c *gin.Context) {
	userID := c.GetInt("user_id")
	
	// Additional security check
	if userID <= 0 {
		logrus.Warn("Invalid user ID in token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	type UpdateUserProfileRequest struct {
		Name        string `json:"name" validate:"required,min=2,max=50"`
		Gender      string `json:"gender"`
		DateOfBirth string `json:"date_of_birth"`
	}

	var req UpdateUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Sanitize inputs
	req.Name = sanitizeString(req.Name)
	req.Gender = sanitizeString(req.Gender)
	req.DateOfBirth = sanitizeString(req.DateOfBirth)

	// Validate inputs
	if !validateName(req.Name) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name must contain only letters and spaces, 2-50 characters"})
		return
	}

	if req.Gender != "" && !validateGender(req.Gender) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gender value"})
		return
	}

	if req.DateOfBirth != "" && !validateDate(req.DateOfBirth) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date of birth. Must be YYYY-MM-DD format and reasonable age"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"name_length": len(req.Name),
		"has_gender": req.Gender != "",
		"has_dob": req.DateOfBirth != "",
	}).Info("Updating user profile")

	// Handle null values properly for optional fields
	var dateOfBirth interface{}
	if req.DateOfBirth == "" || req.DateOfBirth == "null" {
		dateOfBirth = nil
	} else {
		dateOfBirth = req.DateOfBirth
	}
	
	var gender interface{}
	if req.Gender == "" || req.Gender == "null" {
		gender = nil
	} else {
		gender = req.Gender
	}
	
	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"final_gender": gender,
		"final_dob": dateOfBirth,
	}).Info("Final values before database update")
	
	// Optimized update query
	query := `UPDATE users SET name = $1, gender = $2, date_of_birth = $3, updated_at = NOW() 
			 WHERE id = $4`
	
	_, err := database.DB.Exec(query, req.Name, gender, dateOfBirth, userID)
	if err != nil {
		logrus.WithError(err).Error("Database error during user profile update")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Log activity
	utils.LogActivity(userID, "personal_info_update", "User updated personal information", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}