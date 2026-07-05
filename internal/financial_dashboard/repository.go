package financialdashboard

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

func (r *Repository) GetProjectFinancialSummary(ctx context.Context, companyID, projectID string) (*ProjectKPIs, error) {
	// Query analítico que unifica los módulos mediante subconsultas por proyecto
	query := `
		SELECT 
			$1::UUID as company_id,
			$2::UUID as project_id,
			COALESCE((SELECT SUM(total_amount) FROM budgets WHERE project_id = $2 AND company_id = $1), 0.00) as total_budget,
			COALESCE((SELECT SUM(amount) FROM expenses WHERE project_id = $2 AND company_id = $1), 0.00) as total_expenses,
			COALESCE((SELECT SUM(total_amount) FROM purchase_orders WHERE project_id = $2 AND company_id = $1), 0.00) as total_purchased,
			COALESCE((SELECT SUM(total_amount) FROM invoices WHERE project_id = $2 AND company_id = $1 AND type = 'EMITTED'), 0.00) as total_invoiced,
			COALESCE((SELECT SUM(p.amount) FROM payments p JOIN invoices i ON p.invoice_id = i.id WHERE i.project_id = $2 AND i.company_id = $1 AND i.type = 'EMITTED'), 0.00) as total_collected,
			COALESCE((SELECT SUM(p.amount) FROM payments p JOIN invoices i ON p.invoice_id = i.id WHERE i.project_id = $2 AND i.company_id = $1 AND i.type = 'RECEIVED'), 0.00) as total_paid_to_prov
	`

	var kpi ProjectKPIs
	err := r.db.QueryRowContext(ctx, query, companyID, projectID).Scan(
		&kpi.CompanyID,
		&kpi.ProjectID,
		&kpi.TotalBudget,
		&kpi.TotalExpenses,
		&kpi.TotalPurchased,
		&kpi.TotalInvoiced,
		&kpi.TotalCollected,
		&kpi.TotalPaidToProv,
	)
	if err != nil {
		return nil, err
	}

	// Lógica calculada: Desviación (Presupuesto - Lo gastado real en órdenes y cajas chicas)
	kpi.FinancialVariance = kpi.TotalBudget - (kpi.TotalExpenses + kpi.TotalPurchased)

	return &kpi, nil
}
