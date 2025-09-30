package utils

import (
	"bytes"
	"html/template"
	"net/smtp"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type EmailData struct {
	Name       string
	Email      string
	OTP        string
	AppURL     string
	ProfileURL string
}

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func GetSMTPConfig() SMTPConfig {
	return SMTPConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASS"),
		From:     os.Getenv("SMTP_FROM"),
	}
}

func SendEmail(to, subject, templateName string, data EmailData) error {
	config := GetSMTPConfig()
	
	// Load template
	templatePath := filepath.Join("templates", templateName+".html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		logrus.Error("Failed to parse template:", err)
		return err
	}

	// Execute template
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		logrus.Error("Failed to execute template:", err)
		return err
	}

	// Prepare message
	message := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		body.String())

	// Send email
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
	err = smtp.SendMail(config.Host+":"+config.Port, auth, config.From, []string{to}, message)
	
	if err != nil {
		logrus.Error("Failed to send email:", err)
		return err
	}

	logrus.Infof("Email sent successfully to %s", to)
	return nil
}

func SendOTPEmail(email, name, otp string) error {
	data := EmailData{
		Name:  name,
		Email: email,
		OTP:   otp,
	}
	return SendEmail(email, "Your OTP Code - Taaza", "otp", data)
}

func SendWelcomeEmail(email, name string) error {
	data := EmailData{
		Name:   name,
		Email:  email,
		AppURL: os.Getenv("APP_URL"),
	}
	return SendEmail(email, "Welcome to Taaza! 🎉", "welcome", data)
}

func SendProfileReminderEmail(email, name string) error {
	data := EmailData{
		Name:       name,
		Email:      email,
		ProfileURL: os.Getenv("APP_URL") + "/profile/complete",
	}
	return SendEmail(email, "Complete Your Taaza Profile", "profile-reminder", data)
}