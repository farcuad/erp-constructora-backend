package users

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Empresa (Multi-tenant)
type Company struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	NIT       string    `json:"nit"`
	CreatedAt time.Time `json:"created_at"`
}

// Usuario
type User struct {
	ID           string    `json:"id"`
	CompanyID    string    `json:"company_id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // El '-' evita que se envíe la clave en el JSON
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// Estructura para el JSON que enviará React/Flutter al hacer POST /register
type RegisterDTO struct {
	CompanyName string `json:"company_name"`
	CompanyNIT  string `json:"company_nit"`
	AdminName   string `json:"admin_name"`
	AdminEmail  string `json:"admin_email"`
	Password    string `json:"password"`
}

type LoginDto struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// JWTClaims define qué datos irán encriptados de forma pública dentro del token
type JWTClaims struct {
	UserID    string `json:"user_id"`
	CompanyID string `json:"company_id"`
	// Aquí podrías agregar "Role" en el futuro si lo necesitas
	jwt.RegisteredClaims
}
