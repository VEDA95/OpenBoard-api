package settings

import (
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/doug-martin/goqu/v9"
)

type AuthSettings struct {
	AccessTokenLifetime      int64 `db:"access_token_lifetime"`
	RefreshTokenLifetime     int64 `db:"refresh_token_lifetime"`
	RefreshTokenIdleLifetime int64 `db:"refresh_token_idle_lifetime"`
	MultiAuthEnabled         bool  `db:"multi_factor_auth_enabled"`
	ForceMultiAuthEnabled    bool  `db:"force_multi_factor_auth"`
	OTPEnabled               bool  `db:"otp_enabled"`
	AuthenticatorEnabled     bool  `db:"authenticator_enabled"`
	WebAuthnEnabled          bool  `db:"webauthn_enabled"`
}

func (authSettings *AuthSettings) Load() error {
	authSettingQuery := db.Instance.Select("*").From("open_board_auth_settings")
	exists, err := authSettingQuery.ScanStruct(authSettings)

	if err != nil {
		return err
	}

	if !exists {
		accessTokenLifetime := 3600
		refreshTokenLifetime := 7200
		refreshTokenIdleLifetime := 1209600
		createAuthSettingsQuery := db.Instance.From("open_board_auth_settings").Prepared(true).
			Insert().
			Rows(
				goqu.Record{
					"access_token_lifetime":       accessTokenLifetime,
					"refresh_token_lifetime":      refreshTokenLifetime,
					"refresh_token_idle_lifetime": refreshTokenIdleLifetime,
					"multi_factor_auth_enabled":   true,
					"force_multi_factor_auth":     false,
					"otp_enabled":                 true,
					"authenticator_enabled":       true,
					"webauthn_enabled":            true,
				},
			).Executor()

		if _, err := createAuthSettingsQuery.Exec(); err != nil {
			return err
		}

		authSettings.AccessTokenLifetime = int64(accessTokenLifetime)
		authSettings.RefreshTokenLifetime = int64(refreshTokenLifetime)
		authSettings.RefreshTokenIdleLifetime = int64(refreshTokenIdleLifetime)
		authSettings.MultiAuthEnabled = true
		authSettings.ForceMultiAuthEnabled = false
		authSettings.OTPEnabled = true
		authSettings.AuthenticatorEnabled = true
		authSettings.WebAuthnEnabled = true

	}

	return nil
}

func (authSettings *AuthSettings) Save() error {
	authSettingsUpdateQuery := db.Instance.From("open_board_auth_settings").Prepared(true).
		Update().
		Set(
			goqu.Record{
				"access_token_lifetime":       authSettings.AccessTokenLifetime,
				"refresh_token_lifetime":      authSettings.RefreshTokenLifetime,
				"refresh_token_idle_lifetime": authSettings.RefreshTokenIdleLifetime,
				"multi_factor_auth_enabled":   authSettings.MultiAuthEnabled,
				"force_multi_factor_auth":     authSettings.ForceMultiAuthEnabled,
				"otp_enabled":                 authSettings.OTPEnabled,
				"authenticator_enabled":       authSettings.AuthenticatorEnabled,
				"webauthn_enabled":            authSettings.WebAuthnEnabled,
			},
		).Executor()

	if _, err := authSettingsUpdateQuery.Exec(); err != nil {
		return err
	}

	return nil
}
