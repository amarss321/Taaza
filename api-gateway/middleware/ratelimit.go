package middleware

import (
	"api-gateway/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func OTPRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			c.Abort()
			return
		}

		// Restore the body for subsequent handlers
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse the JSON to get email
		var requestBody struct {
			Email string `json:"email"`
		}
		if err := json.Unmarshal(bodyBytes, &requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			c.Abort()
			return
		}

		// Rate limit: 3 OTP requests per email per 10 minutes
		key := fmt.Sprintf("otp_rate_limit:%s", requestBody.Email)
		if !utils.CheckRateLimit(key, 3, 10*time.Minute) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many OTP requests. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}