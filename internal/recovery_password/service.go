package recoverypassword

import (
	"crypto/rand"
	"erp-constructora/internal/services"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo        *AuthRepository
	mailService *services.MailService
}

func NewAuthService(repo *AuthRepository, mailService *services.MailService) *AuthService {
	return &AuthService{
		repo:        repo,
		mailService: mailService,
	}
}

func generate6DigitToken() string {
	var table = []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	b := make([]byte, 6)
	n, err := io.ReadAtLeast(rand.Reader, b, 6)
	if n != 6 || err != nil {
		return "123456"
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

func (s *AuthService) RequestPasswordReset(email string) error {
	token := generate6DigitToken()
	expiresAt := time.Now().Add(15 * time.Minute)

	if err := s.repo.CreateRecoveryToken(email, token, expiresAt); err != nil {
		return err
	}

	go func() {
		htmlBody := fmt.Sprintf(`
			<div style="text-align:center;padding:10px 0;">
				<div style="display:inline-block;background-color:#F99B2E;color:#000000;padding:10px 25px;border-radius:6px;font-weight:800;font-size:14px;letter-spacing:1px;margin-bottom:25px;">RECUPERAR CONTRASEÑA</div>
				<h1 style="color:#000000;font-size:24px;font-weight:700;margin:0 0 15px 0;">Tu código de verificación</h1>
				<div style="display:inline-block;background-color:#F99B2E;color:#000000;padding:14px 40px;border-radius:8px;font-weight:800;font-size:32px;letter-spacing:8px;margin-bottom:20px;">%s</div>
				<p style="color:#666666;font-size:14px;line-height:1.6;">Este código vencerá en 15 minutos.</p>
			</div>`, token)

		if err := s.mailService.SendEmail(email, "Código de recuperación - Construx", htmlBody); err != nil {
			log.Printf("Error enviando email de recuperación a %s: %v", email, err)
		}
	}()

	return nil
}

func (s *AuthService) ResetPassword(email, token, newPassword string) error {
	valid, err := s.repo.ValidateToken(email, token)
	if err != nil || !valid {
		return errors.New("el código es inválido o ha expirado")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateUserPassword(email, string(hashedPassword)); err != nil {
		return err
	}

	return s.repo.MarkTokenAsUsed(email, token)
}