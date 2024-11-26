package settings

import (
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/doug-martin/goqu/v9"
)

type NotificationSettings struct {
	SMTPServer   string `db:"smtp_server"`
	SMTPPort     int64  `db:"smtp_port"`
	SMTPUser     string `db:"smtp_user"`
	SMTPPassword string `db:"smtp_password"`
	Name         string `db:"name"`
	EmailAddress string `db:"email_address"`
	UseTLS       bool   `db:"use_tls"`
	UseStarTTLS  bool   `db:"use_starttls"`
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
			"use_starttls":  notificationSettings.UseStarTTLS,
		}).
		Executor()

	if _, err := updateNotificationSettingsQuery.Exec(); err != nil {
		return err
	}

	return nil
}
