package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

func SendOTPEmail(email, name, otp string) error {
	emailServiceURL := os.Getenv("EMAIL_SERVICE_URL")
	if emailServiceURL == "" {
		emailServiceURL = "http://email-service:8084"
	}

	payload := map[string]interface{}{
		"email": email,
		"name":  name,
		"otp":   otp,
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(emailServiceURL+"/api/v1/email/otp", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logrus.Error("Failed to send OTP email:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		logrus.Error("Email service returned error:", resp.Status)
		return fmt.Errorf("email service error: %s", resp.Status)
	}

	logrus.Infof("OTP email queued for %s", email)
	return nil
}