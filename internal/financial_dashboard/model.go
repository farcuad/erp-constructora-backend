package financialdashboard

type ProjectKPIs struct {
	CompanyID         string  `json:"company_id"`
	ProjectID         string  `json:"project_id"`
	TotalBudget       float64 `json:"total_budget"`       // Cuánto estimamos gastar/cobrar (Mod 4)
	TotalExpenses     float64 `json:"total_expenses"`     // Gastos directos registrados (Mod 5)
	TotalPurchased    float64 `json:"total_purchased"`    // Órdenes de compra aprobadas (Mod 6)
	TotalInvoiced     float64 `json:"total_invoiced"`     // Facturas emitidas al cliente (Mod 16)
	TotalCollected    float64 `json:"total_collected"`    // Dinero real que ha entrado de clientes (Mod 16 - Payments)
	TotalPaidToProv   float64 `json:"total_paid_to_prov"` // Dinero real pagado a proveedores/contratistas
	FinancialVariance float64 `json:"financial_variance"` // Presupuesto vs Gasto Real (Desviación)
}
