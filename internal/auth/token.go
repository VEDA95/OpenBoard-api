package auth

import (
	"encoding/base64"
	"github.com/VEDA95/OpenBoard-API/internal/util"
)

func CreateSessionToken() (string, error) {
	token, err := util.GenerateRandomBytes(32)

	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(token), nil
}
