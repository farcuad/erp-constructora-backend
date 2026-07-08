package expense

type Expense struct {
	ID          string  `json:"id"`
	CompanyID   string  `json:"company_id"`
	ProjectID   string  `json:"project_id"`
	CategoryID  int     `json:"category_id"`
	UserID      string  `json:"user_id"`
	Title       string  `json:"title"`
	Amount      float64 `json:"amount"`
	ExpenseDate string  `json:"expense_date"`
	Description string  `json:"description"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type CreateExpenseRequest struct {
	CompanyID   string  `json:"company_id"`
	ProjectID   string  `json:"project_id"`
	CategoryID  int     `json:"category_id"`
	UserID      string  `json:"user_id"`
	Title       string  `json:"title"`
	Amount      float64 `json:"amount"`
	ExpenseDate string  `json:"expense_date"`
	Description string  `json:"description"`
}

type UpdateExpenseRequest struct {
	Title       *string  `json:"title"`
	Amount      *float64 `json:"amount"`
	ExpenseDate *string  `json:"expense_date"`
	Description *string  `json:"description"`
	CategoryID  *int     `json:"category_id"`
}
