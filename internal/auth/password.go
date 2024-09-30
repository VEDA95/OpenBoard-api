package auth

import (
	"errors"
	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (*string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)

	if err != nil {
		return nil, err
	}

	return &hash, nil
}

func VerifyPassword(hash string, password string) (bool, error) {
	if len(hash) == 0 || len(password) == 0 {
		return false, errors.New("password verification failed")
	}

	match, err := argon2id.ComparePasswordAndHash(hash, password)

	if err != nil {
		return false, err
	}

	return match, nil
}
