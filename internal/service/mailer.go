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
	// Most callers discard the error, so surface it here (no recipient: PII).
	if err := d.DialAndSend(msg); err != nil {
		slog.Error("smtp send failed", "host", m.Host, "subject", subject, "err", err)
		return err
	}
	return nil
}

// StdoutMailer logs to slog; used in dev/test.
type StdoutMailer struct{}

// Send implements Mailer.
func (StdoutMailer) Send(_ context.Context, to, subject, body string) error {
	slog.Info("dev mail", "to", to, "subject", subject, "body", body)
	return nil
}
