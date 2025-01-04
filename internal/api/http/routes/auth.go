package routes

import (
	"fmt"
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
			responsePayload := fiber.Map{"access_token": authResults.AccessToken, "mfa_required": authResults.MFARequired}

			if validatorData.RememberMe {
				responsePayload["refresh_token"] = fmt.Sprintf("%s:%s", authResults.RefreshToken, authResults.Validator)
			}

			authResponse = responses.OKResponse(fiber.StatusOK, responsePayload)
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

		if cookieName == "open_board_session" && validatorData.RememberMe {
			context.Cookie(&fiber.Cookie{
				Name:     "remember_me",
				Value:    fmt.Sprintf("%s:%s", authResults.RefreshToken, authResults.Validator),
				Domain:   "localhost:8080",
				HTTPOnly: true,
				Secure:   false,
			})
		}
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

	if errs := validators.Instance.Validate(validatorData); errs != nil {
		return util.CreateValidationError(errs)
	}

	settingsInterface := *settings.Instance.GetSettings("auth")
	authSettings := settingsInterface.(*settings.AuthSettings)

	if validatorData.Skip && !authSettings.ForceMultiAuthEnabled {
		sessionToken, err := auth.CreateSessionToken()

		if err != nil {
			return err
		}

		transaction, err := db.Instance.Begin()

		if err != nil {
			return err
		}

		now := time.Now()
		sessionQuery := transaction.From("open_board_user_session").Prepared(true).
			Insert().
			Rows(goqu.Record{
				"user_id":      userId,
				"expires_on":   now.Local().Add(time.Duration(authSettings.AccessTokenLifetime)),
				"access_token": sessionToken,
				"user_agent":   string(context.Request().Header.UserAgent()[:]),
				"ip_address":   context.IP(),
			}).
			Executor()
		updateUserQuery := transaction.From("open_board_user").Prepared(true).
			Where(goqu.Ex{"id": userId}).
			Update().
			Set(goqu.Record{"last_login": now}).
			Executor()
		deleteQuery := transaction.From("open_board_multi_auth_challenge").Prepared(true).
			Delete().
			Where(goqu.Ex{"id": token}).
			Executor()

		if _, err := sessionQuery.Exec(); err != nil {
			if err := transaction.Rollback(); err != nil {
				return err
			}

			return err
		}

		if _, err := updateUserQuery.Exec(); err != nil {
			if err := transaction.Rollback(); err != nil {
				return err
			}

			return err
		}

		if _, err := deleteQuery.Exec(); err != nil {
			if err := transaction.Rollback(); err != nil {
				return err
			}

			return err
		}

		if err := transaction.Commit(); err != nil {
			return err
		}

		if validatorData.ReturnType == "session" {
			context.Status(fiber.StatusOK)
			context.Cookie(&fiber.Cookie{
				Name:     "open_board_session",
				Value:    sessionToken,
				Domain:   "localhost:8080",
				HTTPOnly: true,
				Secure:   false,
			})

			return nil
		}

		return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, auth.ProviderAuthResult{
			AccessToken: sessionToken,
			UserId:      userId,
		}))
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

	payload := util.Payload{
		"auth_method_id": methodId,
		"date_updated":   time.Now(),
	}

	if err := auth.UpdateMultiAuthChallenge(payload); err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, responses.GenericMessage{
		Message: "Auth challenge type has been selected...",
	}))
}

func LocalRefresh(context *fiber.Ctx) error {
	token, err := auth.ExtractSessionToken(context, "remember_me")

	if err != nil {
		return err
	}

	providerInterface := *auth.Instance.GetProvider("local")
	refreshResults, err := providerInterface.Refresh(util.Payload{"token": token})

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, fiber.Map{
		"access_token":  refreshResults.AccessToken,
		"refresh_token": fmt.Sprintf("%s:%s", refreshResults.RefreshToken, refreshResults.Validator),
		"user_id":       refreshResults.UserId,
	}))
}

