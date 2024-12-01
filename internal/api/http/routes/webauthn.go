package routes

import (
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

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, challengeResults))
}

func CreateWebAuthnAuthMethodStart(context *fiber.Ctx) error {
	return handleWebAuthnRequestStart(context, "register")
}

func CreateWebAuthnAuthMethodEnd(context *fiber.Ctx) error {
	return handleWebAuthnRequestEnd(context, "register")
}

func CreateMultiAuthChallengeStart(context *fiber.Ctx) error {
	return handleWebAuthnRequestStart(context, "login")
}

func CreateWebAuthnAuthChallengeEnd(context *fiber.Ctx) error {
	return handleWebAuthnRequestEnd(context, "login")
}
