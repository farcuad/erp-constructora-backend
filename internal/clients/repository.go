package clients

import (
	"context"
	"database/sql"
)

type Repository interface {
	Create(ctx context.Context, companyID string, c *CreateClientRequest) (*Client, error)
	GetByCompany(ctx context.Context, companyID string) ([]Client, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, companyID string, c *CreateClientRequest) (*Client, error) {
	query := `
		INSERT INTO clients (company_id, name, nit, address, phone, email)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, company_id, name, nit, address, phone, email, is_active, created_at, updated_at`

	var client Client
	err := r.db.QueryRowContext(ctx, query, companyID, c.Name, c.NIT, c.Address, c.Phone, c.Email).Scan(
		&client.ID, &client.CompanyID, &client.Name, &client.NIT, &client.Address,
		&client.Phone, &client.Email, &client.IsActive, &client.CreatedAt, &client.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *repository) GetByCompany(ctx context.Context, companyID string) ([]Client, error) {
	query := `SELECT id, company_id, name, nit, address, phone, email, is_active, created_at, updated_at 
			  FROM clients WHERE company_id = $1`

	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []Client
	for rows.Next() {
		var client Client
		err := rows.Scan(
			&client.ID, &client.CompanyID, &client.Name, &client.NIT, &client.Address,
			&client.Phone, &client.Email, &client.IsActive, &client.CreatedAt, &client.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	return clients, nil
}
