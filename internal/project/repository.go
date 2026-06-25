package project

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

func (r *Repository) Save(ctx context.Context, p *Project) error {
	query := `INSERT INTO projects (company_id, name, client_name, location, start_date, end_date, budget, status_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, (SELECT id FROM project_statuses WHERE name = 'Planeación' LIMIT 1))
	RETURNING id, status_id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query, p.CompanyID, p.Name, p.ClientID,
		p.Location, p.StartDate, p.EndDate,
		p.Budget).Scan(&p.ID, &p.StatusID, &p.CreatedAt, &p.UpdatedAt)

	return err
}

func (r *Repository) GetAll(ctx context.Context, companyID string) ([]Project, error) {
	query := `
		SELECT id, company_id, name, client_id, location, start_date, end_date, budget, status_id, created_at, updated_at 
		FROM projects 
		WHERE company_id = $1 
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project

	for rows.Next() {
		var p Project
		err := rows.Scan(
			&p.ID, &p.CompanyID, &p.Name, &p.ClientID, &p.Location,
			&p.StartDate, &p.EndDate, &p.Budget, &p.StatusID, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}
