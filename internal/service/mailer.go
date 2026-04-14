package service

import (
	"context"
	"log/slog"

	"gopkg.in/gomail.v2"
)

// Mailer abstracts outbound transactional email.
type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

// SMTPMailer sends through a real SMTP relay.
type SMTPMailer struct {
	Host, User, Password, From string
	Port                       int
}

// Send implements Mailer.
func (m *SMTPMailer) Send(_ context.Context, to, subject, body string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.From)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain; charset=utf-8", body)
	d := gomail.NewDialer(m.Host, m.Port, m.User, m.Password)
	return d.DialAndSend(msg)
}

// StdoutMailer logs to slog; used in dev/test.
type StdoutMailer struct{}

// Send implements Mailer.
func (StdoutMailer) Send(_ context.Context, to, subject, body string) error {
	slog.Info("dev mail", "to", to, "subject", subject, "body", body)
	return nil
}
