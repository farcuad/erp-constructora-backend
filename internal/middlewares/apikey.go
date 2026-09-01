package middlewares

import (
	"crypto/subtle"
	"erp-constructora/internal/utils"
	"net/http"
	"os"
)

// APIKeyHeader es el encabezado HTTP donde se espera la API key.
// Ejemplo: X-API-Key: <valor de APP_API_KEY>
const APIKeyHeader = "X-API-Key"

// RequireAPIKey protege rutas internas (publicación/consulta de versiones de la app)
// validando que el encabezado X-API-Key coincida con la variable de entorno APP_API_KEY.
func RequireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := os.Getenv("APP_API_KEY")
		if expected == "" {
			utils.WriteInternalError(w, "API key no configurada en el servidor")
			return
		}

		provided := r.Header.Get(APIKeyHeader)
		if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			utils.WriteUnauthorized(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}
