package auth

import (
	"errors"
	"github.com/alexedwards/argon2id"
	"time"
)

type PasswordResetToken struct {
	Id          string    `json:"id" db:"id,omitempty"`
	UserId      string    `json:"user_id" db:"user_id,omitempty"`
	DateCreated time.Time `json:"date_created" db:"date_created,omitempty"`
	ExpiresOn   time.Time `json:"expires_on" db:"expires_on,omitempty"`
}

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)

	if err != nil {
		return "", err
	}

	return hash, nil
}

func VerifyPassword(hash string, password string) (bool, error) {
	if len(hash) == 0 || len(password) == 0 {
		return false, errors.New("password verification failed")
	}

	match, err := argon2id.ComparePasswordAndHash(password, hash)

	if err != nil {
		return false, err
	}

	return match, nil
}
