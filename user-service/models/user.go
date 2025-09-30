package models

import (
	"time"
)

type User struct {
	ID                 int       `json:"id" db:"id"`
	Name               string    `json:"name" db:"name"`
	Email              string    `json:"email" db:"email"`
	Mobile             *string   `json:"mobile,omitempty" db:"mobile"`
	Address            *string   `json:"address,omitempty" db:"address"`
	Password           string    `json:"password,omitempty" db:"password_hash"`
	RegistrationStatus string    `json:"registration_status" db:"registration_status"`
	IsVerified         bool      `json:"is_verified" db:"is_verified"`
	ProfileCompleted   bool      `json:"profile_completed" db:"profile_completed"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type UserSession struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type StartRegistrationRequest struct {
	Name  string `json:"name" validate:"required,min=2"`
	Email string `json:"email" validate:"required,email"`
}

type CompleteProfileRequest struct {
	Password string `json:"password" validate:"required,min=6"`
}

type PersonalInfoRequest struct {
	Mobile  string `json:"mobile" validate:"omitempty,len=10"`
	Address string `json:"address"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginWithOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=6"`
}

type UpdateProfileRequest struct {
	Name  string `json:"name" validate:"required,min=2"`
	Email string `json:"email" validate:"required,email"`
}

type LoginVerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=6"`
}