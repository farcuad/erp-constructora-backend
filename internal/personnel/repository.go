package personnel

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

func (r *Repository) CreatePosition(ctx context.Context, p *Position) error {
	query := `INSERT INTO positions (company_id, name, base_salary) VALUES ($1, $2, $3) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, p.CompanyID, p.Name, p.BaseSalary).Scan(&p.ID, &p.CreatedAt)
}

func (r *Repository) CreateEmployee(ctx context.Context, e *Employee) error {
	query := `INSERT INTO employees (company_id, position_id, first_name, last_name, dni, phone, email) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, status, created_at, updated_at`

	var posID interface{} = nil
	if e.PositionID != "" {
		posID = e.PositionID
	}

	return r.db.QueryRowContext(ctx, query, e.CompanyID, posID, e.FirstName, e.LastName, e.DNI, e.Phone, e.Email).
		Scan(&e.ID, &e.Status, &e.CreatedAt, &e.UpdatedAt)
}

func (r *Repository) GetEmployeesByCompany(ctx context.Context, companyID string) ([]Employee, error) {
	query := `SELECT id, company_id, COALESCE(position_id::text, ''), first_name, last_name, dni, COALESCE(phone, ''), COALESCE(email, ''), status, created_at, updated_at 
	          FROM employees WHERE company_id = $1`
	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []Employee
	for rows.Next() {
		var e Employee
		if err := rows.Scan(&e.ID, &e.CompanyID, &e.PositionID, &e.FirstName, &e.LastName, &e.DNI, &e.Phone, &e.Email, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		employees = append(employees, e)
	}
	return employees, nil
}

func (r *Repository) CreateContract(ctx context.Context, c *Contract) error {
	query := `INSERT INTO contracts (employee_id, project_id, contract_type, salary, start_date, end_date) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, status, created_at`

	var prjID interface{} = nil
	if c.ProjectID != "" {
		prjID = c.ProjectID
	}
	var endDate interface{} = nil
	if c.EndDate != "" {
		endDate = c.EndDate
	}

	return r.db.QueryRowContext(ctx, query, c.EmployeeID, prjID, c.ContractType, c.Salary, c.StartDate, endDate).
		Scan(&c.ID, &c.Status, &c.CreatedAt)
}
