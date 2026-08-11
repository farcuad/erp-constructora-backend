// internal/middlewares/logger.go
package middlewares

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

func LoggerAndRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Captura de Panics (Evita que el proceso muera si algo revienta)
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC CRITICAL] %s %s | Error: %v\nStack trace:\n%s",
					r.Method, r.URL.Path, err, string(debug.Stack()))
				http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
			}
		}()

		log.Printf("[REQ START] %s %s", r.Method, r.URL.Path)

		// Ejecuta la petición
		next.ServeHTTP(w, r)

		log.Printf("[REQ END] %s %s | Time: %v", r.Method, r.URL.Path, time.Since(start))
	})
}
