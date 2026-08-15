// internal/middlewares/logger.go
package middlewares

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

// statusRecorder captura el código HTTP que escribe el handler real,
// para que LoggerAndRecovery solo registre las respuestas de error.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.ResponseWriter.Write(b)
}

// LoggerAndRecovery registra solo fallos: panics y respuestas 4xx/5xx.
// No imprime nada cuando la petición termina correctamente.
func LoggerAndRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}

		// Captura de Panics (Evita que el proceso muera si algo revienta)
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC CRITICAL] %s %s | Error: %v\nStack trace:\n%s",
					r.Method, r.URL.Path, err, string(debug.Stack()))
				http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
				return
			}
			if rec.status >= 400 {
				log.Printf("[ERROR] %s %s - status %d - %v", r.Method, r.URL.Path, rec.status, time.Since(start))
			}
		}()

		// Ejecuta la petición
		next.ServeHTTP(rec, r)
	})
}
