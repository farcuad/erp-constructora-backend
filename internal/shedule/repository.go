package schedule

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateTask(ctx context.Context, t *Task) error {
	query := `INSERT INTO tasks (project_id, name, description, start_date, end_date) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING id, progress, status, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, t.ProjectID, t.Name, t.Description, t.StartDate, t.EndDate).
		Scan(&t.ID, &t.Progress, &t.Status, &t.CreatedAt, &t.UpdatedAt)
}

func (r *Repository) AddDependency(ctx context.Context, d *TaskDependency) error {
	query := `INSERT INTO task_dependencies (task_id, depends_on_uuid, dependency_type) 
	          VALUES ($1, $2, $3) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, d.TaskID, d.DependsOnUUID, d.DependencyType).
		Scan(&d.ID, &d.CreatedAt)
}

func (r *Repository) CreateMilestone(ctx context.Context, m *Milestone) error {
	query := `INSERT INTO milestones (project_id, name, due_date) 
	          VALUES ($1, $2, $3) RETURNING id, is_achieved, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, m.ProjectID, m.Name, m.DueDate).
		Scan(&m.ID, &m.IsAchieved, &m.CreatedAt, &m.UpdatedAt)
}

// GetScheduleByProject extrae todas las tareas de una obra para renderizar el Gantt
func (r *Repository) GetScheduleByProject(ctx context.Context, projectID string) ([]Task, error) {
	query := `SELECT id, project_id, name, COALESCE(description, ''), start_date, end_date, progress, status, created_at, updated_at 
	          FROM tasks WHERE project_id = $1 ORDER BY start_date ASC`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.Name, &t.Description, &t.StartDate, &t.EndDate, &t.Progress, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *Repository) UpdateTask(ctx context.Context, t *Task) error {
	query := `UPDATE tasks SET name = $1, description = $2, start_date = $3, end_date = $4, progress = $5, status = $6, updated_at = CURRENT_TIMESTAMP WHERE id = $7`
	_, err := r.db.ExecContext(ctx, query, t.Name, t.Description, t.StartDate, t.EndDate, t.Progress, t.Status, t.ID)
	return err
}

func (r *Repository) DeleteTask(ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *Repository) UpdateMilestone(ctx context.Context, m *Milestone) error {
	query := `UPDATE milestones SET name = $1, due_date = $2, is_achieved = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, m.Name, m.DueDate, m.IsAchieved, m.ID)
	return err
}

func (r *Repository) DeleteMilestone(ctx context.Context, id string) error {
	query := `DELETE FROM milestones WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
