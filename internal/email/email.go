package email

import (
	"crypto/tls"
	"github.com/VEDA95/OpenBoard-API/internal/settings"
	"github.com/wneessen/go-mail"
	"log"
)

type MailClient struct {
	client *mail.Client
}

var Client *MailClient

func NewMailClient() (*MailClient, error) {
	settingsInterface := *settings.Instance.GetSettings("notification")
	notificationSettings := settingsInterface.(*settings.NotificationSettings)

	if !notificationSettings.SMTPServer.Valid && !notificationSettings.SMTPPort.Valid && !notificationSettings.SMTPUser.Valid && !notificationSettings.SMTPPassword.Valid {
		return nil, nil
	}

	SMTPUser, err := notificationSettings.SMTPUser.Value()

	if err != nil {
		return nil, err
	}

	SMTPPassword, err := notificationSettings.SMTPPassword.Value()

	if err != nil {
		return nil, err
	}

	log.Println("SMTP User: ", SMTPUser)
	log.Println("SMTP Password: ", SMTPPassword)

	mailOptions := []mail.Option{
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(SMTPUser.(string)),
		mail.WithPassword(SMTPPassword.(string)),
	}

	smtpPort, err := notificationSettings.SMTPPort.Value()

	if err != nil {
		return nil, err
	}

	SMTPPort := smtpPort.(int)

	log.Println("SMTP Port: ", SMTPPort)

	if SMTPPort > 0 {
		mailOptions = append(mailOptions, mail.WithPort(SMTPPort))
	}

	smtpServer, err := notificationSettings.SMTPServer.Value()

	if err != nil {
		return nil, err
	}

	SMTPServer := smtpServer.(string)

	if notificationSettings.UseTLS {
		mailOptions = append(mailOptions, mail.WithTLSConfig(&tls.Config{
			InsecureSkipVerify: true,
			ClientAuth:         tls.NoClientCert,
			ServerName:         SMTPServer,
		}))
	}

	client, err := mail.NewClient(SMTPServer, mailOptions...)

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
