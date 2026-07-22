package middlewares

import (
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims define qué datos irán encriptados de forma pública dentro del token
type JWTClaims struct {
	UserID       string `json:"user_id"`
	CompanyID    string `json:"company_id"`
	IsSuperAdmin bool   `json:"is_super_admin,omitempty"`
	jwt.RegisteredClaims
}
