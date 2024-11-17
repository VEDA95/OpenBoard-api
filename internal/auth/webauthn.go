package auth

import (
	"errors"
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/VEDA95/OpenBoard-API/internal/settings"
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exec"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"time"
)

type WebAuthnMultiAuth struct {
	WebAuthn *webauthn.WebAuthn
}

func NewWebAuthnMultiAuth() (*WebAuthnMultiAuth, error) {
	config := webauthn.Config{}
	webAuthn, err := webauthn.New(&config)

	if err != nil {
		return nil, err
	}

	return &WebAuthnMultiAuth{WebAuthn: webAuthn}, nil
}

func (webAuthnMultiAuth *WebAuthnMultiAuth) CreateAuthChallenge(challengeType string, payload util.Payload) (util.Payload, error) {
	token, ok := payload["token"].(string)

	if !ok {
		return nil, errors.New("token is missing")
	}

	var challenge MultiAuthChallenge
	_, err := db.Instance.From("open_board_multi_auth_challenge").Prepared(true).
		Select("open_user", "auth_method").
		Join(goqu.T("open_board_user"), goqu.On(goqu.Ex{
			"open_board_multi_auth_challenge.user_id": "open_board_user.id",
		})).
		As("open_user").
		Join(goqu.T("open_board_multi_auth_method"), goqu.On(goqu.Ex{
			"open_board_multi_auth_challenge.auth_method_id": "open_board_multi_auth_method.id",
		})).
		As("auth_method").
		Where(goqu.Ex{"id": token}).
		ScanStruct(&challenge)

	if err != nil {
		return nil, err
	}

	updatePayload := util.Payload{"token": token}
	outputPayload := util.Payload{}

	if challengeType == "register" {
		options, session, err := webAuthnMultiAuth.WebAuthn.BeginRegistration(&challenge.User)

		if err != nil {
			return nil, err
		}

		outputPayload["options"] = options
		updatePayload["data"] = map[string]interface{}{"webauthn_session": session}

	} else if challengeType == "login" {
		if challenge.AuthMethod == nil || challenge.AuthMethod.Type != "webauthn" {
			return nil, errors.New("invalid challenge")
		}

		options, session, err := webAuthnMultiAuth.WebAuthn.BeginLogin(&challenge.User)

		if err != nil {
			return nil, err
		}

		outputPayload["options"] = options
		updatePayload["data"] = map[string]interface{}{"webauthn_session": session}

	} else {
		return nil, errors.New("invalid challenge type")
	}

	if err := UpdateMultiAuthChallenge(updatePayload); err != nil {
		return nil, err
	}

	return outputPayload, nil
}

func (webAuthnMultiAuth *WebAuthnMultiAuth) VerifyAuthChallenge(challengeType string, payload util.Payload) (*ProviderAuthResult, error) {
	token, ok := payload["token"].(string)
	context, okTwo := payload["context"].(*fiber.Ctx)

	if !ok {
		return nil, errors.New("token is missing")
	}

	if !okTwo {
		return nil, errors.New("fiber context is missing")
	}

	userAgent := string(context.Request().Header.UserAgent()[:])
	ipAddress := context.IP()
	convertedContext, err := adaptor.ConvertRequest(context, false)

	if err != nil {
		return nil, err
	}

	var challenge MultiAuthChallenge
	_, dbErr := db.Instance.From("open_board_multi_auth_challenge").Prepared(true).
		Select("open_user", "auth_method", "data").
		Join(goqu.T("open_board_user"), goqu.On(goqu.Ex{
			"open_board_multi_auth_challenge.user_id": "open_board_user.id",
		})).
		As("open_user").
		Join(goqu.T("open_board_multi_auth_method"), goqu.On(goqu.Ex{
			"open_board_multi_auth_challenge.auth_method_id": "open_board_multi_auth_method.id",
		})).
		As("auth_method").
		Where(goqu.Ex{"id": token}).
		ScanStruct(&challenge)

	if dbErr != nil {
		return nil, dbErr
	}

	sessionData, ok := challenge.Data["webauthn_session"].(webauthn.SessionData)

	if !ok {
		return nil, errors.New("webauthn_session is missing")
	}

	var webAuthnCredential *webauthn.Credential

	if challengeType == "register" {
		credential, err := webAuthnMultiAuth.WebAuthn.FinishRegistration(&challenge.User, sessionData, convertedContext)

		if err != nil {
			return nil, err
		}

		webAuthnCredential = credential

	} else if challengeType == "login" {
		if challenge.AuthMethod == nil || challenge.AuthMethod.Type != "webauthn" {
			return nil, errors.New("invalid challenge")
		}

		credential, err := webAuthnMultiAuth.WebAuthn.FinishLogin(&challenge.User, sessionData, convertedContext)

		if err != nil {
			return nil, err
		}

		webAuthnCredential = credential

	} else {
		return nil, errors.New("invalid challenge type")
	}

	var sessionToken string
	transaction, err := db.Instance.Begin()

	if err != nil {
		return nil, err
	}

	settingInterface := *settings.Instance.GetSettings("auth")
	authSettings := settingInterface.(*settings.AuthSettings)
	var authMethodQuery exec.QueryExecutor

	if challengeType == "register" {
		authMethodQuery = transaction.From("open_board_multi_auth_method").Prepared(true).
			Insert().
			Rows(goqu.Record{
				"user_id":         challenge.User.Id,
				"name":            payload["name"].(string),
				"type":            "webauthn",
				"credential_data": webAuthnCredential,
			}).
			Executor()
	} else {
		seshToken, err := CreateSessionToken()

		if err != nil {
			return nil, err
		}

		sessionToken = seshToken
		authMethodQuery = transaction.From("open_board_multi_auth_method").Prepared(true).
			Update().
			Where(goqu.Ex{"id": challenge.AuthMethod.Id}).
			Set(goqu.Record{"credential_data": webAuthnCredential}).
			Executor()
		sessionQuery := transaction.From("open_board_user_session").Prepared(true).
			Insert().
			Rows(goqu.Record{
				"user_id":      challenge.User.Id,
				"expires_on":   time.Now().Local().Add(time.Duration(authSettings.AccessTokenLifetime)),
				"access_token": sessionToken,
				"user_agent":   userAgent,
				"ip_address":   ipAddress,
			}).
			Executor()

		if _, err := sessionQuery.Exec(); err != nil {
			if err := transaction.Rollback(); err != nil {
				return nil, err
			}

			return nil, err
		}
	}

	deleteQuery := transaction.From("open_board_multi_auth_challenge").Prepared(true).
		Delete().
		Where(goqu.Ex{"id": token}).
		Executor()

	if _, err := authMethodQuery.Exec(); err != nil {
		if err := transaction.Rollback(); err != nil {
			return nil, err
		}

		return nil, err
	}

	if _, err := deleteQuery.Exec(); err != nil {
		if err := transaction.Rollback(); err != nil {
			return nil, err
		}

		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		return nil, err
	}

	return &ProviderAuthResult{AccessToken: sessionToken, UserId: challenge.User.Id}, nil
}
