package settings

import (
	"database/sql"
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/doug-martin/goqu/v9"
)

type NotificationSettings struct {
	SMTPServer   sql.NullString `db:"smtp_server,omitempty" validate:"url,omitempty"`
	SMTPPort     sql.NullInt64  `db:"smtp_port,omitempty" validate:"hostname_port,omitempty"`
	SMTPUser     sql.NullString `db:"smtp_user,omitempty" validate:"username,omitempty"`
	SMTPPassword sql.NullString `db:"smtp_password,omitempty" validate:"password,omitempty"`
	Name         sql.NullString `db:"name,omitempty" validate:"alphanum,omitempty"`
	EmailAddress sql.NullString `db:"email_address, omitempty" validate:"email,omitempty"`
	UseTLS       bool           `db:"use_tls" validate:"boolean, omitempty"`
}

func (notificationSettings *NotificationSettings) Load() error {
	notificationSettingsQuery := db.Instance.Select("*").From("open_board_notification_settings")
	exists, err := notificationSettingsQuery.ScanStruct(notificationSettings)

	if err != nil {
		return err
	}

	if !exists {
		createNotificationSettingsQuery := db.Instance.From("open_board_notification_settings").Prepared(true).
			Insert().
			Rows(goqu.Record{}).
			Executor()

		if _, err := createNotificationSettingsQuery.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (notificationSettings *NotificationSettings) Save() error {
	updateNotificationSettingsQuery := db.Instance.From("open_board_notification_settings").Prepared(true).
		Update().
		Set(goqu.Record{
			"smtp_server":   notificationSettings.SMTPServer,
			"smtp_port":     notificationSettings.SMTPPort,
			"smtp_user":     notificationSettings.SMTPUser,
			"smtp_password": notificationSettings.SMTPPassword,
			"name":          notificationSettings.Name,
			"email_address": notificationSettings.EmailAddress,
			"use_tls":       notificationSettings.UseTLS,
		}).
		Executor()

	if _, err := updateNotificationSettingsQuery.Exec(); err != nil {
		return err
	}

	return nil
}
