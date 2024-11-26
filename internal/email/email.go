package email

import (
	"crypto/tls"
	"github.com/VEDA95/OpenBoard-API/internal/settings"
	"github.com/wneessen/go-mail"
)

type MailClient struct {
	client *mail.Client
}

var Client *MailClient

func NewMailClient() (*MailClient, error) {
	settingsInterface := *settings.Instance.GetSettings("notification")
	notificationSettings := settingsInterface.(*settings.NotificationSettings)
	mailOptions := []mail.Option{
		mail.WithPort(int(notificationSettings.SMTPPort)),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(notificationSettings.SMTPUser),
		mail.WithPassword(notificationSettings.SMTPPassword),
	}

	if notificationSettings.UseStarTTLS {
		mailOptions = append(mailOptions, mail.WithTLSConfig(&tls.Config{
			InsecureSkipVerify: true,
			ClientAuth:         tls.NoClientCert,
			ServerName:         notificationSettings.SMTPServer,
		}))
	}

	client, err := mail.NewClient(notificationSettings.SMTPServer, mailOptions...)

	if err != nil {
		return nil, err
	}

	return &MailClient{client}, nil
}

func InitializeMailClient() error {
	client, err := NewMailClient()

	if err != nil {
		return err
	}

	Client = client

	return nil
}

func (client *MailClient) Send(from string, to []string, subject string, body string, isHtml bool) error {
	message := mail.NewMsg()

	if err := message.From(from); err != nil {
		return err
	}

	if err := message.To(to...); err != nil {
		return err
	}

	if isHtml {
		message.SetBodyString(mail.TypeTextHTML, body)

	} else {
		message.SetBodyString(mail.TypeTextPlain, body)
	}

	message.Subject(subject)

	if err := client.client.DialAndSend(message); err != nil {
		return err
	}

	return nil
}
