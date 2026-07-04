package schedule

import "time"

type Task struct {
	ID           string           `json:"id"`
	ProjectID    string           `json:"project_id"`
	Name         string           `json:"name"`
	Description  string           `json:"description,omitempty"`
	StartDate    string           `json:"start_date"` // YYYY-MM-DD
	EndDate      string           `json:"end_date"`   // YYYY-MM-DD
	Progress     float64          `json:"progress"`
	Status       string           `json:"status"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	Dependencies []TaskDependency `json:"dependencies,omitempty"`
}

type TaskDependency struct {
	ID             string    `json:"id"`
	TaskID         string    `json:"task_id"`
	DependsOnUUID  string    `json:"depends_on_uuid"`
	DependencyType string    `json:"dependency_type"` // FS, SS, FF, SF
	CreatedAt      time.Time `json:"created_at"`
}

type Milestone struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"project_id"`
	Name       string    `json:"name"`
	DueDate    string    `json:"due_date"` // YYYY-MM-DD
	IsAchieved bool      `json:"is_achieved"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
