package auth

import "time"

type User struct {
	Id                   string     `json:"id"`
	DateCreated          time.Time  `json:"date_created"`
	DateUpdated          *time.Time `json:"date_updated,omitempty"`
	LastLogin            *time.Time `json:"last_login,omitempty"`
	Username             string     `json:"username"`
	Email                string     `json:"email"`
	FirstName            *string    `json:"first_name,omitempty"`
	LastName             *string    `json:"last_name,omitempty"`
	DarkMode             bool       `json:"dark_mode" default:"false"`
	Enabled              bool       `json:"enabled" default:"true"`
	EmailVerified        bool       `json:"email_verified" default:"false"`
	ResetPasswordOnLogin bool       `json:"reset_password_on_login" default:"false"`
}
