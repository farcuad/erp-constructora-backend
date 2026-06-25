package main

import (
	"log"
	"net/http"
	"os"

	"erp-constructora/internal/database"

	"github.com/joho/godotenv"
)

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
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
