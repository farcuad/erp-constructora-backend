package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // Driver de PostgreSQL
)

// NewPostgresDB inicializa y configura el pool de conexiones
func NewPostgresDB(host, port, user, password, dbname string) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Configuración del pool de conexiones (Go maneja esto de forma nativa y eficiente)
	db.SetMaxOpenConns(25)                 // Máximo de conexiones abiertas simultáneas
	db.SetMaxIdleConns(25)                 // Máximo de conexiones inactivas retenidas
	db.SetConnMaxLifetime(5 * time.Minute) // Tiempo máximo de vida de una conexión

	// Verificar si realmente hay comunicación con la BD
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
