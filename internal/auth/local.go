package auth

import (
	"errors"
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/VEDA95/OpenBoard-API/internal/settings"
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/doug-martin/goqu/v9"
	"time"
)

type LocalAuthProvider struct {
	Config interface{}
}

func (localAuthProvider LocalAuthProvider) Login(payload ProviderPayload) (*ProviderAuthResult, error) {
	username, ok := payload["username"].(string)
	password, secondOk := payload["password"].(string)

	if !ok || !secondOk {
		return nil, errors.New("username or password is missing")
	}

	userQuery := db.Instance.Dialect.Select("id, hashed_password").From("open_board_user").Where(goqu.Ex{"username": username})
	queryResult, err := db.Instance.ExecSingleQuery(userQuery, []string{"id", "hashed_password"})

	if err != nil {
		return nil, err
	}

	hashedPassword, thirdOk := queryResult.Row["hashed_password"].(string)

	if !thirdOk {
		return nil, errors.New("hashed password is missing from userQuery")
	}

	if match, err := VerifyPassword(password, hashedPassword); err != nil || !match {
		return nil, err
	}

	token, err := CreateSessionToken()

	if err != nil {
		return nil, err
	}

	authSettings, err := util.ConvertType[*settings.SettingsInterface, settings.AuthSettings](settings.Instance.GetSettings("auth"))

	if err != nil {
		return nil, err
	}

	now := time.Now()
	userId := queryResult.Row["id"].(string)
	sessionQuery := db.Instance.Dialect.Insert("open_board_user_session").Rows(
		goqu.Record{
			"access_token": token,
			"expires_on":   now.Local().Add(time.Duration(authSettings.AccessTokenLifetime)),
			"ip_address":   payload["ip_address"].(string),
			"user_agent":   payload["user_agent"].(string),
			"remember_me":  payload["remember_me"].(bool),
		},
	)
	updateUserQuery := db.Instance.Dialect.Update("open_board_user").Set(goqu.Record{"last_login": now}).Where(goqu.Ex{"id": userId})

	if _, err := db.Instance.ExecQuery(updateUserQuery); err != nil {
		return nil, err
	}

	if _, err := db.Instance.ExecQuery(sessionQuery); err != nil {
		return nil, err
	}

	return &ProviderAuthResult{AccessToken: token, UserId: userId}, nil
}

func (localAuthProvider LocalAuthProvider) Refresh(payload ProviderPayload) (*ProviderAuthResult, error) {

	return nil, nil
}

func (localAuthProvider LocalAuthProvider) Logout(payload ProviderPayload) error {
	return nil
}

func (localAuthProvider LocalAuthProvider) GetUser(id string) (*User, error) {
	return nil, nil
}
