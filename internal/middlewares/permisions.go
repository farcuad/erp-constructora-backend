package middlewares

import (
	"encoding/json"
	"net/http"
)

// RequireRole evalúa si el rol del usuario (desde el JWT) está permitido para la ruta.
// El rol "Administrador" siempre tiene acceso total y no necesita la lista.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			// 1. Extraer el rol del contexto (inyectado por AuthMiddleware)
			userRole, ok := GetUserRoleFromContext(r.Context())
			if !ok || userRole == "" {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{"message": "Acceso denegado: rol no encontrado"})
				return
			}

			// 2. El administrador tiene acceso total a todas las rutas
			if userRole == AdminRole {
				next.ServeHTTP(w, r)
				return
			}

			// 3. Verificar si el rol del usuario está dentro de los roles permitidos
			allowed := false
			for _, role := range allowedRoles {
				if role == userRole {
					allowed = true
					break
				}
			}

			if !allowed {
				w.WriteHeader(http.StatusForbidden) // 403 Forbidden
				json.NewEncoder(w).Encode(map[string]string{
					"message": "Acceso denegado: tu rol no tiene permisos para realizar esta acción",
				})
				return
			}

			// 4. Rol autorizado, continuar al Handler
			next.ServeHTTP(w, r)
		})
	}
}