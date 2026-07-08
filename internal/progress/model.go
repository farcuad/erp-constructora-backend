package progress

import (
	"time"
)

type DailyReport struct {
	ID               string          `json:"id"`
	CompanyID        string          `json:"company_id"`
	ProjectID        string          `json:"project_id"`
	UserID           string          `json:"user_id"`
	ReportDate       time.Time       `json:"report_date"`
	WeatherCondition string          `json:"weather_condition"`
	Observations     string          `json:"observations"`
	CreatedAt        time.Time       `json:"created_at"`
	ProgressEntries  []ProgressEntry `json:"progress_entries,omitempty"`
}

type UpdateDailyReportRequest struct {
	WeatherCondition *string    `json:"weather_condition,omitempty"`
	Observations     *string    `json:"observations,omitempty"`
	ReportDate       *time.Time `json:"report_date,omitempty"`
}

type ProgressEntry struct {
	ID                 string  `json:"id"`
	CompanyID          string  `json:"company_id"`
	ProjectID          string  `json:"project_id"`
	DailyReportID      string  `json:"daily_report_id"`
	TaskID             string  `json:"task_id"`
	ProgressPercentage float64 `json:"progress_percentage"`
	QuantityExecuted   float64 `json:"quantity_executed"`
	Notes              string  `json:"notes"`
}
