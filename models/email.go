package models

import (
	"fmt"
	"net/smtp"
	"strings"

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

func (m *Mailer) Mailer(toEmail, subject, body string) error {
	headers := make(map[string]string)
	headers["From"] = m.From
	headers["To"] = toEmail
	headers["Subject"] = subject
	headers["Content-Type"] = "text/plain; charset=UTF-8"

	var msgBuilder strings.Builder
	for key, value := range headers {
		msgBuilder.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	// body
	msgBuilder.WriteString("\r\n" + body)

	msg := []byte(msgBuilder.String())

	// SMTP sunucusu ile bağlantıyı kur
	auth := smtp.PlainAuth("", m.From, m.Password, m.Host)
	return smtp.SendMail(m.Host+":"+m.Port, auth, m.From, []string{toEmail}, msg)
}
