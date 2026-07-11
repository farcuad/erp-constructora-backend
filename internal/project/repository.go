package project

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

func (r *Repository) GetByID(ctx context.Context, companyID string, id string) (*Project, error) {
	query := `
		SELECT id, company_id, name, client_id, location, start_date, end_date, budget, status_id, created_at, updated_at
		FROM projects
		WHERE company_id = $1 AND id = $2`

	var p Project
	err := r.db.QueryRowContext(ctx, query, companyID, id).Scan(
		&p.ID, &p.CompanyID, &p.Name, &p.ClientID, &p.Location,
		&p.StartDate, &p.EndDate, &p.Budget, &p.StatusID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) Update(ctx context.Context, companyID string, id string, dto UpdateProjectDTO) error {
	var setClauses []string
	var args []interface{}
	argIdx := 1

	if dto.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name=$%d", argIdx))
		args = append(args, *dto.Name)
		argIdx++
	}
	if dto.ClientID != nil {
		setClauses = append(setClauses, fmt.Sprintf("client_id=$%d", argIdx))
		args = append(args, *dto.ClientID)
		argIdx++
	}
	if dto.Location != nil {
		setClauses = append(setClauses, fmt.Sprintf("location=$%d", argIdx))
		args = append(args, *dto.Location)
		argIdx++
	}
	if dto.StartDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("start_date=$%d", argIdx))
		args = append(args, *dto.StartDate)
		argIdx++
	}
	if dto.EndDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("end_date=$%d", argIdx))
		args = append(args, *dto.EndDate)
		argIdx++
	}
	if dto.Budget != nil {
		setClauses = append(setClauses, fmt.Sprintf("budget=$%d", argIdx))
		args = append(args, *dto.Budget)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at=NOW()"))
	args = append(args, companyID, id)

	query := fmt.Sprintf("UPDATE projects SET %s WHERE company_id=$%d AND id=$%d",
		strings.Join(setClauses, ", "), argIdx, argIdx+1)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) Delete(ctx context.Context, companyID string, id string) error {
	query := `DELETE FROM projects WHERE company_id = $1 AND id = $2`
	_, err := r.db.ExecContext(ctx, query, companyID, id)
	return err
}

func (r *Repository) GetAll(ctx context.Context, companyID string) ([]Project, error) {
	query := `
		SELECT id, company_id, name, client_name, location, start_date, end_date, budget, status_id, created_at, updated_at 
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
