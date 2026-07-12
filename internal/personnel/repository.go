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

func (r *Repository) GetPosition(ctx context.Context, companyID string) ([]Position, error) {
	query := `SELECT id, company_id, name, base_salary, created_at
	          FROM positions WHERE company_id = $1`
	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []Position
	for rows.Next() {
		var e Position
		if err := rows.Scan(&e.ID, &e.CompanyID, &e.Name, &e.BaseSalary, &e.CreatedAt); err != nil {
			return nil, err
		}
		positions = append(positions, e)
	}
	return positions, nil
}
func (r *Repository) UpdatePosition(ctx context.Context, p *Position) error {
	query := `UPDATE positions SET name = $1, base_salary = $2 WHERE company_id = $3 AND id = $4`
	_, err := r.db.ExecContext(ctx, query, p.Name, p.BaseSalary, p.CompanyID, p.ID)
	return err
}

func (r *Repository) DeletePosition(ctx context.Context, companyID, id string) error {
	query := `DELETE FROM positions WHERE company_id = $1 AND id = $2`
	_, err := r.db.ExecContext(ctx, query, companyID, id)
	return err
}

func (r *Repository) UpdateEmployee(ctx context.Context, e *Employee) error {
	query := `UPDATE employees SET position_id = $1, first_name = $2, last_name = $3, dni = $4, phone = $5, email = $6, status = $7, updated_at = CURRENT_TIMESTAMP WHERE company_id = $8 AND id = $9`
	var posID interface{} = nil
	if e.PositionID != "" {
		posID = e.PositionID
	}
	_, err := r.db.ExecContext(ctx, query, posID, e.FirstName, e.LastName, e.DNI, e.Phone, e.Email, e.Status, e.CompanyID, e.ID)
	return err
}

func (r *Repository) DeleteEmployee(ctx context.Context, companyID, id string) error {
	query := `DELETE FROM employees WHERE company_id = $1 AND id = $2`
	_, err := r.db.ExecContext(ctx, query, companyID, id)
	return err
}

func (r *Repository) GetContract(ctx context.Context, companyID string) ([]Contract, error) {
	query := `SELECT id, employee_id, project_id, contract_type, salary, start_date, end_date, status
	          FROM contracts WHERE company_id = $1`
	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contracts []Contract
	for rows.Next() {
		var e Contract
		if err := rows.Scan(&e.ID, &e.EmployeeID, &e.ProjectID, &e.ContractType, &e.Salary,
			&e.StartDate, &e.EndDate, &e.CreatedAt); err != nil {
			return nil, err
		}
		contracts = append(contracts, e)
	}
	return contracts, nil
}
func (r *Repository) UpdateContract(ctx context.Context, c *Contract) error {
	query := `UPDATE contracts SET contract_type = $1, salary = $2, start_date = $3, end_date = $4, status = $5 WHERE id = $6`
	_, err := r.db.ExecContext(ctx, query, c.ContractType, c.Salary, c.StartDate, c.EndDate, c.Status, c.ID)
	return err
}

func (r *Repository) DeleteContract(ctx context.Context, id string) error {
	query := `DELETE FROM contracts WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
