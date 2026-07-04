package photos

import (
	"time"
)

type ProjectPhoto struct {
	ID            string    `json:"id"`
	CompanyID     string    `json:"company_id"`
	ProjectID     string    `json:"project_id"`
	TaskID        *string   `json:"task_id,omitempty"`         // Puntero para soportar nulos en JSON
	DailyReportID *string   `json:"daily_report_id,omitempty"` // Puntero para soportar nulos en JSON
	UserID        string    `json:"user_id"`
	PhotoURL      string    `json:"photo_url"`
	Description   string    `json:"description"`
	Latitude      *float64  `json:"latitude,omitempty"`
	Longitude     *float64  `json:"longitude,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
