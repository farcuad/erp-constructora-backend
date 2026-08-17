package financialdashboard

type ProjectKPIs struct {
	CompanyID         string  `json:"company_id"`
	ProjectID         string  `json:"project_id"`
	TotalBudget       float64 `json:"total_budget"`       // Cuánto estimamos gastar/cobrar (Mod 4)
	TotalExpenses     float64 `json:"total_expenses"`     // Gastos directos registrados (Mod 5)
	TotalPurchased    float64 `json:"total_purchased"`    // Órdenes de compra aprobadas (Mod 6)
	TotalInvoiced     float64 `json:"total_invoiced"`     // Facturas emitidas al cliente (Mod 16)
	TotalCollected    float64 `json:"total_collected"`    // Dinero real que ha entrado de clientes
	TotalPaidToProv   float64 `json:"total_paid_to_prov"` // Dinero real pagado a proveedores/contratistas
	FinancialVariance float64 `json:"financial_variance"` // Presupuesto vs Gasto Real (Desviación)
}

type MonthlyData struct {
	Month     string  `json:"month"`     // Ej: "Jan", "Feb", "Mar"
	Invoiced  float64 `json:"invoiced"`  // Facturado
	Collected float64 `json:"collected"` // Cobrado real
	Expenses  float64 `json:"expenses"`  // Gastos reales
}

type CategoryExpense struct {
	Category string  `json:"category"` // Materiales, Personal, etc.
	Budgeted float64 `json:"budgeted"`
	Spent    float64 `json:"spent"`
}

type AdvancedDashboardKPIs struct {
	ProjectKPIs                          // Hereda la estructura anterior
	MonthlyTrends      []MonthlyData     `json:"monthly_trends"`
	ExpensesByCategory []CategoryExpense `json:"expenses_by_category"`
}
