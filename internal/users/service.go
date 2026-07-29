package users

import (
	"context"
	"errors"
	"log"
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

func (s *Service) RegisterCompanyAndAdmin(ctx context.Context, dto RegisterDTO) (*Company, *User, error) {
	// Regla de negocio 1: Verificar si el correo ya está en uso
	exists, err := s.repo.EmailExists(ctx, dto.AdminEmail)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, errors.New("el correo electrónico ya se encuentra registrado")
	}

	// Regla de negocio 2: Encriptar la contraseña del administrador
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	// Preparar las entidades
	company := &Company{
		Name: dto.CompanyName,
		NIT:  dto.CompanyNIT,
	}

	adminUser := &User{
		Name:         dto.AdminName,
		Email:        dto.AdminEmail,
		PasswordHash: string(hashedPassword),
		IsActive:     true,
	}

	// Lista de roles iniciales requeridos para el flujo de la constructora según el diseño original
	defaultRoles := []string{"Administrador", "Gerente", "Ingeniero", "Supervisor", "Compras", "Contabilidad", "Almacén"} //

	// Enviar al repositorio para ejecutar la transacción
	err = s.repo.ExecRegistryTransaction(ctx, company, adminUser, defaultRoles)
	if err != nil {
		return nil, nil, err
	}

	return company, adminUser, nil
}

func (s *Service) Login(ctx context.Context, dto LoginDto) (*LoginResponse, error) {
	user, err := s.repo.GetEmailUserWithDetails(ctx, dto.Email)
	if err != nil {
		log.Println("[DEBUG LOGIN ERROR]:", err)
		return nil, errors.New("Credenciales Incorrectas")
	}

	if !user.IsActive {
		return nil, errors.New("El usuario se encuentra inactivo")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(dto.Password)) != nil {
		log.Println("[DEBUG BCRYPT ERROR]:", err) // Muestra si la clave ingresada no coincide con el hash
		return nil, errors.New("Credenciales Incorrectas")
	}

	tokenString, err := s.generateJwt(user)
	if err != nil {
		return nil, errors.New("Error al generar el token de acceso")
	}

	return &LoginResponse{
		Token: tokenString,
		User: UserResponse{
			ID:         user.ID,
			Name:       user.Name,
			Email:      user.Email,
			Role:       user.RoleName,
			EmployeeID: user.EmployeeID,
		},
	}, nil
}

func (s *Service) generateJwt(user *User) (string, error) {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		// Llave por defecto segura si olvidas configurarla en desarrollo
		secretKey = "mi_clave_secreta_super_segura_para_la_constructora"
	}

	// Crear los Claims (Carga útil del token)
	claims := middlewares.JWTClaims{
		UserID:    user.ID,
		CompanyID: user.CompanyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Expira en 24 horas
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Crear el token usando el algoritmo de firma HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Firmar el token con nuestra clave secreta
	return token.SignedString([]byte(secretKey))
}
