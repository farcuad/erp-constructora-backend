package expense

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Repository interface {
	Create(ctx context.Context, companyID string, userID string, e *CreateExpenseRequest) (*Expense, error)
	GetByProject(ctx context.Context, companyID string, projectID string) ([]Expense, error)
	GetByID(ctx context.Context, companyID string, id string) (*Expense, error)
	Update(ctx context.Context, companyID string, id string, req *UpdateExpenseRequest) error
	Delete(ctx context.Context, companyID string, id string) error
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

func (r *repository) GetByID(ctx context.Context, companyID string, id string) (*Expense, error) {
	query := `
		SELECT id, company_id, project_id, category_id, user_id, title, amount, expense_date, description, created_at, updated_at
		FROM expenses
		WHERE company_id = $1 AND id = $2`

	var exp Expense
	var expenseTime time.Time

	err := r.db.QueryRowContext(ctx, query, companyID, id).Scan(
		&exp.ID, &exp.CompanyID, &exp.ProjectID, &exp.CategoryID, &exp.UserID, &exp.Title, &exp.Amount, &expenseTime, &exp.Description, &exp.CreatedAt, &exp.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	exp.ExpenseDate = expenseTime.Format("2006-01-02")
	return &exp, nil
}

func (r *repository) Update(ctx context.Context, companyID string, id string, req *UpdateExpenseRequest) error {
	var setClauses []string
	var args []interface{}
	argIdx := 1

	if req.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title=$%d", argIdx))
		args = append(args, *req.Title)
		argIdx++
	}
	if req.Amount != nil {
		setClauses = append(setClauses, fmt.Sprintf("amount=$%d", argIdx))
		args = append(args, *req.Amount)
		argIdx++
	}
	if req.ExpenseDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("expense_date=$%d", argIdx))
		args = append(args, *req.ExpenseDate)
		argIdx++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description=$%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.CategoryID != nil {
		setClauses = append(setClauses, fmt.Sprintf("category_id=$%d", argIdx))
		args = append(args, *req.CategoryID)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at=NOW()"))
	args = append(args, companyID, id)

	query := fmt.Sprintf("UPDATE expenses SET %s WHERE company_id=$%d AND id=$%d",
		strings.Join(setClauses, ", "), argIdx, argIdx+1)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *repository) Delete(ctx context.Context, companyID string, id string) error {
	query := `DELETE FROM expenses WHERE company_id = $1 AND id = $2`
	_, err := r.db.ExecContext(ctx, query, companyID, id)
	return err
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
