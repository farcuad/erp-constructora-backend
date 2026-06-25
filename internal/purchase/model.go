package purchase

import "time"

type PurchaseOrder struct {
	ID           string    `json:"id"`
	CompanyID    string    `json:"company_id"`
	ProjectID    string    `json:"project_id"`
	SupplierID   string    `json:"supplier_id"`
	UserID       string    `json:"user_id"`
	OrderNumber  int       `json:"order_number"`
	Status       string    `json:"status"`
	TotalAmount  float64   `json:"total_amount"`
	DeliveryDate string    `json:"delivery_date"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PurchaseItemRequest struct {
	Description string  `json:"description"`
	Unit        string  `json:"unit"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

type CreatePurchaseOrderRequest struct {
	ProjectID    string                `json:"project_id"`
	SupplierID   string                `json:"supplier_id"`
	DeliveryDate string                `json:"delivery_date"` // YYYY-MM-DD
	Notes        string                `json:"notes"`
	Items        []PurchaseItemRequest `json:"items"`
}
