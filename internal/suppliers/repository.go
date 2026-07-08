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

func (r *Repository) GetSupplierByID(ctx context.Context, id, companyID string) (*Supplier, error) {
	query := `SELECT id, company_id, name, nit, address, phone, email, is_active, created_at, updated_at 
	          FROM suppliers WHERE id = $1 AND company_id = $2`
	var s Supplier
	err := r.db.QueryRowContext(ctx, query, id, companyID).Scan(&s.ID, &s.CompanyID, &s.Name, &s.NIT, &s.Address, &s.Phone, &s.Email, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) UpdateSupplier(ctx context.Context, s *Supplier) error {
	query := `UPDATE suppliers SET name = $1, nit = $2, address = $3, phone = $4, email = $5, updated_at = CURRENT_TIMESTAMP WHERE id = $6 AND company_id = $7`
	_, err := r.db.ExecContext(ctx, query, s.Name, s.NIT, s.Address, s.Phone, s.Email, s.ID, s.CompanyID)
	return err
}

func (r *Repository) DeleteSupplier(ctx context.Context, id, companyID string) error {
	query := `DELETE FROM suppliers WHERE id = $1 AND company_id = $2`
	_, err := r.db.ExecContext(ctx, query, id, companyID)
	return err
}
