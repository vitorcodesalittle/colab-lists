package infra

import (
	"crypto/tls"

	gomail "gopkg.in/mail.v2"
	"vilmasoftware.com/colablists/pkg/config"
)

func SendEmail(to []string, subject string, body string) error {
	smtpConfig := config.GetConfig().SmtpConfig
	m := gomail.NewMessage()
	m.SetHeader("From", smtpConfig.FromNoReply)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)
	d := gomail.NewDialer(smtpConfig.Host, smtpConfig.Port, smtpConfig.Username, smtpConfig.Password)

	// TODO: use tls when not in dev mode
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
