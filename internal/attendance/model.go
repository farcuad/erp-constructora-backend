package attendance

import "time"

type Attendance struct {
	ID        string          `json:"id"`
	CompanyID string          `json:"company_id"`
	ProjectID string          `json:"project_id"`
	Date      string          `json:"date"` // Formato YYYY-MM-DD
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Logs      []AttendanceLog `json:"logs,omitempty"`
}

type UpdateAttendanceLogRequest struct {
	Status      string  `json:"status"`
	HoursWorked float64 `json:"hours_worked"`
	Notes       string  `json:"notes,omitempty"`
}

type AttendanceLog struct {
	ID           string    `json:"id"`
	AttendanceID string    `json:"attendance_id"`
	EmployeeID   string    `json:"employee_id"`
	Status       string    `json:"status"` // Present, Absent, Late, Justified Absence
	HoursWorked  float64   `json:"hours_worked"`
	Notes        string    `json:"notes,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
