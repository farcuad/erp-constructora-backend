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
	rolesWithPermissions := map[string][]string{
		"Administrador": {"*"}, // Comodín: acceso a todas las acciones
		"Gerente": {
			"projects:read", "projects:create", "projects:update",
			"budgets:read", "budgets:approve",
			"purchases:read", "purchases:approve",
			"inventory:read", "users:read",
		},
		"Ingeniero": {
			"projects:read", "projects:create", "projects:update",
			"budgets:read", "purchases:create", "purchases:read",
			"inventory:read",
		},
		"Compras": {
			"purchases:read", "purchases:create", "purchases:approve",
			"inventory:read", "projects:read",
		},
		"Contabilidad": {
			"budgets:read", "purchases:read", "projects:read",
		},
		"Almacén": {
			"inventory:read", "inventory:manage",
			"purchases:read", "projects:read",
		},
		"Supervisor": {
			"projects:read", "inventory:read", "budgets:read",
		},
	}

	// Enviar al repositorio para ejecutar la transacción
	err = s.repo.ExecRegistryTransaction(ctx, company, adminUser, rolesWithPermissions)
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
			ID:          user.ID,
			Name:        user.Name,
			Email:       user.Email,
			Role:        user.RoleName,
			Permissions: user.Permissions,
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
		UserID:      user.ID,
		CompanyID:   user.CompanyID,
		Permissions: user.Permissions,
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

func (s *Service) GetRoles(ctx context.Context, companyID string) ([]Role, error) {
	return s.repo.GetRolesByCompanyID(ctx, companyID)
}

func (s *Service) GetUsers(ctx context.Context, companyID string) ([]UserResponse, error) {
	return s.repo.GetUsersExcludingAdmin(ctx, companyID)
}

func (s *Service) CreateUser(ctx context.Context, companyID string, dto CreateUserDTO) (*UserResponse, error) {
	if dto.Name == "" || dto.Email == "" || dto.Password == "" || dto.RoleID == "" {
		return nil, errors.New("todos los campos son obligatorios")
	}

	exists, err := s.repo.EmailExists(ctx, dto.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("el correo electrónico ya se encuentra registrado")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return s.repo.CreateUserTransaction(ctx, companyID, dto, string(hashedPassword))
}

func (s *Service) UpdateUser(ctx context.Context, companyID string, dto UpdateUserDTO) error {
	if dto.ID == "" || dto.Name == "" || dto.Email == "" || dto.RoleID == "" {
		return errors.New("ID, nombre, correo y rol son obligatorios")
	}
	return s.repo.UpdateUserTransaction(ctx, companyID, dto)
}

func (s *Service) DeleteUser(ctx context.Context, companyID, userID string) error {
	if userID == "" {
		return errors.New("el ID del usuario es obligatorio")
	}
	return s.repo.DeleteUser(ctx, companyID, userID)
}
