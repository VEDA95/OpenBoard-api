package middleware

import (
	"github.com/VEDA95/OpenBoard-API/internal/auth"
	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(context *fiber.Ctx) error {
	token, err := auth.ExtractSessionToken(context, "open_board_session")

	if err != nil {
		return err
	}

	if err := auth.CheckAuthSession(token); err != nil {
		return err
	}

	return context.Next()
}
