package routes

import (
	"github.com/VEDA95/OpenBoard-API/internal/api/http/responses"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/validators"
	"github.com/VEDA95/OpenBoard-API/internal/auth"
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/gofiber/fiber/v2"
)

func handleOTPStart(context *fiber.Ctx, challengeType string) error {
	token, err := auth.ExtractSessionToken(context, "open_board_mfa_challenge")

	if err != nil {
		return err
	}

	multiAuthMethod := *auth.MultiAuthMethods.GetMultiAuthMethod("otp")
	payload := util.Payload{
		"token": token,
	}
	_, authErr := multiAuthMethod.CreateAuthChallenge(challengeType, payload)

	if authErr != nil {
		return authErr
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, responses.GenericMessage{Message: "OTP sent successfully!"}))
}

func handleOTPEnd(context *fiber.Ctx, challengeType string) error {
	token, err := auth.ExtractSessionToken(context, "open_board_mfa_challenge")

	if err != nil {
		return err
	}

	inputData := new(validators.OTPValidator)

	if err := context.BodyParser(&inputData); err != nil {
		return err
	}

	if errs := validators.Instance.Validate(inputData); errs != nil {
		return util.CreateValidationError(errs)
	}

	multiAuthMethod := *auth.MultiAuthMethods.GetMultiAuthMethod("otp")
	payload := util.Payload{
		"token":   token,
		"otp":     inputData.Otp,
		"context": context,
	}
	challengeResults, err := multiAuthMethod.VerifyAuthChallenge(challengeType, payload)

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, challengeResults))
}

func CreateOTPAuthMethodStart(context *fiber.Ctx) error {
	return handleOTPStart(context, "register")
}

func CreateOTPAuthMethodEnd(context *fiber.Ctx) error {
	return handleOTPEnd(context, "register")
}

func CreateOTPChallengeStart(context *fiber.Ctx) error {
	return handleOTPStart(context, "login")
}

func CreateOTPChallengeEnd(context *fiber.Ctx) error {
	return handleOTPEnd(context, "login")
}