func LocalLogout(context *fiber.Ctx) error {
	token, err := auth.ExtractSessionToken(context, "open_board_session")

	if err != nil {
		return err
	}

	validatorData := new(validators.LogoutValidator)

	if err := context.BodyParser(validatorData); err != nil {
		return err
	}

	if errs := validators.Instance.Validate(validatorData); len(errs) > 0 {
		return util.CreateValidationError(errs)
	}

	providerInterface := *auth.Instance.GetProvider("local")

	if err := providerInterface.Logout(util.Payload{"token": token}); err != nil {
		return err
	}

	if validatorData.ReturnType == "session" {
		context.ClearCookie("open_board_session")
		context.ClearCookie("remember_me")
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, fiber.Map{}))
}

func RegisterUser(context *fiber.Ctx) error {
	validatorData := new(validators.UserRegisterValidator)

	if err := context.BodyParser(&validatorData); err != nil {
		return err
	}

	if err := validators.Instance.Validate(validatorData); err != nil {
		return util.CreateValidationError(err)
	}

	userCount := 0
	_, err := db.Instance.From("open_board_user").Prepared(true).
		Select(goqu.COUNT("id")).
		Where(goqu.Ex{"username": validatorData.Username}).
		ScanVal(&userCount)

	if err != nil {
		return err
	}

	if userCount > 0 {
		errorMap := make(validators.ErrorResponseMap)
		errorMap["username"] = &validators.ErrorResponse{
			FailedField: "username",
			Tag:         "",
			Value:       validatorData.Username,
			ErrValue:    "The username provided has already been taken",
		}

		return util.CreateValidationError(errorMap)
	}

	hashedPassword, err := auth.HashPassword(validatorData.Password)

	if err != nil {
		return err
	}

	queryRecord := goqu.Record{
		"username":        validatorData.Username,
		"email":           validatorData.Email,
		"hashed_password": hashedPassword,
	}

	if validatorData.FirstName != nil && len(*validatorData.FirstName) > 0 {
		queryRecord["first_name"] = *validatorData.FirstName
	}

	if validatorData.LastName != nil && len(*validatorData.LastName) > 0 {
		queryRecord["last_name"] = *validatorData.LastName
	}

	var user auth.User
	createUserQuery := db.Instance.From("open_board_user").Prepared(true).
		Insert().
		Returning(defaultUserColumns...).
		Rows(queryRecord).
		Executor()

	if _, err := createUserQuery.ScanStruct(&user); err != nil {
		return err
	}

	sessionToken, err := auth.CreateSessionToken()

	if err != nil {
		return err
	}

	transaction, err := db.Instance.Begin()

	if err != nil {
		return err
	}

	now := time.Now()
	settingsInterface := *settings.Instance.GetSettings("auth")
	authSettings := settingsInterface.(*settings.AuthSettings)
	tokenLifetime := now.Local().Add(time.Duration(authSettings.AccessTokenLifetime))
	sessionPayload := goqu.Record{
		"session_type": "local",
		"user_id":      user.Id,
		"access_token": sessionToken,
		"expires_on":   tokenLifetime,
		"ip_address":   context.IP(),
		"user_agent":   string(context.Request().Header.UserAgent()[:]),
		"remember_me":  false,
	}
	sessionQuery := transaction.From("open_board_user_session").Prepared(true).
		Insert().
		Rows(sessionPayload).
		Executor()
	updateUserQuery := transaction.From("open_board_user").Prepared(true).
		Where(goqu.Ex{"id": user.Id}).
		Update().
		Set(goqu.Record{"last_login": now}).
		Executor()

	if _, err := updateUserQuery.Exec(); err != nil {
		if err := transaction.Rollback(); err != nil {
			return err
		}

		return err
	}

	if _, err := sessionQuery.Exec(); err != nil {
		if err := transaction.Rollback(); err != nil {
			return err
		}

		return err
	}

	if err := transaction.Commit(); err != nil {
		return err
	}

	authResults := &auth.ProviderAuthResult{AccessToken: sessionToken, UserId: user.Id, MFARequired: false}

	if validatorData.ReturnType == "token" {
		return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, fiber.Map{
			"access_token": authResults.AccessToken,
			"mfa_required": authResults.MFARequired,
		}))
	}

	if validatorData.ReturnType == "session" {
		context.Status(fiber.StatusOK)
		context.Cookie(&fiber.Cookie{
			Name:     "open_board_session",
			Value:    authResults.AccessToken,
			Domain:   "localhost",
			HTTPOnly: true,
			Secure:   false,
		})
	}

	return nil
}
