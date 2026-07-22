package superadmin

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"erp-constructora/internal/middlewares"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Login(ctx context.Context, dto LoginDTO) (string, error) {
	if dto.Email == "" || dto.Password == "" {
		return "", errors.New("correo y contraseña son obligatorios")
	}

	admin, err := s.repo.GetByEmail(ctx, dto.Email)
	if err != nil {
		return "", errors.New("credenciales inválidas")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(dto.Password)); err != nil {
		return "", errors.New("credenciales inválidas")
	}

	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "mi_clave_secreta_super_segura_para_la_constructora"
	}

	claims := middlewares.JWTClaims{
		UserID:       admin.ID,
		IsSuperAdmin: true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}
