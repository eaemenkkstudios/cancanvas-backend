package service

import (
	"net/smtp"
	"os"
)

// MailerService interface
type MailerService interface {
	SendMail(recipient, message string) error
}

type mailerService struct {
	from     string
	password string
}

func (s *mailerService) SendMail(recipient, message string) error {
	msg := "From: " + s.from + "\n" +
		"To: " + recipient + "\n" +
		"Subject: Password recovery for " + recipient + "\n\n" +
		message
	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", s.from, s.password, "smtp.gmail.com"),
		s.from,
		[]string{recipient},
		[]byte(msg),
	)
	return err
}

// NewMailerService function
func NewMailerService() MailerService {
	from := os.Getenv("MAILER_FROM")
	password := os.Getenv("MAILER_PASSWORD")
	return &mailerService{
		from,
		password,
	}
}
