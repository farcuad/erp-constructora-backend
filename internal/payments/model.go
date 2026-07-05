package payments

import (
	"time"
)

type Invoice struct {
	ID              string        `json:"id"`
	CompanyID       string        `json:"company_id"`
	ProjectID       string        `json:"project_id"`
	InvoiceNumber   string        `json:"invoice_number"`
	Type            string        `json:"type"` // "EMITTED" o "RECEIVED"
	Status          string        `json:"status"`
	ClientID        *string       `json:"client_id,omitempty"`
	SupplierID      *string       `json:"supplier_id,omitempty"`
	ContractorID    *string       `json:"contractor_id,omitempty"`
	IssueDate       time.Time     `json:"issue_date"`
	DueDate         time.Time     `json:"due_date"`
	Subtotal        float64       `json:"subtotal"`
	TaxAmount       float64       `json:"tax_amount"`
	TotalAmount     float64       `json:"total_amount"`
	RemainingAmount float64       `json:"remaining_amount"`
	Notes           string        `json:"notes"`
	CreatedAt       time.Time     `json:"created_at"`
	Items           []InvoiceItem `json:"items,omitempty"`
}

type InvoiceItem struct {
	ID          string  `json:"id"`
	CompanyID   string  `json:"company_id"`
	InvoiceID   string  `json:"invoice_id"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
}

type Payment struct {
	ID            string    `json:"id"`
	CompanyID     string    `json:"company_id"`
	ProjectID     string    `json:"project_id"`
	InvoiceID     string    `json:"invoice_id"`
	PaymentDate   time.Time `json:"payment_date"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	Reference     string    `json:"reference"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
}
