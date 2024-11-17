package routes

import (
	"github.com/VEDA95/OpenBoard-API/internal/api/http/responses"
	"github.com/VEDA95/OpenBoard-API/internal/auth"
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/gofiber/fiber/v2"
)

func CreateMultiAuthMethodStart(context *fiber.Ctx) error {
	token, err := auth.ExtractSessionToken(context, "open_board_mfa_challenge")

	if err != nil {
		return err
	}

	multiAuthMethod := *auth.MultiAuthMethods.GetMultiAuthMethod("webauthn")
	payload := util.Payload{
		"token": token,
	}
	challengeResults, err := multiAuthMethod.CreateAuthChallenge("register", payload)

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, challengeResults["options"]))
}

func CreateMultiAuthMethodEnd(context *fiber.Ctx) error {
	token, err := auth.ExtractSessionToken(context, "open_board_mfa_challenge")

	if err != nil {
		return err
	}

	multiAuthMethod := *auth.MultiAuthMethods.GetMultiAuthMethod("webauthn")
	payload := util.Payload{
		"token":   token,
		"context": context,
	}
	challengeResults, err := multiAuthMethod.VerifyAuthChallenge("register", payload)

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, challengeResults))
}

func CreateMultiAuthChallengeStart(context *fiber.Ctx) error {
	token, err := auth.ExtractSessionToken(context, "open_board_mfa_challenge")

	if err != nil {
		return err
	}

	multiAuthMethod := *auth.MultiAuthMethods.GetMultiAuthMethod("webauthn")
	payload := util.Payload{
		"token": token,
	}
	challengeResults, err := multiAuthMethod.CreateAuthChallenge("login", payload)

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, challengeResults["options"]))
}

func CreateMultiAuthChallengeEnd(context *fiber.Ctx) error {
	token, err := auth.ExtractSessionToken(context, "open_board_mfa_challenge")

	if err != nil {
		return err
	}

	multiAuthMethod := *auth.MultiAuthMethods.GetMultiAuthMethod("webauthn")
	payload := util.Payload{
		"token":   token,
		"context": context,
	}
	challengeResults, err := multiAuthMethod.VerifyAuthChallenge("login", payload)

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, challengeResults))
}
