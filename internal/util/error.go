package util

import (
	"errors"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/responses"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/validators"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"strings"
)

func ErrorHandler(context *fiber.Ctx, err error) error {
	context.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	code := fiber.StatusInternalServerError
	var fiberErr *fiber.Error

	if errors.As(err, &fiberErr) {
		code = fiberErr.Code
	}

	errorString := err.Error()
	runes := []rune(errorString)

	if strings.Contains(string(runes[0]), "{") && strings.Contains(string(runes[len(runes)-1]), "}") {
		validationErrs := make(validators.ErrorResponseMap)
		err := json.Unmarshal([]byte(errorString), &validationErrs)

		if err == nil {
			return context.
				Status(fiber.StatusUnprocessableEntity).
				JSON(responses.ErrorResp(fiber.StatusUnprocessableEntity, validationErrs))
		}

		errorString = err.Error()
	}

	return context.Status(code).JSON(responses.ErrorRespMessage(code, errorString))
}

func CreateValidationError(errs validators.ErrorResponseMap) error {
	serializedData, err := json.Marshal(errs)

	if err != nil {
		return err
	}

	return errors.New(string(serializedData))
}
