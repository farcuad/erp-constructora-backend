package app

import (
	"context"
	"database/sql"
	"errors"
	"log"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create inserta una nueva versión publicada de la app.
func (r *Repository) Create(ctx context.Context, req *CreateReleaseRequest) (*Release, error) {
	query := `
		INSERT INTO app_releases (version, app_url, description, file_size, checksum, is_mandatory)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, version, app_url, description, file_size, checksum, is_mandatory, created_at`

	var rel Release
	err := r.db.QueryRowContext(ctx, query,
		req.Version, req.AppURL, req.Description, req.FileSize, req.Checksum, req.IsMandatory,
	).Scan(&rel.ID, &rel.Version, &rel.AppURL, &rel.Description, &rel.FileSize, &rel.Checksum, &rel.IsMandatory, &rel.CreatedAt)
	if err != nil {
		log.Printf("[DB QUERY ERROR] app.Create (version=%s): %v", req.Version, err)
		return nil, err
	}
	return &rel, nil
}

// GetLatest devuelve la subida más reciente de la app (por fecha de creación).
// Retorna (nil, nil) si aún no hay ninguna versión registrada.
func (r *Repository) GetLatest(ctx context.Context) (*Release, error) {
	query := `
		SELECT id, version, app_url, description, file_size, checksum, is_mandatory, created_at
		FROM app_releases
		ORDER BY created_at DESC
		LIMIT 1`

	var rel Release
	err := r.db.QueryRowContext(ctx, query).Scan(
		&rel.ID, &rel.Version, &rel.AppURL, &rel.Description, &rel.FileSize,
		&rel.Checksum, &rel.IsMandatory, &rel.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		log.Printf("[DB QUERY ERROR] app.GetLatest: %v", err)
		return nil, err
	}
	return &rel, nil
}
