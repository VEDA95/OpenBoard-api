package auth

import (
	"encoding/base64"
	"errors"
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/gofiber/fiber/v2"
	"strings"
)

func CreateSessionToken() (string, error) {
	token, err := util.GenerateRandomBytes(32)

	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(token), nil
}

func ExtractSessionToken(context *fiber.Ctx, cookieField string) (string, error) {
	tokenHeader := string(context.Request().Header.Peek("Authorization")[:])
	splitHeader := strings.Split(tokenHeader, " ")

	if len(cookieField) == 0 {
		cookieField = "open_board_session"
	}

	if len(splitHeader) != 2 {
		sessionCookie := context.Cookies(cookieField)

		if len(sessionCookie) == 0 {
			return "", errors.New("no session cookie found")
		}

		return sessionCookie, nil
	}

	return splitHeader[1], nil
}
