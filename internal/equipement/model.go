package equipement

import "time"

type EquipmentType struct {
	ID        string    `json:"id"`
	CompanyID string    `json:"company_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Equipment struct {
	ID            string    `json:"id"`
	CompanyID     string    `json:"company_id"`
	TypeID        string    `json:"type_id,omitempty"`
	Name          string    `json:"name"`
	PlateNumber   string    `json:"plate_number,omitempty"`
	Model         string    `json:"model,omitempty"`
	Brand         string    `json:"brand,omitempty"`
	Status        string    `json:"status"`         // Available, Assigned, In Maintenance, etc.
	OwnershipType string    `json:"ownership_type"` // Owned, Rented
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type EquipmentAssignment struct {
	ID          string    `json:"id"`
	EquipmentID string    `json:"equipment_id"`
	ProjectID   string    `json:"project_id"`
	AssignedBy  string    `json:"assigned_by"`
	StartDate   string    `json:"start_date"` // YYYY-MM-DD
	EndDate     string    `json:"end_date,omitempty"`
	Notes       string    `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type MaintenanceRecord struct {
	ID              string    `json:"id"`
	EquipmentID     string    `json:"equipment_id"`
	MaintenanceType string    `json:"maintenance_type"` // Preventive, Corrective
	Description     string    `json:"description"`
	Cost            float64   `json:"cost"`
	MaintenanceDate string    `json:"maintenance_date"` // YYYY-MM-DD
	NextDueDate     string    `json:"next_due_date,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}
