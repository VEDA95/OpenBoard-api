package auth

import (
	"bytes"
	"errors"
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/VEDA95/OpenBoard-API/internal/settings"
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exec"
	"github.com/gofiber/fiber/v2"
	"github.com/pquerna/otp/totp"
	"image/png"
	"time"
)

type AuthenticatorMultiAuth struct{}

func (authenticator *AuthenticatorMultiAuth) CreateAuthChallenge(challengeType string, payload util.Payload) (util.Payload, error) {
	settingsInterface := *settings.Instance.GetSettings("auth")
	authSettings := settingsInterface.(*settings.AuthSettings)

	if !authSettings.MultiAuthEnabled || !authSettings.AuthenticatorEnabled {
		return nil, errors.New("OTP via authenticator is not enabled")
	}

	if challengeType != "register" && challengeType != "login" {
		return nil, errors.New("invalid challenge type")
	}

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

	if challengeType == "login" && challenge.AuthMethod != nil {
		secret, ok := challenge.AuthMethod.CredentialData["secret"].(string)

		if !ok {
			return nil, errors.New("secret is missing")
		}

		updatePayload["data"] = map[string]interface{}{"totp_secret": secret, "type": challengeType}

		if err := UpdateMultiAuthChallenge(updatePayload); err != nil {
			return nil, err
		}

		return nil, nil
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "localhost:8080",
		AccountName: challenge.User.Email,
	})

	if err != nil {
		return nil, err
	}

	image, err := key.Image(200, 200)

	if err != nil {
		return nil, err
	}

	var QRCodeBuffer bytes.Buffer

	if err := png.Encode(&QRCodeBuffer, image); err != nil {
		return nil, err
	}

	updatePayload["data"] = map[string]interface{}{"totp_secret": key.Secret(), "type": challengeType}

	if err := UpdateMultiAuthChallenge(updatePayload); err != nil {
		return nil, err
	}

	return util.Payload{"qr_code": &QRCodeBuffer}, nil
}

func (authenticator *AuthenticatorMultiAuth) VerifyAuthChallenge(challengeType string, payload util.Payload) (*ProviderAuthResult, error) {
	settingsInterface := *settings.Instance.GetSettings("auth")
	authSettings := settingsInterface.(*settings.AuthSettings)

	if !authSettings.MultiAuthEnabled || !authSettings.AuthenticatorEnabled {
		return nil, errors.New("OTP via authenticator is not enabled")
	}

	if challengeType != "register" && challengeType != "login" {
		return nil, errors.New("invalid challenge type")
	}

	token, ok := payload["token"].(string)
	passcode, okTwo := payload["passcode"].(string)

	if !ok {
		return nil, errors.New("token is missing")
	}

	if !okTwo {
		return nil, errors.New("passcode is missing")
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

	secret, ok := challenge.Data["totp_secret"].(string)
	recordedType, okTwo := challenge.Data["type"].(string)

	if !ok {
		return nil, errors.New("secret is missing")
	}

	if !okTwo {
		return nil, errors.New("recorded otp type is missing")
	}

	if recordedType != challengeType {
		return nil, errors.New("invalid otp type")
	}

	if valid := totp.Validate(passcode, secret); !valid {
		return nil, errors.New("invalid passcode")
	}

	transaction, err := db.Instance.Begin()

	if err != nil {
		return nil, err
	}

	var authMethodQuery exec.QueryExecutor

	if challengeType == "register" {
		authMethodQuery = transaction.From("open_board_multi_auth_method").Prepared(true).
			Insert().
			Rows(goqu.Record{
				"user_id":         challenge.User.Id,
				"name":            payload["name"].(string),
				"type":            "authenticator",
				"credential_data": map[string]interface{}{"secret": secret},
			}).
			Executor()
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

		return &ProviderAuthResult{UserId: challenge.User.Id}, nil
	}

	context, ok := payload["context"].(*fiber.Ctx)

	if !ok {
		return nil, errors.New("context is missing")
	}

	sessionToken, err := CreateSessionToken()

	if err != nil {
		return nil, err
	}

	now := time.Now()
	userAgent := string(context.Request().Header.UserAgent()[:])
	ipAddress := context.IP()
	authMethodQuery = transaction.From("open_board_multi_auth_method").Prepared(true).
		Update().
		Where(goqu.Ex{"id": challenge.AuthMethod.Id}).
		Set(goqu.Record{"date_updated": now}).
		Executor()
	sessionQuery := transaction.From("open_board_user_session").Prepared(true).
		Insert().
		Rows(goqu.Record{
			"user_id":      challenge.User.Id,
			"expires_on":   now.Local().Add(time.Duration(authSettings.AccessTokenLifetime)),
			"access_token": sessionToken,
			"user_agent":   userAgent,
			"ip_address":   ipAddress,
		}).
		Executor()
	updateUserQuery := transaction.From("open_board_user").Prepared(true).
		Where(goqu.Ex{"id": challenge.User.Id}).
		Update().
		Set(goqu.Record{"last_login": now}).
		Executor()
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

	if _, err := sessionQuery.Exec(); err != nil {
		if err := transaction.Rollback(); err != nil {
			return nil, err
		}

		return nil, err
	}

	if _, err := updateUserQuery.Exec(); err != nil {
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

	return &ProviderAuthResult{AccessToken: token, UserId: challenge.User.Id}, nil
}
