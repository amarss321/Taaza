package handlers

import (
	"email-service/queue"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type SendEmailRequest struct {
	Type         string                 `json:"type" binding:"required"`
	To           string                 `json:"to" binding:"required,email"`
	Subject      string                 `json:"subject"`
	TemplateName string                 `json:"template_name"`
	Data         map[string]interface{} `json:"data"`
	ScheduleAt   *time.Time             `json:"schedule_at,omitempty"`
}

func SendEmail(c *gin.Context) {
	var req SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate job ID
	jobID := generateJobID()

	job := queue.EmailJob{
		ID:           jobID,
		Type:         req.Type,
		To:           req.To,
		Subject:      req.Subject,
		TemplateName: req.TemplateName,
		Data:         req.Data,
		Attempts:     0,
		MaxAttempts:  3,
		CreatedAt:    time.Now(),
		ScheduledAt:  req.ScheduleAt,
	}

	if err := queue.EnqueueEmail(job); err != nil {
		logrus.Error("Failed to enqueue email:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue email"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Email queued successfully",
		"job_id":  jobID,
	})
}

func SendOTP(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Name  string `json:"name" binding:"required"`
		OTP   string `json:"otp" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job := queue.EmailJob{
		ID:          generateJobID(),
		Type:        "otp",
		To:          req.Email,
		Subject:     "Your OTP Code - Taaza",
		Data:        map[string]interface{}{"name": req.Name, "otp": req.OTP},
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now(),
	}

	if err := queue.EnqueueEmail(job); err != nil {
		logrus.Error("Failed to enqueue OTP email:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue email"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "OTP email queued successfully"})
}

func SendWelcome(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Name  string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job := queue.EmailJob{
		ID:          generateJobID(),
		Type:        "welcome",
		To:          req.Email,
		Subject:     "Welcome to Taaza! 🎉",
		Data:        map[string]interface{}{"name": req.Name},
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now(),
	}

	if err := queue.EnqueueEmail(job); err != nil {
		logrus.Error("Failed to enqueue welcome email:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue email"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Welcome email queued successfully"})
}

func SendProfileReminder(c *gin.Context) {
	var req struct {
		Email string     `json:"email" binding:"required,email"`
		Name  string     `json:"name" binding:"required"`
		Delay *time.Time `json:"delay,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job := queue.EmailJob{
		ID:          generateJobID(),
		Type:        "profile_reminder",
		To:          req.Email,
		Subject:     "Complete Your Taaza Profile",
		Data:        map[string]interface{}{"name": req.Name},
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now(),
		ScheduledAt: req.Delay,
	}

	if err := queue.EnqueueEmail(job); err != nil {
		logrus.Error("Failed to enqueue profile reminder email:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue email"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Profile reminder email queued successfully"})
}

func GetQueueStats(c *gin.Context) {
	// This would return queue statistics
	c.JSON(http.StatusOK, gin.H{
		"queue_length":      "N/A", // Implement Redis queue length check
		"processed_today":   "N/A", // Implement daily counter
		"failed_jobs":       "N/A", // Implement dead letter queue count
		"workers_active":    "N/A", // Implement worker status check
	})
}

func generateJobID() string {
	return "email_" + time.Now().Format("20060102150405") + "_" + randomString(6)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}