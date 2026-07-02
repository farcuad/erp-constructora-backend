package inventory

import "time"

type MaterialCategory struct {
	ID        string    `json:"id"`
	CompanyID string    `json:"company_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Material struct {
	ID         string    `json:"id"`
	CompanyID  string    `json:"company_id"`
	CategoryID string    `json:"category_id,omitempty"`
	Name       string    `json:"name"`
	Code       string    `json:"code,omitempty"`
	Unit       string    `json:"unit"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Warehouse struct {
	ID        string    `json:"id"`
	CompanyID string    `json:"company_id"`
	ProjectID string    `json:"project_id"`
	Name      string    `json:"name"`
	Location  string    `json:"location,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type StockMovement struct {
	ID           string    `json:"id"`
	WarehouseID  string    `json:"warehouse_id"`
	MaterialID   string    `json:"material_id"`
	UserID       string    `json:"user_id"`
	MovementType string    `json:"movement_type"` // "INPUT" o "OUTPUT"
	Quantity     float64   `json:"quantity"`
	ReferenceID  string    `json:"reference_id,omitempty"`
	Description  string    `json:"description,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// Estructura útil para reportes de stock actual
type MaterialStock struct {
	MaterialID   string  `json:"material_id"`
	MaterialName string  `json:"material_name"`
	Code         string  `json:"code"`
	Unit         string  `json:"unit"`
	Quantity     float64 `json:"quantity"`
}
