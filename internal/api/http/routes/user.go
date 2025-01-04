package routes

import (
	"github.com/VEDA95/OpenBoard-API/internal/api/http/responses"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/validators"
	"github.com/VEDA95/OpenBoard-API/internal/auth"
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v2"
	"strings"
	"time"
)

var defaultUserColumns = []interface{}{
	"id",
	"date_created",
	"date_updated",
	"last_login",
	"external_provider_id",
	"username",
	"email",
	"first_name",
	"last_name",
	"thumbnail",
	"dark_mode",
	"enabled",
	"email_verified",
	"reset_password_on_login",
}

func ShowUsers(context *fiber.Ctx) error {
	localProvider := *auth.Instance.GetProvider("local")
	users, err := localProvider.GetUsers(util.Payload{})

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKCollectionResponse(fiber.StatusOK, *users))
}

func CreateUser(context *fiber.Ctx) error {
	validatorData := new(validators.UserCreate)

	if err := context.BodyParser(&validatorData); err != nil {
		return err
	}

	if errs := validators.Instance.Validate(&validatorData); errs != nil {
		return util.CreateValidationError(errs)
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

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, user))
}

func ShowUser(context *fiber.Ctx) error {
	userId := context.Params("id")
	localProvider := *auth.Instance.GetProvider("local")
	user, err := localProvider.GetUser(util.Payload{"id": userId})

	if err != nil {
		errMsg := err.Error()

		if strings.Contains(errMsg, "not found") {
			return fiber.NewError(fiber.StatusNotFound, errMsg)
		}

		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, user))
}

func UpdateUser(context *fiber.Ctx) error {
	userId := context.Params("id")
	localProvider := *auth.Instance.GetProvider("local")
	_, err := localProvider.GetUser(util.Payload{"id": userId, "columns": []interface{}{"id"}})

	if err != nil {
		return err
	}

	validatorData := new(validators.UserUpdate)

	if err := context.BodyParser(&validatorData); err != nil {
		return err
	}

	if errs := validators.Instance.Validate(&validatorData); errs != nil {
		return util.CreateValidationError(errs)
	}

	updateRecord := make(goqu.Record)

	if validatorData.Username != nil && len(*validatorData.Username) > 0 {
		updateRecord["username"] = *validatorData.Username
	}

	if validatorData.Email != nil && len(*validatorData.Email) > 0 {
		updateRecord["email"] = *validatorData.Email
	}

	if validatorData.FirstName != nil {
		updateRecord["first_name"] = *validatorData.FirstName
	}

	if validatorData.LastName != nil {
		updateRecord["last_name"] = *validatorData.LastName
	}

	if validatorData.Thumbnail != nil && len(*validatorData.Thumbnail) > 0 {
		updateRecord["thumbnail"] = *validatorData.Thumbnail
	}

	if validatorData.DarkMode != nil {
		updateRecord["dark_mode"] = *validatorData.DarkMode
	}

	var user auth.User
	updateUserQuery := db.Instance.From("open_board_user").Prepared(true).
		Update().
		Returning(defaultUserColumns...).
		Set(updateRecord).
		Where(goqu.Ex{"id": userId}).
		Executor()

	if _, err := updateUserQuery.ScanStruct(&user); err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, user))
}

func DeleteUser(context *fiber.Ctx) error {
	userId := context.Params("id")
	localProvider := *auth.Instance.GetProvider("local")
	_, err := localProvider.GetUser(util.Payload{"id": userId, "columns": []interface{}{"id"}})

	if err != nil {
		return err
	}

	deleteUserQuery := db.Instance.From("open_board_user").Prepared(true).
		Where(goqu.Ex{"id": userId}).
		Delete().
		Executor()

	if _, err := deleteUserQuery.Exec(); err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, make(map[string]interface{})))
}

func Me(context *fiber.Ctx) error {
	token, err := auth.ExtractSessionToken(context, "")

	if err != nil {
		return err
	}

	var userId string
	userSessionQuery := db.Instance.From("open_board_user_session").Prepared(true).
		Select("user_id").
		Where(goqu.Ex{"access_token": token}).
		Executor()
	exists, err := userSessionQuery.ScanVal(&userId)

	if err != nil {
		return err
	}

	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}

	localProvider := *auth.Instance.GetProvider("local")
	user, err := localProvider.GetUser(util.Payload{"id": userId})

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, user))
}

