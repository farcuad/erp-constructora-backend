package superadmin

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*SuperAdmin, error) {
	query := `SELECT id, name, email, password_hash, created_at FROM super_admins WHERE email = $1`
	var sa SuperAdmin
	err := r.db.QueryRowContext(ctx, query, email).Scan(&sa.ID, &sa.Name, &sa.Email, &sa.PasswordHash, &sa.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &sa, nil
}
