package middlewares

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Definimos un tipo único para las llaves del contexto (evita colisiones de nombres)
type contextKey string

const (
	UserIDKey       contextKey = "userID"
	CompanyIDKey    contextKey = "companyKey"
	IsSuperAdminKey contextKey = "isSuperAdmin"
)

// AuthMiddleware protege las rutas verificando el token JWT
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string
		// 1. Obtener el encabezado Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		if tokenString == "" {
			tokenString = r.URL.Query().Get("token")
		}

		if tokenString == "" {
			http.Error(w, "Se requiere token de autenticación", http.StatusUnauthorized)
			return
		}

		// 3. Parsear y validar el token JWT
		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Validar el método de firma
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("método de firma inesperado")
			}
			secretKey := os.Getenv("JWT_SECRET")
			if secretKey == "" {
				secretKey = "mi_clave_secreta_super_segura_para_la_constructora"
			}
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Token inválido o expirado", http.StatusUnauthorized)
			return
		}

		// 4. Inyectar los datos del token en el Contexto de la petición HTTP
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, CompanyIDKey, claims.CompanyID)
		ctx = context.WithValue(ctx, IsSuperAdminKey, claims.IsSuperAdmin)
		ctx = context.WithValue(ctx, PermissionsKey, claims.Permissions)

		// 5. Dejar pasar la petición al siguiente Handler con el nuevo contexto enriquecido
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helpers útiles para que cualquier módulo (Proyectos, Clientes) extraiga los datos del contexto de forma limpia
func GetCompanyIDFromContext(ctx context.Context) (string, bool) {
	companyID, ok := ctx.Value(CompanyIDKey).(string)
	return companyID, ok
}

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

func IsSuperAdminFromContext(ctx context.Context) bool {
	isSuperAdmin, ok := ctx.Value(IsSuperAdminKey).(bool)
	return ok && isSuperAdmin
}

func RequireSuperAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsSuperAdminFromContext(r.Context()) {
			http.Error(w, "Acceso denegado: solo el administrador del sistema", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
