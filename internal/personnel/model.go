package personnel

import "time"

type Position struct {
	ID         string    `json:"id"`
	CompanyID  string    `json:"company_id"`
	Name       string    `json:"name"`
	BaseSalary float64   `json:"base_salary"`
	CreatedAt  time.Time `json:"created_at"`
}

type Employee struct {
	ID         string    `json:"id"`
	CompanyID  string    `json:"company_id"`
	PositionID string    `json:"position_id,omitempty"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	DNI        string    `json:"dni"`
	Phone      string    `json:"phone,omitempty"`
	Email      string    `json:"email,omitempty"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type UpdatePositionRequest struct {
	Name       string  `json:"name"`
	BaseSalary float64 `json:"base_salary"`
}

type UpdateEmployeeRequest struct {
	PositionID string `json:"position_id,omitempty"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	DNI        string `json:"dni"`
	Phone      string `json:"phone,omitempty"`
	Email      string `json:"email,omitempty"`
	Status     string `json:"status"`
}

type UpdateContractRequest struct {
	ContractType string  `json:"contract_type"`
	Salary       float64 `json:"salary"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date,omitempty"`
	Status       string  `json:"status"`
}

type Contract struct {
	ID           string    `json:"id"`
	EmployeeID   string    `json:"employee_id"`
	ProjectID    string    `json:"project_id,omitempty"`
	ContractType string    `json:"contract_type"`
	Salary       float64   `json:"salary"`
	StartDate    string    `json:"start_date"` // YYYY-MM-DD
	EndDate      string    `json:"end_date,omitempty"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}
