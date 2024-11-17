package routes

import (
	"github.com/VEDA95/OpenBoard-API/internal/api/http/responses"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/validators"
	"github.com/VEDA95/OpenBoard-API/internal/auth"
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/VEDA95/OpenBoard-API/internal/settings"
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v2"
	"time"
)

func LocalLogin(context *fiber.Ctx) error {
	validatorData := new(validators.LoginValidator)

	if err := context.BodyParser(validatorData); err != nil {
		return err
	}

	if errs := validators.Instance.Validate(validatorData); errs != nil {
		return util.CreateValidationError(errs)
	}

	authPayload := util.Payload{
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

	settingsInterface := *settings.Instance.GetSettings("auth")
	authSettings := settingsInterface.(*settings.AuthSettings)

	if validatorData.ReturnType == "token" {
		var authResponse interface{}

		if authSettings.MultiAuthEnabled {
			authResponse = responses.OKResponse(fiber.StatusOK, fiber.Map{
				"code":          fiber.StatusOK,
				"mfa_challenge": authResults.AccessToken,
			})

		} else {
			authResponse = responses.OKResponse(fiber.StatusOK, authResults)
		}

		return util.JSONResponse(context, fiber.StatusOK, authResponse)
	}

	if validatorData.ReturnType == "session" {
		context.Status(fiber.StatusOK)

		var cookieName string

		if authSettings.MultiAuthEnabled {
			cookieName = "open_board_mfa_challenge"

		} else {
			cookieName = "open_board_session"
		}

		context.Cookie(&fiber.Cookie{
			Name:     cookieName,
			Value:    authResults.AccessToken,
			Domain:   "localhost:8080",
			HTTPOnly: true,
			Secure:   false,
		})
	}

	return nil
}

func GETMFAMethods(context *fiber.Ctx) error {
	token, err := auth.ExtractSessionToken(context, "open_board_mfa_challenge")

	if err != nil {
		return err
	}

	var userId string
	_, dbErr := db.Instance.From("open_board_multi_auth_challenge").Prepared(true).
		Select("open_user.id").
		Join(goqu.T("open_board_user"), goqu.On(goqu.Ex{
			"open_board_multi_auth_challenge.user_id": "open_board_user.id",
		})).
		As("open_user").
		Where(goqu.Ex{"id": token}).
		ScanVal(&userId)

	if dbErr != nil {
		return dbErr
	}

	if len(userId) == 0 {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}

	var authMethods []auth.MultiAuthMethod
	_, dbErr2 := db.Instance.From("open_board_multi_auth_methods").Prepared(true).
		Select("id", "date_created", "date_updated", "name", "type").
		Where(goqu.Ex{"open_board_multi_auth_method.user_id": userId}).
		ScanStruct(&authMethods)

	if dbErr2 != nil {
		return dbErr2
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, authMethods))
}

func SelectMFAMethod(context *fiber.Ctx) error {
	token, err := auth.ExtractSessionToken(context, "open_board_mfa_challenge")

	if err != nil {
		return err
	}

	var userId string
	_, dbErr := db.Instance.From("open_board_multi_auth_challenge").Prepared(true).
		Select("open_user.id").
		Join(goqu.T("open_board_user"), goqu.On(goqu.Ex{
			"open_board_multi_auth_challenge.user_id": "open_board_user.id",
		})).
		As("open_user").
		Where(goqu.Ex{"id": token}).
		ScanVal(&userId)

	if dbErr != nil {
		return dbErr
	}

	if len(userId) == 0 {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}

	validatorData := new(validators.MFASelectValidator)

	if err := context.BodyParser(&validatorData); err != nil {
		return err
	}

	if errs := validators.Instance.Validate(&validatorData); errs != nil {
		return util.CreateValidationError(errs)
	}

	var methodId string
	_, dbErr2 := db.Instance.From("open_board_multi_auth_method").Prepared(true).
		Select("id").
		Where(goqu.Ex{"user_id": userId, "type": validatorData.MFAType}).
		ScanVal(&methodId)

	if dbErr2 != nil {
		return dbErr2
	}

	if len(methodId) == 0 {
		return fiber.NewError(fiber.StatusNotFound, "MFA method does not exist for user")
	}

	updateChallengeQuery := db.Instance.From("open_board_multi_auth_challenge").Prepared(true).
		Update().
		Where(goqu.Ex{"id": token}).
		Set(goqu.Record{"auth_method_id": methodId, "date_updated": time.Now()}).
		Executor()

	if _, err := updateChallengeQuery.Exec(); err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, responses.GenericMessage{
		Message: "Auth challenge type has been selected...",
	}))
}
