package settings

import (
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/doug-martin/goqu/v9"
)

type AuthSettings struct {
	AccessTokenLifetime      int64
	RefreshTokenLifetime     int64
	RefreshTokenIdleLifetime int64
}

func (authSettings AuthSettings) Load() error {
	authSettingQuery := db.Instance.Dialect.Select("access_token_lifetime", "refresh_token_lifetime", "refresh_token_idle_lifetime").From("open_board_auth_settings")
	authSettingsQueryResults, err := db.Instance.ExecQuery(authSettingQuery)

	if err != nil {
		return err
	}

	var authSettingData db.ExtractedRow

	if authSettingsQueryResults.Size == 0 {
		createAuthSettingsQuery := db.Instance.Dialect.Insert("open_board_auth_settings").Returning(
			"access_token_lifetime",
			"refresh_token_lifetime",
			"refresh_token_idle_lifetime",
		).Rows(
			goqu.Record{
				"access_token_lifetime":       3600,
				"refresh_token_lifetime":      7200,
				"refresh_token_idle_lifetime": 1209600,
			},
		)
		newAuthSettingResults, err := db.Instance.ExecSingleQuery(createAuthSettingsQuery, []string{
			"access_token_lifetime",
			"refresh_token_lifetime",
			"refresh_token_idle_lifetime",
		})

		if err != nil {
			return err
		}

		authSettingData = newAuthSettingResults.Row

	} else {
		authSettingData = authSettingsQueryResults.Rows[0]
	}

	authSettings.AccessTokenLifetime = authSettingData["access_token_lifetime"].(int64)
	authSettings.RefreshTokenLifetime = authSettingData["refresh_token_lifetime"].(int64)
	authSettings.RefreshTokenIdleLifetime = authSettingData["refresh_token_lifetime"].(int64)
	return nil
}

func (authSettings AuthSettings) Save() error {
	authSettingsUpdateQuery := db.Instance.Dialect.Update("open_board_auth_settings").Set(
		goqu.Record{
			"access_token_lifetime":       authSettings.AccessTokenLifetime,
			"refresh_token_lifetime":      authSettings.RefreshTokenLifetime,
			"refresh_token_idle_lifetime": authSettings.RefreshTokenIdleLifetime,
		},
	)

	if _, err := db.Instance.ExecQuery(authSettingsUpdateQuery); err != nil {
		return err
	}

	return nil
}
