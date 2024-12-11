package settings

import (
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/doug-martin/goqu/v9"
)

type AuthSettings struct {
	AccessTokenLifetime      int64 `db:"access_token_lifetime" validate:"number,omitempty"`
	RefreshTokenLifetime     int64 `db:"refresh_token_lifetime" validate:"number,omitempty"`
	RefreshTokenIdleLifetime int64 `db:"refresh_token_idle_lifetime" validate:"number,omitempty"`
	MultiAuthEnabled         bool  `db:"multi_factor_auth_enabled" validate:"boolean,omitempty"`
	ForceMultiAuthEnabled    bool  `db:"force_multi_factor_auth" validate:"boolean,omitempty"`
	OTPEnabled               bool  `db:"otp_enabled" validate:"boolean,omitempty"`
	AuthenticatorEnabled     bool  `db:"authenticator_enabled" validate:"boolean,omitempty"`
	WebAuthnEnabled          bool  `db:"webauthn_enabled" validate:"boolean,omitempty"`
}

func (authSettings *AuthSettings) Load() error {
	authSettingQuery := db.Instance.Select("*").From("open_board_auth_settings")
	exists, err := authSettingQuery.ScanStruct(authSettings)

	if err != nil {
		return err
	}

	if !exists {
		createAuthSettingsQuery := db.Instance.From("open_board_auth_settings").Prepared(true).
			Insert().
			Rows(
				goqu.Record{},
			).Executor()

		if _, err := createAuthSettingsQuery.Exec(); err != nil {
			return err
		}

		authSettings.AccessTokenLifetime = int64(3600)
		authSettings.RefreshTokenLifetime = int64(7200)
		authSettings.RefreshTokenIdleLifetime = int64(1209600)
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
