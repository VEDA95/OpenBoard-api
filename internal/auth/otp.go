package auth

import (
	"errors"
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/VEDA95/OpenBoard-API/internal/settings"
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exec"
	"github.com/gofiber/fiber/v2"
	"time"
)

type OTPMultiAuth struct{}

func (otpMultiAuth *OTPMultiAuth) CreateAuthChallenge(challengeType string, payload util.Payload) (util.Payload, error) {
	settingsInterface := *settings.Instance.GetSettings("auth")
	authSettings := settingsInterface.(*settings.AuthSettings)

	if !authSettings.OTPEnabled {
		return nil, errors.New("OTP via email or text is not enabled")
	}

	if challengeType != "register" && challengeType != "login" {
		return nil, errors.New("invalid challenge type")
	}

	token, ok := payload["token"].(string)
	transportMethod, okTwo := payload["transport_method"].(string)

	if !ok {
		return nil, errors.New("token is missing")
	}

	if !okTwo || (transportMethod != "sms" && transportMethod != "email") {
		return nil, errors.New("invalid transport method")
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

	otp, err := util.GenerateOTP(6)

	if err != nil {
		return nil, err
	}

	updatePayload := util.Payload{
		"token": token,
		"data":  map[string]interface{}{"otp": otp, "type": challengeType},
	}

	if err := UpdateMultiAuthChallenge(updatePayload); err != nil {
		return nil, err
	}

	secondSettingsInterface := *settings.Instance.GetSettings("notification")
	notificationSettings := settingsInterface.(*settings.NotificationSettings)

	if transportMethod == "sms" {
	}

	if transportMethod == "email" {
	}

	return nil, nil
}

func (otpMultiAuth *OTPMultiAuth) VerifyAuthChallenge(challengeType string, payload util.Payload) (*ProviderAuthResult, error) {
	settingsInterface := *settings.Instance.GetSettings("auth")
	authSettings := settingsInterface.(*settings.AuthSettings)

	if !authSettings.OTPEnabled {
		return nil, errors.New("OTP via email or text is not enabled")
	}

	if challengeType != "register" && challengeType != "login" {
		return nil, errors.New("invalid challenge type")
	}

	token, ok := payload["token"].(string)
	otp, okTwo := payload["otp"].(string)
	context, okThree := payload["context"].(*fiber.Ctx)

	if !ok {
		return nil, errors.New("token is missing")
	}

	if !okTwo {
		return nil, errors.New("otp is missing")
	}

	if !okThree {
		return nil, errors.New("fiber context is missing")
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

	recordedOTP, ok := challenge.Data["otp"].(string)
	otpType, okTwo := challenge.Data["type"].(string)

	if !ok {
		return nil, errors.New(" recorded otp is missing")
	}

	if !okTwo {
		return nil, errors.New("recorded otp type is missing")
	}

	if otpType != challengeType {
		return nil, errors.New("invalid otp type")
	}

	if recordedOTP != otp {
		return nil, errors.New("invalid otp")
	}

	var sessionToken string
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
				"type":            "otp",
				"credential_data": map[string]interface{}{},
			}).
			Executor()

	} else {
		seshToken, err := CreateSessionToken()

		if err != nil {
			return nil, err
		}

		now := time.Now()
		userAgent := string(context.Request().Header.UserAgent()[:])
		ipAddress := context.IP()
		sessionToken = seshToken
		authMethodQuery = transaction.From("open_board_multi_auth_method").Prepared(true).
			Update().
			Where(goqu.Ex{"id": challenge.AuthMethod.Id}).
			Set(goqu.Record{"credential_data": map[string]interface{}{}}).
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

	return nil, nil
}
