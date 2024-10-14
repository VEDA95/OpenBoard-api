package routes

import (
	"github.com/VEDA95/OpenBoard-API/internal/api/http/responses"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/validators"
	"github.com/VEDA95/OpenBoard-API/internal/auth"
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/gofiber/fiber/v2"
)

func LocalLogin(context *fiber.Ctx) error {
	validatorData := new(validators.LoginValidator)

	if err := context.BodyParser(validatorData); err != nil {
		return err
	}

	if errs := validators.Instance.Validate(validatorData); errs != nil {
		return util.CreateValidationError(errs)
	}

	authPayload := auth.ProviderPayload{
		"identity":    validatorData.Username,
		"password":    validatorData.Password,
		"remember_me": validatorData.RememberMe,
		"user_agent":  string(context.Request().Header.UserAgent()[:]),
		"ip_address":  context.IP(),
	}
	localProvider := *auth.Instance.GetProvider("local")
	authResults, err := localProvider.Login(authPayload)

	if err != nil {
		return err
	}

	if validatorData.ReturnType == "token" {
		return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, authResults))
	}

	if validatorData.ReturnType == "session" {
		context.Status(fiber.StatusOK)
		context.Cookie(&fiber.Cookie{
			Name:     "open_board_session",
			Value:    authResults.AccessToken,
			Domain:   "localhost:8080",
			HTTPOnly: true,
			Secure:   false,
		})
	}

	return nil
}
