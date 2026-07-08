package budgets

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Repository interface {
	CreateWithItems(ctx context.Context, companyID string, userID string, b *CreateBudgetWithItemsRequest) (*Budget, error)
	GetBudgetsProjectID(ctx context.Context, companyID string, projectID string) ([]Budget, error)
	GetByID(ctx context.Context, companyID string, id string) (*Budget, error)
	Update(ctx context.Context, companyID string, id string, req *UpdateBudgetRequest) error
	Delete(ctx context.Context, companyID string, id string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateWithItems(ctx context.Context, companyID string, userID string, b *CreateBudgetWithItemsRequest) (*Budget, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // Se cancela automáticamente si hay un error antes del Commit

	// 1. Calcular el total acumulado de los ítems en Go para guardar en la cabecera
	var totalBudget float64
	for _, item := range b.Items {
		totalBudget += item.Quantity * item.UnitPrice
	}

	// 2. Insertar en budgets
	budgetQuery := `
		INSERT INTO budgets (company_id, project_id, title, description, total_amount)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	var budget Budget
	budget.CompanyID = companyID
	budget.ProjectID = b.ProjectID
	budget.Title = b.Title
	budget.Description = b.Description
	budget.TotalAmount = totalBudget

	err = tx.QueryRowContext(ctx, budgetQuery, companyID, b.ProjectID, b.Title, b.Description, totalBudget).Scan(
		&budget.ID, &budget.CreatedAt, &budget.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// 3. Insertar la versión inicial (Versión 1)
	versionQuery := `
		INSERT INTO budget_versions (budget_id, version_number, status, total_amount, changed_by, description)
		VALUES ($1, 1, 'Draft', $2, $3, 'Creación inicial del presupuesto')
		RETURNING id`

	var versionID string
	err = tx.QueryRowContext(ctx, versionQuery, budget.ID, totalBudget, userID).Scan(&versionID)
	if err != nil {
		return nil, err
	}

	// 4. Insertar todos los ítems vinculados a esa versión
	itemQuery := `
		INSERT INTO budget_items (budget_version_id, category, description, unit, quantity, unit_price)
		VALUES ($1, $2, $3, $4, $5, $6)`

	for _, item := range b.Items {
		_, err = tx.ExecContext(ctx, itemQuery, versionID, item.Category, item.Description, item.Unit, item.Quantity, item.UnitPrice)
		if err != nil {
			return nil, err
		}
	}

	// Confirmar la transacción completa
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &budget, nil
}

func (r *repository) GetByID(ctx context.Context, companyID string, id string) (*Budget, error) {
	query := `SELECT id, company_id, project_id, title, description, total_amount, created_at, updated_at
			  FROM budgets WHERE company_id = $1 AND id = $2`

	var budget Budget
	err := r.db.QueryRowContext(ctx, query, companyID, id).Scan(
		&budget.ID, &budget.CompanyID, &budget.ProjectID, &budget.Title, &budget.Description,
		&budget.TotalAmount, &budget.CreatedAt, &budget.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &budget, nil
}

func (r *repository) Update(ctx context.Context, companyID string, id string, req *UpdateBudgetRequest) error {
	var setClauses []string
	var args []interface{}
	argIdx := 1

	if req.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title=$%d", argIdx))
		args = append(args, *req.Title)
		argIdx++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description=$%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at=NOW()"))
	args = append(args, companyID, id)

	query := fmt.Sprintf("UPDATE budgets SET %s WHERE company_id=$%d AND id=$%d",
		strings.Join(setClauses, ", "), argIdx, argIdx+1)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *repository) Delete(ctx context.Context, companyID string, id string) error {
	query := `DELETE FROM budgets WHERE company_id = $1 AND id = $2`
	_, err := r.db.ExecContext(ctx, query, companyID, id)
	return err
}

func (r *repository) GetBudgetsProjectID(ctx context.Context, companyID string, projectID string) ([]Budget, error) {
	query := "SELECT * FROM budgets WHERE company_id = $1 AND project_id = $2"
	rows, err := r.db.QueryContext(ctx, query, companyID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []Budget
	for rows.Next() {
		var budget Budget
		if err := rows.Scan(&budget.ID, &budget.CompanyID, &budget.ProjectID, &budget.Title, &budget.Description, &budget.TotalAmount, &budget.CreatedAt, &budget.UpdatedAt); err != nil {
			return nil, err
		}
		budgets = append(budgets, budget)
	}
	return budgets, nil
}
