package settings

import (
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/doug-martin/goqu/v9"
)

type AuthSettings struct {
	AccessTokenLifetime      int64 `db:"access_token_lifetime"`
	RefreshTokenLifetime     int64 `db:"refresh_token_lifetime"`
	RefreshTokenIdleLifetime int64 `db:"refresh_token_idle_lifetime"`
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
		createAuthSettingsQuery := db.Instance.Insert("open_board_auth_settings").Rows(
			goqu.Record{
				"access_token_lifetime":       accessTokenLifetime,
				"refresh_token_lifetime":      refreshTokenLifetime,
				"refresh_token_idle_lifetime": refreshTokenIdleLifetime,
			},
		).Executor()

		if _, err := createAuthSettingsQuery.Exec(); err != nil {
			return err
		}

		authSettings.AccessTokenLifetime = int64(accessTokenLifetime)
		authSettings.RefreshTokenLifetime = int64(refreshTokenLifetime)
		authSettings.RefreshTokenIdleLifetime = int64(refreshTokenIdleLifetime)

	}

	return nil
}

func (authSettings *AuthSettings) Save() error {
	authSettingsUpdateQuery := db.Instance.Update("open_board_auth_settings").Set(
		goqu.Record{
			"access_token_lifetime":       authSettings.AccessTokenLifetime,
			"refresh_token_lifetime":      authSettings.RefreshTokenLifetime,
			"refresh_token_idle_lifetime": authSettings.RefreshTokenIdleLifetime,
		},
	).Executor()

	if _, err := authSettingsUpdateQuery.Exec(); err != nil {
		return err
	}

	return nil
}
