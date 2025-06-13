package mailer

import (
	"bytes"
	"fmt"
	"net/smtp"
)

type SMTPConfig struct {
	Username string
	Password string
	Host     string
	HostAddr string
	From     string
}

type GmailMailer struct {
	Config    SMTPConfig
	PlainAuth smtp.Auth
}

func NewGmailMailer(cfg SMTPConfig) *GmailMailer {

	return &GmailMailer{
		Config: cfg,
		PlainAuth: smtp.PlainAuth(
			"",
			cfg.Username,
			cfg.Password,
			cfg.Host,
		),
	}
}

func (g *GmailMailer) Send(to string, sub string, body *bytes.Buffer) error {
	if body.Len() <= 0 {
		return fmt.Errorf("the body of the email cannot be empty")
	}

	headers := fmt.Sprintf(
		"From: %s\nTo: %s\nSubject: %s\nMIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";",
		g.Config.From, to, sub,
	)

	var msg bytes.Buffer
	msg.WriteString(headers + "\n\n")
	msg.Write(body.Bytes())

	return smtp.SendMail(
		g.Config.HostAddr,
		g.PlainAuth,
		g.Config.From,
		[]string{to},
		msg.Bytes(),
	)
}
