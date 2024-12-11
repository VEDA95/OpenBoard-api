package middleware

import (
	"errors"
	"github.com/VEDA95/OpenBoard-API/internal/auth"
	"github.com/gofiber/fiber/v2"
)

func AuthenticationMiddleware(context *fiber.Ctx) error {
	token, err := auth.ExtractSessionToken(context, "open_board_session")

	if err != nil {
		return err
	}

	session, err := auth.CheckAuthSession(token)

	if err != nil {
		return err
	}

	context.Locals("access_token", token)
	context.Locals("session", session)

	return context.Next()
}

func AuthorizationMiddleware(permissions ...string) func(*fiber.Ctx) error {
	return func(context *fiber.Ctx) error {
		session, ok := context.Locals("session").(*auth.AuthSession)

		if !ok {
			return errors.New("user must be authenticated to verify resource access")
		}

		isAuthorized, err := auth.IsUserAuthorized(session.UserId, permissions...)

		if err != nil {
			return err
		}

		if !isAuthorized {
			return fiber.NewError(fiber.StatusUnauthorized, "not authorized")
		}

		return context.Next()
	}
}
