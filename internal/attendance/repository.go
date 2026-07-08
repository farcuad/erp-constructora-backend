package attendance

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

// SaveDailyAttendance inserta la cabecera y todos sus registros en una sola transacción
func (r *Repository) SaveDailyAttendance(ctx context.Context, a *Attendance) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insertar o recuperar cabecera de asistencia (si ya existiera por algún reintento)
	queryHeader := `INSERT INTO attendance (company_id, project_id, date) 
	                VALUES ($1, $2, $3) 
	                ON CONFLICT (project_id, date) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
	                RETURNING id, created_at, updated_at`

	err = tx.QueryRowContext(ctx, queryHeader, a.CompanyID, a.ProjectID, a.Date).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return err
	}

	// 2. Insertar los registros individuales de asistencia para cada empleado
	queryLog := `INSERT INTO attendance_logs (attendance_id, employee_id, status, hours_worked, notes) 
	             VALUES ($1, $2, $3, $4, $5)
	             ON CONFLICT (attendance_id, employee_id) 
	             DO UPDATE SET status = EXCLUDED.status, hours_worked = EXCLUDED.hours_worked, notes = EXCLUDED.notes, updated_at = CURRENT_TIMESTAMP
	             RETURNING id, created_at, updated_at`

	for i := range a.Logs {
		a.Logs[i].AttendanceID = a.ID
		err = tx.QueryRowContext(ctx, queryLog, a.ID, a.Logs[i].EmployeeID, a.Logs[i].Status, a.Logs[i].HoursWorked, a.Logs[i].Notes).
			Scan(&a.Logs[i].ID, &a.Logs[i].CreatedAt, &a.Logs[i].UpdatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetAttendanceByProjectAndDate obtiene la lista de asistencia de un día específico
func (r *Repository) GetAttendanceByProjectAndDate(ctx context.Context, projectID string, date string) (*Attendance, error) {
	queryHeader := `SELECT id, company_id, project_id, date, created_at, updated_at 
	                FROM attendance WHERE project_id = $1 AND date = $2`

	var a Attendance
	err := r.db.QueryRowContext(ctx, queryHeader, projectID, date).
		Scan(&a.ID, &a.CompanyID, &a.ProjectID, &a.Date, &a.CreatedAt, &a.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil // No hay registros para ese día
	} else if err != nil {
		return nil, err
	}

	queryLogs := `SELECT id, attendance_id, employee_id, status, hours_worked, COALESCE(notes, ''), created_at, updated_at 
	              FROM attendance_logs WHERE attendance_id = $1`

	rows, err := r.db.QueryContext(ctx, queryLogs, a.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var log AttendanceLog
		if err := rows.Scan(&log.ID, &log.AttendanceID, &log.EmployeeID, &log.Status, &log.HoursWorked, &log.Notes, &log.CreatedAt, &log.UpdatedAt); err != nil {
			return nil, err
		}
		a.Logs = append(a.Logs, log)
	}

	return &a, nil
}

func (r *Repository) UpdateAttendanceLog(ctx context.Context, log *AttendanceLog) error {
	query := `UPDATE attendance_logs SET status = $1, hours_worked = $2, notes = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, log.Status, log.HoursWorked, log.Notes, log.ID)
	return err
}

func (r *Repository) DeleteAttendance(ctx context.Context, companyID, id string) error {
	query := `DELETE FROM attendance WHERE company_id = $1 AND id = $2`
	_, err := r.db.ExecContext(ctx, query, companyID, id)
	return err
}