func PasswordResetUnlock(context *fiber.Ctx) error {
	token, err := auth.ExtractSessionToken(context, "open_board_session")

	if err != nil {
		return err
	}

	var user auth.User
	exists, err := db.Instance.From("open_board_user_session").Prepared(true).
		Select("open_user.id", "open_user.external_provider_id", "open_user.hashed_password").
		Join(goqu.T("open_board_user"), goqu.On(goqu.Ex{
			"open_board_user_session.user_id": "open_board_user.id",
		})).
		As("open_user").
		Where(goqu.Ex{"access_token": token}).
		ScanStruct(&user)

	if err != nil {
		return err
	}

	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "User does not exist")
	}

	if *user.ExternalProviderID != "" {
		return fiber.NewError(fiber.StatusConflict, "Password cannot be set for user")
	}

	validatorData := new(validators.PasswordResetUnlockValidator)

	if err := context.BodyParser(validatorData); err != nil {
		return err
	}

	if errs := validators.Instance.Validate(validatorData); len(errs) > 0 {
		return util.CreateValidationError(errs)
	}

	match, err := auth.VerifyPassword(*user.HashedPassword, validatorData.Password)

	if err != nil {
		return err
	}

	if !match {
		return fiber.NewError(fiber.StatusConflict, "Password is incorrect")
	}

	resetToken, err := auth.CreateSessionToken()

	if err != nil {
		return err
	}

	tokenExpirationTime := time.Now().Add(time.Minute * 5)
	createResetTokenQuery := db.Instance.From("open_board_user_password_reset_token").Prepared(true).
		Insert().
		Rows(goqu.Record{
			"id":         resetToken,
			"user_id":    user.Id,
			"expires_on": tokenExpirationTime,
		}).
		Executor()

	if _, err := createResetTokenQuery.Exec(); err != nil {
		return err
	}

	passwordResetMessage := "Password reset token has been set"

	if validatorData.ReturnType == "session" {
		context.Cookie(&fiber.Cookie{
			Name:     "open_board_password_reset_token",
			Value:    resetToken,
			Domain:   "localhost:8080",
			HTTPOnly: true,
			Secure:   false,
			Expires:  tokenExpirationTime,
		})

		return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(
			fiber.StatusOK,
			responses.GenericMessage{Message: passwordResetMessage},
		))
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(
		fiber.StatusOK,
		fiber.Map{"message": passwordResetMessage, "token": resetToken},
	))
}

func PasswordReset(context *fiber.Ctx) error {
	validatorData := new(validators.PasswordResetValidator)
	resetTokenCookie := context.Cookies("open_board_password_reset_token")

	if err := context.BodyParser(validatorData); err != nil {
		return err
	}

	if errs := validators.Instance.Validate(validatorData); len(errs) > 0 {
		return util.CreateValidationError(errs)
	}

	if resetTokenCookie == "" && validatorData.Token == "" {
		return fiber.NewError(fiber.StatusBadRequest, "reset token is invalid")
	}

	var resetToken string
	var resetTokenQuery auth.PasswordResetToken

	if validatorData.Token != "" {
		resetToken = validatorData.Token

	} else {
		resetToken = resetTokenCookie
	}

	exists, err := db.Instance.From("open_board_user_password_reset_token").Prepared(true).
		Select("expires_on", "user_id").
		Where(goqu.Ex{"id": resetToken}).
		ScanStruct(&resetTokenQuery)

	if err != nil {
		return err
	}

	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "reset token is invalid")
	}

	now := time.Now()

	if now.After(resetTokenQuery.ExpiresOn) {
		return fiber.NewError(fiber.StatusBadRequest, "reset token is invalid")
	}

	NewHashedPassword, err := auth.HashPassword(validatorData.Password)

	if err != nil {
		return err
	}

	transaction, err := db.Instance.Begin()

	if err != nil {
		return err
	}

	updateUserQuery := db.Instance.From("open_board_user").Prepared(true).
		Update().
		Set(goqu.Record{
			"date_updated":    now,
			"hashed_password": NewHashedPassword,
		}).
		Where(goqu.Ex{"id": resetTokenQuery.UserId}).
		Executor()
	removeSessionsQuery := db.Instance.From("open_board_session").Prepared(true).
		Delete().
		Where(goqu.Ex{"user_id": resetTokenQuery.UserId}).
		Executor()

	if _, err := updateUserQuery.Exec(); err != nil {
		if err := transaction.Rollback(); err != nil {
			return err
		}

		return err
	}

	if _, err := removeSessionsQuery.Exec(); err != nil {
		if err := transaction.Rollback(); err != nil {
			return err
		}

		return err
	}

	if err := transaction.Commit(); err != nil {
		return err
	}

	sessionCookie := context.Cookies("open_board_session")

	if sessionCookie != "" {
		context.ClearCookie("open_board_session")
	}

	if resetTokenCookie != "" {
		context.ClearCookie("open_board_reset_token")
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(
		fiber.StatusOK,
		responses.GenericMessage{Message: "Password has been reset"},
	))
}
