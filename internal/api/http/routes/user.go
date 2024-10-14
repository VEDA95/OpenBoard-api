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
	users, err := localProvider.GetUsers(auth.ProviderPayload{})

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
	user, err := localProvider.GetUser(auth.ProviderPayload{"id": userId})

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
	_, err := localProvider.GetUser(auth.ProviderPayload{"id": userId, "columns": []interface{}{"id"}})

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
	_, err := localProvider.GetUser(auth.ProviderPayload{"id": userId, "columns": []interface{}{"id"}})

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
	tokenHeader := string(context.Request().Header.Peek("Authorization")[:])
	splitHeader := strings.Split(tokenHeader, " ")
	var token string

	if len(splitHeader) != 2 {
		sessionCookie := context.Cookies("open_board_session")

		if len(sessionCookie) == 0 {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		token = sessionCookie

	} else {
		token = splitHeader[1]
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
	user, err := localProvider.GetUser(auth.ProviderPayload{"id": userId})

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, user))
}
