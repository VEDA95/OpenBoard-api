package routes

import (
	"fmt"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/responses"
	"github.com/VEDA95/OpenBoard-API/internal/auth"
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/gofiber/fiber/v2"
)

func handleWebAuthnRequestStart(context *fiber.Ctx, challengeType string) error {
	token, err := auth.ExtractSessionToken(context, "open_board_mfa_challenge")

	if err != nil {
		return err
	}

	multiAuthMethod := *auth.MultiAuthMethods.GetMultiAuthMethod("webauthn")
	payload := util.Payload{
		"token": token,
	}
	challengeResults, err := multiAuthMethod.CreateAuthChallenge(challengeType, payload)

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, challengeResults["options"]))
}

func handleWebAuthnRequestEnd(context *fiber.Ctx, challengeType string) error {
	token, err := auth.ExtractSessionToken(context, "open_board_mfa_challenge")

	if err != nil {
		return err
	}

	multiAuthMethod := *auth.MultiAuthMethods.GetMultiAuthMethod("webauthn")
	payload := util.Payload{
		"token":   token,
		"context": context,
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

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, responsePayload))
}

func CreateWebAuthnAuthMethodStart(context *fiber.Ctx) error {
	return handleWebAuthnRequestStart(context, "register")
}

func CreateWebAuthnAuthMethodEnd(context *fiber.Ctx) error {
	return handleWebAuthnRequestEnd(context, "register")
}

func CreateWebAuthnChallengeStart(context *fiber.Ctx) error {
	return handleWebAuthnRequestStart(context, "login")
}

func CreateWebAuthnChallengeEnd(context *fiber.Ctx) error {
	return handleWebAuthnRequestEnd(context, "login")
}
