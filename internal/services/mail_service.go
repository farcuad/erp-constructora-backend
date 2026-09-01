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

	htmlBody = s.buildTemplate(htmlBody)

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

func (s *MailService) buildTemplate(content string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Construx ERP</title>
</head>
<body style="margin:0;padding:0;background-color:#F5F5F5;font-family:'Segoe UI',Arial,sans-serif;">
<table align="center" cellpadding="0" cellspacing="0" width="100%%" style="max-width:600px;">
<tr>
<td style="background-color:#000000;text-align:center;padding:30px 20px;">
<table align="center" cellpadding="0" cellspacing="0" width="100%%" style="max-width:560px;">
<tr>
<td style="text-align:center;">
<h1 style="color:#F99B2E;font-size:28px;font-weight:800;margin:0;letter-spacing:2px;">CONSTUEX</h1>
<p style="color:#999999;font-size:13px;margin:6px 0 0 0;letter-spacing:3px;">ERP &nbsp;|&nbsp; CONSTRUCTORA</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td style="background-color:#FFFFFF;padding:40px 30px;">
%s
</td>
</tr>
<tr>
<td style="background-color:#000000;padding:20px;text-align:center;">
<p style="color:#555555;font-size:11px;margin:0;letter-spacing:1px;">Construx ERP &mdash; Todos los derechos reservados</p>
</td>
</tr>
</table>
</body>
</html>`, content)
}