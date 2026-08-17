package financialdashboard

import (
	"context"
	"database/sql"
	"log"
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
		log.Printf("[DB QUERY ERROR] financialdashboard.GetProjectFinancialSummary (company_id=%s project_id=%s): %v", companyID, projectID, err)
		return nil, err
	}

	// Lógica calculada: Desviación (Presupuesto - Lo gastado real en órdenes y cajas chicas)
	kpi.FinancialVariance = kpi.TotalBudget - (kpi.TotalExpenses + kpi.TotalPurchased)

	return &kpi, nil
}

func (r *Repository) GetMonthlyTrends(ctx context.Context, companyID, projectID string) ([]MonthlyData, error) {
	query := `
		SELECT 
			TO_CHAR(date_series, 'Mon') as month,
			COALESCE(SUM(i.total_amount), 0.00) as invoiced,
			COALESCE(SUM(p.amount), 0.00) as collected,
			COALESCE(SUM(e.amount), 0.00) as expenses
		FROM generate_series(
			DATE_TRUNC('month', CURRENT_DATE - INTERVAL '5 months'),
			DATE_TRUNC('month', CURRENT_DATE),
			'1 month'::interval
		) date_series
		LEFT JOIN invoices i ON DATE_TRUNC('month', i.created_at) = date_series 
			AND i.project_id = $2::UUID AND i.company_id = $1::UUID AND i.type = 'EMITTED'
		LEFT JOIN payments p ON DATE_TRUNC('month', p.created_at) = date_series 
			AND p.invoice_id = i.id
		LEFT JOIN expenses e ON DATE_TRUNC('month', e.created_at) = date_series 
			AND e.project_id = $2::UUID AND e.company_id = $1::UUID
		GROUP BY date_series
		ORDER BY date_series ASC;
	`
	rows, err := r.db.QueryContext(ctx, query, companyID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trends []MonthlyData
	for rows.Next() {
		var m MonthlyData
		if err := rows.Scan(&m.Month, &m.Invoiced, &m.Collected, &m.Expenses); err != nil {
			return nil, err
		}
		trends = append(trends, m)
	}
	return trends, nil
}
