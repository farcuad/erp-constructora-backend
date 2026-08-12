package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // Driver de PostgreSQL
)

// NewPostgresDB inicializa y configura el pool de conexiones
func NewPostgresDB(host, port, user, password, dbname string) (*sql.DB, error) {
	// Agregamos statement_cache_mode=describe o binary_parameters=no dependiendo del driver,
	// pero para lib/pq con PgBouncer/Supabase la clave es connect_timeout=10
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=10 binary_parameters=no",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Configuración del pool de conexiones (Go maneja esto de forma nativa y eficiente)
	db.SetMaxOpenConns(25)                  // Máximo de conexiones abiertas simultáneas
	db.SetMaxIdleConns(5)                   // Máximo de conexiones inactivas retenidas
	db.SetConnMaxLifetime(15 * time.Minute) // Tiempo máximo de vida de una conexión
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Verificar si realmente hay comunicación con la BD
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
