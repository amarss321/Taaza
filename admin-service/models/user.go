package models

import "time"

type User struct {
	ID                 int        `json:"id" db:"id"`
	Name               string     `json:"name" db:"name"`
	Email              string     `json:"email" db:"email"`
	Mobile             *string    `json:"mobile,omitempty" db:"mobile"`
	Gender             *string    `json:"gender,omitempty" db:"gender"`
	DateOfBirth        *time.Time `json:"date_of_birth,omitempty" db:"date_of_birth"`
	RegistrationStatus string     `json:"registration_status" db:"registration_status"`
	IsVerified         bool       `json:"is_verified" db:"is_verified"`
	ProfileCompleted   bool       `json:"profile_completed" db:"profile_completed"`
	Status             string     `json:"status" db:"status"`
	LastLogin          *time.Time `json:"last_login,omitempty" db:"last_login"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

type UserStats struct {
	TotalUsers        int `json:"total_users"`
	ActiveUsers       int `json:"active_users"`
	BlockedUsers      int `json:"blocked_users"`
	VerifiedUsers     int `json:"verified_users"`
	NewUsersToday     int `json:"new_users_today"`
	CompletedProfiles int `json:"completed_profiles"`
}

type UserActivity struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Action    string    `json:"action" db:"action"`
	Details   string    `json:"details" db:"details"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type UpdateUserRequest struct {
	Status     *string `json:"status,omitempty"`
	IsVerified *bool   `json:"is_verified,omitempty"`
	Reason     string  `json:"reason,omitempty"`
}