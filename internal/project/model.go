package project

import "time"

type Project struct {
	ID        string    `json:"id"`
	CompanyID string    `json:"company_id"`
	Name      string    `json:"name"`
	ClientID  string    `json:"client_id"` // Asociado al Módulo 3 de Clientes
	Location  string    `json:"location"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Budget    float64   `json:"budget"`
	StatusID  int       `json:"status_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DTO para recibir la creación de un proyecto desde el frontend
type CreateProjectDTO struct {
	Name      string    `json:"name"`
	ClientID  string    `json:"client_id"`
	Location  string    `json:"location"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Budget    float64   `json:"budget"`
}

type UpdateProjectDTO struct {
	Name      *string    `json:"name"`
	ClientID  *string    `json:"client_id"`
	Location  *string    `json:"location"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	Budget    *float64   `json:"budget"`
}
