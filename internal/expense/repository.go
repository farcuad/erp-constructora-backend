package expense

import (
	"context"
	"database/sql"

	"time"
)

type Repository interface {
	Create(ctx context.Context, companyID string, userID string, e *CreateExpenseRequest) (*Expense, error)
	GetByProject(ctx context.Context, companyID string, projectID string) ([]Expense, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, companyID string, userID string, e *CreateExpenseRequest) (*Expense, error) {
	query := `
		INSERT INTO expenses (company_id, project_id, category_id, user_id, title, amount, expense_date, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, company_id, project_id, category_id, user_id, title, amount, expense_date, description, created_at, updated_at`

	var exp Expense
	var expenseTime time.Time

	err := r.db.QueryRowContext(ctx, query, companyID, e.ProjectID, e.CategoryID, userID, e.Title, e.Amount, e.ExpenseDate, e.Description).Scan(
		&exp.ID, &exp.CompanyID, &exp.ProjectID, &exp.CategoryID, &exp.UserID, &exp.Title, &exp.Amount, &expenseTime, &exp.Description, &exp.CreatedAt, &exp.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	exp.ExpenseDate = expenseTime.Format("2006-01-02")
	return &exp, nil
}

func (r *repository) GetByProject(ctx context.Context, companyID string, projectID string) ([]Expense, error) {
	query := `
		SELECT id, company_id, project_id, category_id, user_id, title, amount, expense_date, description, created_at, updated_at
		FROM expenses 
		WHERE company_id = $1 AND project_id = $2
		ORDER BY expense_date DESC`

	rows, err := r.db.QueryContext(ctx, query, companyID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []Expense
	for rows.Next() {
		var exp Expense
		var expenseTime time.Time
		err := rows.Scan(
			&exp.ID, &exp.CompanyID, &exp.ProjectID, &exp.CategoryID, &exp.UserID, &exp.Title, &exp.Amount, &expenseTime, &exp.Description, &exp.CreatedAt, &exp.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		exp.ExpenseDate = expenseTime.Format("2006-01-02")
		expenses = append(expenses, exp)
	}
	return expenses, nil
}
