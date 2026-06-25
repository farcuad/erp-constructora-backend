package budgets

import "time"

type Budget struct {
	ID          string    `json:"id"`
	CompanyID   string    `json:"company_id"`
	ProjectID   string    `json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	TotalAmount float64   `json:"total_amount"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BudgetVersion struct {
	ID            string    `json:"id"`
	BudgetID      string    `json:"budget_id"`
	VersionNumber int       `json:"version_number"`
	Status        string    `json:"status"`
	TotalAmount   float64   `json:"total_amount"`
	ChangedBy     string    `json:"changed_by"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

type BudgetItemRequest struct {
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Unit        string  `json:"unit"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

type CreateBudgetWithItemsRequest struct {
	ProjectID   string              `json:"project_id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Items       []BudgetItemRequest `json:"items"`
}
