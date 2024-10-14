package auth

import (
	"errors"
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/VEDA95/OpenBoard-API/internal/settings"
	"github.com/doug-martin/goqu/v9"
	"log"
	"time"
)

var defaultColumns = []interface{}{
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

type LocalAuthProvider struct {
	Config interface{}
}

func (localAuthProvider *LocalAuthProvider) Login(payload ProviderPayload) (*ProviderAuthResult, error) {
	identity, ok := payload["identity"].(string)
	password, secondOk := payload["password"].(string)

	if !ok || !secondOk {
		return nil, errors.New("identity or password is missing")
	}

	var user User
	userQuery := db.Instance.From("open_board_user").Prepared(true).Select("id", "hashed_password").Where(
		goqu.Or(
			goqu.C("username").Eq(identity),
			goqu.C("email").Eq(identity),
		),
	)
	exists, err := userQuery.ScanStruct(&user)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("user not found")
	}

	if match, err := VerifyPassword(*user.HashedPassword, password); err != nil || !match {
		return nil, err
	}

	token, err := CreateSessionToken()

	if err != nil {
		return nil, err
	}

	now := time.Now()
	settingsInterface := *settings.Instance.GetSettings("auth")
	authSettings := settingsInterface.(*settings.AuthSettings)
	transaction, err := db.Instance.Begin()

	if err != nil {
		return nil, err
	}

	sessionQuery := transaction.From("open_board_user_session").Prepared(true).Insert().Rows(goqu.Record{
		"user_id":      user.Id,
		"access_token": token,
		"expires_on":   now.Local().Add(time.Duration(authSettings.AccessTokenLifetime)),
		"ip_address":   payload["ip_address"].(string),
		"user_agent":   payload["user_agent"].(string),
		"remember_me":  payload["remember_me"].(bool),
	}).Executor()
	updateUserQuery := transaction.From("open_board_user").Prepared(true).
		Where(goqu.Ex{"id": user.Id}).
		Update().
		Set(goqu.Record{"last_login": now}).
		Executor()

	if _, err := updateUserQuery.Exec(); err != nil {
		if err := transaction.Rollback(); err != nil {
			return nil, err
		}

		return nil, err
	}

	if _, err := sessionQuery.Exec(); err != nil {
		if err := transaction.Rollback(); err != nil {
			return nil, err
		}

		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		return nil, err
	}

	return &ProviderAuthResult{AccessToken: token, UserId: user.Id}, nil
}

func (localAuthProvider *LocalAuthProvider) Logout(payload ProviderPayload) error {
	return nil
}

func (localAuthProvider *LocalAuthProvider) GetUser(payload ProviderPayload) (*User, error) {
	columns, ok := payload["columns"].([]interface{})

	if !ok {
		columns = defaultColumns

	}

	var userQuery *goqu.SelectDataset
	id, ok := payload["id"].(string)

	if !ok {
		log.Print(payload)
		identity, ok := payload["identity"].(string)

		if !ok {
			return nil, errors.New("id is missing")
		}

		userQuery = db.Instance.From("open_board_user").Prepared(true).Select(columns...).Where(
			goqu.Or(
				goqu.C("username").Eq(identity),
				goqu.C("email").Eq(identity),
			),
		)

	} else {
		userQuery = db.Instance.From("open_board_user").Prepared(true).Select(columns...).Where(goqu.Ex{"id": id})
	}

	var user User
	exists, err := userQuery.ScanStruct(&user)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("user not found")
	}

	return &user, nil
}

func (localAuthProvider *LocalAuthProvider) GetUsers(payload ProviderPayload) (*[]User, error) {
	columns, ok := payload["columns"].([]interface{})

	if !ok {
		columns = defaultColumns
	}

	var users []User
	usersQuery := db.Instance.Select(columns...).From("open_board_user")

	if err := usersQuery.ScanStructs(&users); err != nil {
		return nil, err
	}

	return &users, nil
}
