package main

import (
	"bufio"
	"errors"
	"log"
	"net"
	"net/http"
	"os"

	"erp-constructora/internal/database"

	"time"

	"github.com/joho/godotenv"
)

func enableCORS(next http.Handler) http.Handler {
	allowedOrigins := map[string]bool{
		"http://localhost:5173":                             true,
		"https://erp-constructora-frontend-6pfx.vercel.app": true,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Si el origen de la petición está permitido, lo seteamos dinámicamente
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PATCH, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-API-Key")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Manejar Preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// statusRecorder captura el código HTTP que escribe el handler real,
// para que loggerAndRecovery solo registre las respuestas de error.
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

// Hijack permite que gorilla/websocket pueda hacer el upgrade de la conexión
// (101 Switching Protocols) para /notifications/ws. Sin esto, el WebSocket
// falla porque el ResponseWriter envuelto no implementa http.Hijacker.
func (s *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := s.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("websocket: la respuesta no implementa http.Hijacker")
	}
	return hj.Hijack()
}

func loggerAndRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}

		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC CRITICAL] %s %s | Error: %v", r.Method, r.URL.Path, err)
				http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
				return
			}
			// Solo se registra cuando la llamada falló (4xx/5xx), nunca en éxito.
			if rec.status >= 400 {
				log.Printf("[ERROR] %s %s - status %d - %v", r.Method, r.URL.Path, rec.status, time.Since(start))
			}
		}()

		next.ServeHTTP(rec, r)
	})
}

func main() {
	// 1. Cargar el archivo .env que está en la raíz del proyecto
	// Como el ejecutable se corre desde la raíz, buscará el archivo .env ahí de forma nativa
	err := godotenv.Load("../..") // Si ejecutas desde cmd/api/, sube dos niveles.
	// Es mejor cargarlo buscando en la raíz:
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: No se encontró el archivo .env, se usarán variables de entorno globales")
	}

	// 2. Leer las variables de entorno usando el paquete "os" nativo de Go
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_DATABASE")

	// Validar que las variables críticas existan
	if dbHost == "" || dbName == "" {
		log.Fatal("Error: Las variables de entorno de la base de datos no están completas en el .env")
	}

	// 3. Inicializar la conexión a PostgreSQL pasando las variables leídas
	db, err := database.NewPostgresDB(dbHost, dbPort, dbUser, dbPass, dbName)
	if err != nil {
		log.Fatalf("No se pudo conectar a la base de datos: %v", err)
	}
	defer db.Close() // Se cerrará cuando apagues el servidor

	log.Println("Conexión exitosa a PostgreSQL desde el archivo .env")

	router := SetupRoutes(db)
	// 6. Encender el servidor HTTP
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Puerto por defecto si no está en el .env
	}

	log.Printf("Servidor corriendo en el puerto :%s...", port)
	finalHandler := enableCORS(loggerAndRecovery(router))

	if err := http.ListenAndServe(":"+port, finalHandler); err != nil {
		log.Fatal(err)
	}
}
