package contractors

import "time"

type Contractor struct {
	ID             string    `json:"id"`
	CompanyID      string    `json:"company_id"`
	Name           string    `json:"name"`
	NIT            string    `json:"nit"`
	Representative string    `json:"representative,omitempty"`
	Phone          string    `json:"phone,omitempty"`
	Email          string    `json:"email,omitempty"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UpdateContractorRequest struct {
	Name           string `json:"name"`
	NIT            string `json:"nit"`
	Representative string `json:"representative,omitempty"`
	Phone          string `json:"phone,omitempty"`
	Email          string `json:"email,omitempty"`
	IsActive       *bool  `json:"is_active,omitempty"`
}

type UpdateContractorContractRequest struct {
	Title       string  `json:"title"`
	TotalAmount float64 `json:"total_amount"`
	Balance     float64 `json:"balance"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date,omitempty"`
	Status      string  `json:"status"`
}

type ContractorContract struct {
	ID           string    `json:"id"`
	CompanyID    string    `json:"company_id"`
	ContractorID string    `json:"contractor_id"`
	ProjectID    string    `json:"project_id"`
	Title        string    `json:"title"`
	TotalAmount  float64   `json:"total_amount"`
	Balance      float64   `json:"balance"`
	StartDate    string    `json:"start_date"` // YYYY-MM-DD
	EndDate      string    `json:"end_date,omitempty"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ContractorPayment struct {
	ID              string    `json:"id"`
	ContractID      string    `json:"contract_id"`
	UserID          string    `json:"user_id"`
	Amount          float64   `json:"amount"`
	PaymentDate     string    `json:"payment_date"` // YYYY-MM-DD
	ReferenceNumber string    `json:"reference_number,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}
