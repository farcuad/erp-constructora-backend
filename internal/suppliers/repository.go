package suppliers

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

// --- MÉTODOS DE PROVEEDORES ---

func (r *Repository) CreateSupplier(ctx context.Context, s *Supplier) error {
	query := `INSERT INTO suppliers (company_id, name, nit, address, phone, email) 
	          VALUES ($1, $2, $3, $4, $5, $6) 
	          RETURNING id, is_active, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, s.CompanyID, s.Name, s.NIT, s.Address, s.Phone, s.Email).
		Scan(&s.ID, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
}

func (r *Repository) GetSuppliersByCompany(ctx context.Context, companyID string) ([]Supplier, error) {
	query := `SELECT id, company_id, name, nit, address, phone, email, is_active, created_at, updated_at 
	          FROM suppliers WHERE company_id = $1`
	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suppliers []Supplier
	for rows.Next() {
		var s Supplier
		if err := rows.Scan(&s.ID, &s.CompanyID, &s.Name, &s.NIT, &s.Address, &s.Phone, &s.Email, &s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		suppliers = append(suppliers, s)
	}
	return suppliers, nil
}
