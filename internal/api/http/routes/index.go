package routes

import "github.com/gofiber/fiber/v2"

func HelloWorld(context *fiber.Ctx) error {
	return context.JSON(fiber.Map{"message": "Hello World!"})
}
