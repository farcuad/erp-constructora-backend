package suppliers

import (
	"time"
)

// Supplier representa la tabla 'suppliers'
type Supplier struct {
	ID        string    `json:"id"`
	CompanyID string    `json:"company_id"`
	Name      string    `json:"name"`
	NIT       string    `json:"nit"`
	Address   string    `json:"address,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Email     string    `json:"email,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SupplierContact representa la tabla 'supplier_contacts'
type SupplierContact struct {
	ID         string    `json:"id"`
	SupplierID string    `json:"supplier_id"`
	Name       string    `json:"name"`
	Position   string    `json:"position,omitempty"`
	Phone      string    `json:"phone,omitempty"`
	Email      string    `json:"email,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type UpdateSupplierRequest struct {
	Name    *string `json:"name,omitempty"`
	NIT     *string `json:"nit,omitempty"`
	Address *string `json:"address,omitempty"`
	Phone   *string `json:"phone,omitempty"`
	Email   *string `json:"email,omitempty"`
}
