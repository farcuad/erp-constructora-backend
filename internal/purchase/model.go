package purchase

import (
	"time"
)

// PurchaseOrder representa la tabla 'purchase_orders'
type PurchaseOrder struct {
	ID           string              `json:"id"`
	CompanyID    string              `json:"company_id"`
	ProjectID    string              `json:"project_id"`
	SupplierID   string              `json:"supplier_id"`
	UserID       string              `json:"user_id"`
	OrderNumber  int                 `json:"order_number,omitempty"`
	Status       string              `json:"status"`
	TotalAmount  float64             `json:"total_amount"`
	DeliveryDate string              `json:"delivery_date,omitempty"` // Formato YYYY-MM-DD
	Notes        string              `json:"notes,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	Items        []PurchaseOrderItem `json:"items,omitempty"`
}

// PurchaseOrderItem representa la tabla 'purchase_order_items'
type PurchaseOrderItem struct {
	ID              string  `json:"id"`
	PurchaseOrderID string  `json:"purchase_order_id"`
	Description     string  `json:"description"`
	Unit            string  `json:"unit"`
	Quantity        float64 `json:"quantity"`
	UnitPrice       float64 `json:"unit_price"`
	TotalPrice      float64 `json:"total_price"` // Calculado automáticamente por la BD
}
