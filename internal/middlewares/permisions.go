package middlewares

import (
	"encoding/json"
	"net/http"
)

const PermissionsKey contextKey = "permissions"

// RequirePermission evalúa si el JWT posee el permiso requerido
func RequirePermission(requiredPermission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			// 1. Extraer el slice de permisos del contexto
			userPerms, ok := r.Context().Value(PermissionsKey).([]string)
			if !ok {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{"message": "Acceso denegado: permisos no encontrados"})
				return
			}

			// 2. Verificar si el usuario tiene el permiso específico o acceso total "*"
			hasPermission := false
			for _, perm := range userPerms {
				if perm == requiredPermission || perm == "*" {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				w.WriteHeader(http.StatusForbidden) // 403 Forbidden
				json.NewEncoder(w).Encode(map[string]string{
					"message": "Acceso denegado: no tienes permisos para realizar esta acción",
				})
				return
			}

			// 3. Permiso concedido, continuar al Handler
			next.ServeHTTP(w, r)
		})
	}
}
