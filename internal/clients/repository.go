package clients

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Repository interface {
	Create(ctx context.Context, companyID string, c *CreateClientRequest) (*Client, error)
	GetByCompany(ctx context.Context, companyID string) ([]Client, error)
	GetByID(ctx context.Context, companyID string, id string) (*Client, error)
	Update(ctx context.Context, companyID string, id string, req *UpdateClientRequest) error
	Delete(ctx context.Context, companyID string, id string) error
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

func (r *repository) GetByID(ctx context.Context, companyID string, id string) (*Client, error) {
	query := `SELECT id, company_id, name, nit, address, phone, email, is_active, created_at, updated_at
			  FROM clients WHERE company_id = $1 AND id = $2`

	var client Client
	err := r.db.QueryRowContext(ctx, query, companyID, id).Scan(
		&client.ID, &client.CompanyID, &client.Name, &client.NIT, &client.Address,
		&client.Phone, &client.Email, &client.IsActive, &client.CreatedAt, &client.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *repository) Update(ctx context.Context, companyID string, id string, req *UpdateClientRequest) error {
	var setClauses []string
	var args []interface{}
	argIdx := 1

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name=$%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.NIT != nil {
		setClauses = append(setClauses, fmt.Sprintf("nit=$%d", argIdx))
		args = append(args, *req.NIT)
		argIdx++
	}
	if req.Address != nil {
		setClauses = append(setClauses, fmt.Sprintf("address=$%d", argIdx))
		args = append(args, *req.Address)
		argIdx++
	}
	if req.Phone != nil {
		setClauses = append(setClauses, fmt.Sprintf("phone=$%d", argIdx))
		args = append(args, *req.Phone)
		argIdx++
	}
	if req.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email=$%d", argIdx))
		args = append(args, *req.Email)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at=NOW()"))
	args = append(args, companyID, id)

	query := fmt.Sprintf("UPDATE clients SET %s WHERE company_id=$%d AND id=$%d",
		strings.Join(setClauses, ", "), argIdx, argIdx+1)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *repository) Delete(ctx context.Context, companyID string, id string) error {
	query := `DELETE FROM clients WHERE company_id = $1 AND id = $2`
	_, err := r.db.ExecContext(ctx, query, companyID, id)
	return err
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
