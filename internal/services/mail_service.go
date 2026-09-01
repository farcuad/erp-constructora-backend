package services

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"
	"os"
)

type MailService struct {
	from string
}

func NewMailService() *MailService {
	return &MailService{
		from: fmt.Sprintf("Construx ERP <%s>", os.Getenv("GMAIL_USER")),
	}
}

func (s *MailService) SendEmail(toEmail, subject, htmlBody string) error {
	if toEmail == "" {
		return errors.New("el correo de destino es requerido")
	}

	gmailUser := os.Getenv("GMAIL_USER")
	gmailPass := os.Getenv("GMAIL_APP_PASSWORD")
	if gmailUser == "" || gmailPass == "" {
		return errors.New("GMAIL_USER y GMAIL_APP_PASSWORD no están configurados")
	}

	auth := smtp.PlainAuth("", gmailUser, gmailPass, "smtp.gmail.com")

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.from, toEmail, subject, htmlBody,
	)

	tlsConfig := &tls.Config{
		ServerName: "smtp.gmail.com",
	}

	conn, err := tls.Dial("tcp", "smtp.gmail.com:465", tlsConfig)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, "smtp.gmail.com")
	if err != nil {
		return err
	}

	if err = c.Auth(auth); err != nil {
		return err
	}

	if err = c.Mail(gmailUser); err != nil {
		return err
	}

	if err = c.Rcpt(toEmail); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}