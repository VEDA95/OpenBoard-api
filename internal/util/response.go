package util

import "github.com/gofiber/fiber/v2"

func JSONResponse(context *fiber.Ctx, code int, data interface{}) error {
	context.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	return context.Status(code).JSON(data)
}
