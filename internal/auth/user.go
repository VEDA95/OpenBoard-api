package auth

import (
	"fmt"
	"github.com/go-webauthn/webauthn/webauthn"
	"time"
)

type User struct {
	Id                   string     `json:"id" db:"id,omitempty"`
	DateCreated          time.Time  `json:"date_created" db:"date_created,omitempty"`
	DateUpdated          *time.Time `json:"date_updated" db:"date_updated,omitempty"`
	LastLogin            *time.Time `json:"last_login" db:"last_login,omitempty"`
	ExternalProviderID   *string    `json:"external_provider_id" db:"external_provider_id,omitempty"`
	Username             string     `json:"username" db:"username,omitempty"`
	Email                string     `json:"email" db:"email,omitempty"`
	HashedPassword       *string    `json:"hashed_password,omitempty" db:"hashed_password,omitempty"`
	FirstName            *string    `json:"first_name" db:"first_name,omitempty"`
	LastName             *string    `json:"last_name" db:"last_name,omitempty"`
	Thumbnail            *string    `json:"thumbnail" db:"thumbnail,omitempty"`
	DarkMode             bool       `json:"dark_mode" default:"false" db:"dark_mode,omitempty"`
	Enabled              bool       `json:"enabled" default:"true" db:"enabled,omitempty"`
	EmailVerified        bool       `json:"email_verified" default:"false" db:"email_verified,omitempty"`
	ResetPasswordOnLogin bool       `json:"reset_password_on_login" default:"false" db:"reset_password_on_login,omitempty"`
}

func (user *User) WebAuthnID() []byte {
	return []byte(user.Id)
}

func (user *User) WebAuthnName() string {
	return user.Username
}

func (user *User) WebAuthnDisplayName() string {
	return fmt.Sprintf("%s %s", *user.FirstName, *user.LastName)
}

func (user *User) WebAuthnCredentials() []webauthn.Credential {
	return []webauthn.Credential{}
}
