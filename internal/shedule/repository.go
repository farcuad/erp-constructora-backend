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

func (r *Repository) GetByProject(ctx context.Context, projectID string) ([]Task, error) {
	query := `SELECT id, project_id, name, COALESCE(description, ''), start_date, end_date, progress, status, created_at, updated_at
	          FROM tasks WHERE project_id = $1 ORDER BY created_at ASC`

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
