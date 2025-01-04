package routes

import (
	"fmt"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/responses"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/validators"
	"github.com/VEDA95/OpenBoard-API/internal/auth"
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/gofiber/fiber/v2"
)

func handleAuthenticatorStart(context *fiber.Ctx, challengeType string) error {
	token, err := auth.ExtractSessionToken(context, "open_board_mfa_challenge")

	if err != nil {
		return err
	}

	multiAuthMethod := *auth.MultiAuthMethods.GetMultiAuthMethod("authenticator")
	payload := util.Payload{
		"token": token,
	}
	challengeResults, err := multiAuthMethod.CreateAuthChallenge(challengeType, payload)

	if err != nil {
		return err
	}

	if challengeType == "login" {
		return util.JSONResponse(context, fiber.StatusOK, nil)
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, challengeResults))
}

func handleAuthenticatorEnd(context *fiber.Ctx, challengeType string) error {
	token, err := auth.ExtractSessionToken(context, "open_board_mfa_challenge")

	if err != nil {
		return err
	}

	inputData := new(validators.AuthenticatorValidator)

	if err := context.BodyParser(inputData); err != nil {
		return err
	}

	if errs := validators.Instance.Validate(inputData); errs != nil {
		return util.CreateValidationError(errs)
	}

	multiAuthMethod := *auth.MultiAuthMethods.GetMultiAuthMethod("authenticator")
	payload := util.Payload{
		"token":    token,
		"passcode": inputData.Passcode,
		"context":  context,
	}
	challengeResults, err := multiAuthMethod.VerifyAuthChallenge(challengeType, payload)

	if err != nil {
		return err
	}

	responsePayload := fiber.Map{"user_id": challengeResults.UserId}

	if challengeResults.AccessToken != "" {
		responsePayload["access_token"] = challengeResults.AccessToken
	}

	if challengeResults.RefreshToken != "" && challengeResults.Validator != "" {
		responsePayload["refresh_token"] = fmt.Sprintf("%s:%s", challengeResults.RefreshToken, challengeResults.Validator)
	}

	context.ClearCookie("open_board_mfa_challenge")

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, responsePayload))
}

func CreateAuthenticatorAuthMethodStart(context *fiber.Ctx) error {
	return handleAuthenticatorStart(context, "register")
}

func CreateAuthenticatorAuthMethodEnd(context *fiber.Ctx) error {
	return handleAuthenticatorEnd(context, "register")
}

func CreateAuthenticatorChallengeStart(context *fiber.Ctx) error {
	return handleAuthenticatorStart(context, "login")
}

func CreateAuthenticatorChallengeEnd(context *fiber.Ctx) error {
	return handleAuthenticatorEnd(context, "login")
}
