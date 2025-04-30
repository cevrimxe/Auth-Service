package models

import (
	"fmt"
	"net/smtp"

	"github.com/cevrimxe/auth-service/config"
)

type Mailer struct {
	From     string
	Password string
	Host     string
	Port     string
}

func NewMailer() *Mailer {
	config.LoadEnv()

	return &Mailer{
		From:     config.GetEnv("SMTP_SENDER_EMAIL"),
		Password: config.GetEnv("SMTP_SENDER_PASSWORD"),
		Host:     config.GetEnv("SMTP_HOST"),
		Port:     config.GetEnv("SMTP_PORT"),
	}
}

func (m *Mailer) SendVerifyEmail(toEmail, token string) error {
	auth := smtp.PlainAuth("", m.From, m.Password, m.Host)
	subject := "Subject: Email Verification\n"
	body := fmt.Sprintf("Click to verify your email: http://localhost:8080/verify?token=%s", token)
	msg := []byte(subject + "\n" + body)

	return smtp.SendMail(m.Host+":"+m.Port, auth, m.From, []string{toEmail}, msg)
}
