package auth

import (
	"errors"
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/VEDA95/OpenBoard-API/internal/settings"
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/doug-martin/goqu/v9"
	"time"
)

type AuthSession struct {
	Id             string       `json:"id" db:"id,omitempty"`
	UserId         string       `json:"-" db:"user_id,omitempty"`
	User           User         `json:"user" db:"open_user,omitempty"`
	DateCreated    time.Time    `json:"date_created" db:"date_created,omitempty"`
	DateUpdated    time.Time    `json:"date_updated" db:"date_updated,omitempty"`
	ExpiresOn      time.Time    `json:"expires_on" db:"expires_on,omitempty"`
	SessionType    string       `json:"-" db:"session_type,omitempty"`
	RememberMe     bool         `json:"-" db:"remember_me,omitempty"`
	AccessToken    string       `json:"-" db:"access_token,omitempty"`
	RefreshToken   string       `json:"-" db:"refresh_token,omitempty"`
	IPAddress      string       `json:"ip_address" db:"ip_address,omitempty"`
	UserAgent      string       `json:"user_agent" db:"user_agent,omitempty"`
	AdditionalInfo util.Payload `json:"-" db:"additional_info,omitempty"`
}

func CheckAuthSession(token string) (*AuthSession, error) {
	var sessionData AuthSession
	exists, err := db.Instance.From("open_board_user_session").Prepared(true).
		Select("*", "open_user").
		Join(goqu.T("open_board_user"), goqu.On(goqu.Ex{
			"open_board_user_session.user_id": "open_board_user.id",
		})).
		As("open_user").
		Where(goqu.Ex{"access_token": token}).
		ScanStruct(&sessionData)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("session not found")
	}

	if !sessionData.User.Enabled {
		return nil, errors.New("user is disabled")
	}

	now := time.Now()

	if now.After(sessionData.ExpiresOn) {
		return nil, errors.New("session expired")
	}

	return &sessionData, nil
}

func RefreshAuthSession(token string) (*ProviderAuthResult, error) {
	refreshCredentials, err := ExtractRefreshCredential(token)

	if err != nil {
		return nil, err
	}

	var sessionData AuthSession
	exists, err := db.Instance.From("open_board_user_session").Prepared(true).
		Select("*", "open_user", "additional_info").
		Join(goqu.T("open_board_user"), goqu.On(goqu.Ex{
			"open_board_user_session.user_id": "open_board_user.id",
		})).
		As("open_user").
		Where(goqu.Ex{"refresh_token": refreshCredentials.Selector}).
		ScanStruct(&sessionData)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("session not found")
	}

	if !sessionData.User.Enabled {
		return nil, errors.New("user is disabled")
	}

	if !sessionData.RememberMe {
		return nil, errors.New("unable to refresh session")
	}

	now := time.Now()
	settingsInterface := *settings.Instance.GetSettings("auth")
	authSettings := settingsInterface.(*settings.AuthSettings)
	refreshExpiration := sessionData.ExpiresOn.Add(time.Duration(authSettings.RefreshTokenLifetime))
	refreshIdleExpiration := sessionData.User.DateUpdated.Add(time.Duration(authSettings.RefreshTokenIdleLifetime))

	if now.After(refreshExpiration) && now.After(refreshIdleExpiration) {
		return nil, errors.New("session expired")
	}

	if err := CompareValidatorToken(refreshCredentials.Selector, sessionData.AdditionalInfo["hashed_validator"].(string)); err != nil {
		return nil, err
	}

	sessionToken, err := CreateSessionToken()

	if err != nil {
		return nil, err
	}

	credentials, err := CreateRefreshCredentials()

	if err != nil {
		return nil, err
	}

	sessionData.AdditionalInfo["hashed_validator"] = credentials.HashedValidator
	updateSessionQuery := db.Instance.From("open_board_user_session").Prepared(true).
		Update().
		Set(goqu.Record{
			"date_updated":    now,
			"access_token":    sessionToken,
			"refresh_token":   credentials.Selector,
			"additional_info": sessionData.AdditionalInfo,
		}).
		Where(goqu.Ex{"refresh_token": refreshCredentials.Selector}).
		Executor()

	if _, err := updateSessionQuery.Exec(); err != nil {
		return nil, err
	}

	return &ProviderAuthResult{
		AccessToken:  sessionToken,
		RefreshToken: credentials.Selector,
		Validator:    credentials.Validator,
		UserId:       sessionData.UserId,
	}, nil
}
