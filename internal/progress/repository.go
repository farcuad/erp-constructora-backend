package progress

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ExecInTx helper para ejecutar lógica dentro de una transacción de forma segura
func (r *Repository) ExecInTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Se descarta automáticamente si no hay Commit

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) CreateReportTx(ctx context.Context, tx *sql.Tx, report *DailyReport) error {
	query := `
		INSERT INTO daily_reports (company_id, project_id, user_id, report_date, weather_condition, observations)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	return tx.QueryRowContext(ctx, query,
		report.CompanyID,
		report.ProjectID,
		report.UserID,
		report.ReportDate,
		report.WeatherCondition,
		report.Observations,
	).Scan(&report.ID, &report.CreatedAt)
}

func (r *Repository) CreateProgressEntryTx(ctx context.Context, tx *sql.Tx, entry *ProgressEntry) error {
	query := `
		INSERT INTO progress_entries (company_id, project_id, daily_report_id, task_id, progress_percentage, quantity_executed, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	return tx.QueryRowContext(ctx, query,
		entry.CompanyID,
		entry.ProjectID,
		entry.DailyReportID,
		entry.TaskID,
		entry.ProgressPercentage,
		entry.QuantityExecuted,
		entry.Notes,
	).Scan(&entry.ID)
}

// ApplyTaskProgressTx refleja el avance acumulado de una tarea en la tabla tasks.
// El progreso nunca baja (GREATEST) y el estado se deriva automáticamente.
func (r *Repository) ApplyTaskProgressTx(ctx context.Context, tx *sql.Tx, taskID string, newProgress float64) error {
	query := `
		UPDATE tasks SET
			progress = GREATEST(progress, $1),
			status = CASE
				WHEN GREATEST(progress, $1) >= 100 THEN 'Done'
				WHEN GREATEST(progress, $1) > 0 THEN 'In Progress'
				ELSE 'To Do'
			END,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $2`

	_, err := tx.ExecContext(ctx, query, newProgress, taskID)
	return err
}

// GetReportTaskIDs obtiene las tareas asociadas a un reporte (antes de borrarlo,
// porque el DELETE propaga en cascada a progress_entries).
func (r *Repository) GetReportTaskIDs(ctx context.Context, tx *sql.Tx, reportID string) ([]string, error) {
	query := `SELECT DISTINCT task_id FROM progress_entries WHERE daily_report_id = $1`

	rows, err := tx.QueryContext(ctx, query, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// RecalcTaskProgressTx recalcula el progreso de una tarea a partir de los
// progress_entries restantes (máximo acumulado), ya aplicado el borrado del reporte.
func (r *Repository) RecalcTaskProgressTx(ctx context.Context, tx *sql.Tx, taskID string) error {
	query := `
		WITH cur AS (
			SELECT COALESCE(MAX(progress_percentage), 0) AS p
			FROM progress_entries WHERE task_id = $1
		)
		UPDATE tasks t SET
			progress = cur.p,
			status = CASE
				WHEN cur.p >= 100 THEN 'Done'
				WHEN cur.p > 0 THEN 'In Progress'
				ELSE 'To Do'
			END,
			updated_at = CURRENT_TIMESTAMP
		FROM cur WHERE t.id = $1`

	_, err := tx.ExecContext(ctx, query, taskID)
	return err
}

func (r *Repository) GetReportWithProgress(ctx context.Context, companyID, projectID string, date string) (*DailyReport, error) {
	// Consulta con JOIN para traer los datos del reporte diario y sus entradas asociadas en un solo viaje
	query := `
		SELECT 
			dr.id, dr.company_id, dr.project_id, dr.user_id, dr.report_date, dr.weather_condition, dr.observations, dr.created_at,
			pe.id, pe.task_id, pe.progress_percentage, pe.quantity_executed, pe.notes
		FROM daily_reports dr
		LEFT JOIN progress_entries pe ON dr.id = pe.daily_report_id
		WHERE dr.company_id = $1 AND dr.project_id = $2 AND dr.report_date = $3`

	rows, err := r.db.QueryContext(ctx, query, companyID, projectID, date)
	if err != nil {
		log.Printf("[DB QUERY ERROR] progress.GetReportWithProgress (company_id=%s project_id=%s date=%s): %v", companyID, projectID, date, err)
		return nil, err
	}
	defer rows.Close()

	var report *DailyReport

	for rows.Next() {
		var (
			drID, drCompanyID, drProjectID, drUserID, drWeather, drObs string
			drDate, drCreatedAt                                        time.Time

			// Variables nulables por si el reporte existe pero aún no tiene renglones de avance (LEFT JOIN)
			peID, peTaskID, peNotes                  sql.NullString
			peProgressPercentage, peQuantityExecuted sql.NullFloat64
		)

		err := rows.Scan(
			&drID, &drCompanyID, &drProjectID, &drUserID, &drDate, &drWeather, &drObs, &drCreatedAt,
			&peID, &peTaskID, &peProgressPercentage, &peQuantityExecuted, &peNotes,
		)
		if err != nil {
			log.Printf("[DB SCAN ERROR] progress.GetReportWithProgress (company_id=%s project_id=%s date=%s): %v", companyID, projectID, date, err)
			return nil, err
		}

		// Inicializamos el objeto contenedor solo en la primera iteración
		if report == nil {
			report = &DailyReport{
				ID:               drID,
				CompanyID:        drCompanyID,
				ProjectID:        drProjectID,
				UserID:           drUserID,
				ReportDate:       drDate,
				WeatherCondition: drWeather,
				Observations:     drObs,
				CreatedAt:        drCreatedAt,
				ProgressEntries:  []ProgressEntry{},
			}
		}

		// Si existe una línea de progreso válida en esta fila del JOIN, la añadimos a la lista
		if peID.Valid {
			entry := ProgressEntry{
				ID:                 peID.String,
				CompanyID:          drCompanyID,
				ProjectID:          drProjectID,
				DailyReportID:      drID,
				TaskID:             peTaskID.String,
				ProgressPercentage: peProgressPercentage.Float64,
				QuantityExecuted:   peQuantityExecuted.Float64,
				Notes:              peNotes.String,
			}
			report.ProgressEntries = append(report.ProgressEntries, entry)
		}
	}

	if err = rows.Err(); err != nil {
		log.Printf("[DB ROWS ITERATION ERROR] progress.GetReportWithProgress (company_id=%s project_id=%s date=%s): %v", companyID, projectID, date, err)
		return nil, fmt.Errorf("error iterando filas: %w", err)
	}

	if report == nil {
		return nil, sql.ErrNoRows
	}

	return report, nil
}

func (r *Repository) UpdateReport(ctx context.Context, companyID, id string, req UpdateDailyReportRequest) error {
	query := `
		UPDATE daily_reports
		SET weather_condition = COALESCE($1, weather_condition),
		    observations = COALESCE($2, observations),
		    report_date = COALESCE($3, report_date)
		WHERE company_id = $4 AND id = $5`

	var weather, obs interface{}
	if req.WeatherCondition != nil {
		weather = *req.WeatherCondition
	} else {
		weather = nil
	}
	if req.Observations != nil {
		obs = *req.Observations
	} else {
		obs = nil
	}

	_, err := r.db.ExecContext(ctx, query, weather, obs, req.ReportDate, companyID, id)
	return err
}

func (r *Repository) DeleteReportTx(ctx context.Context, tx *sql.Tx, companyID, id string) error {
	query := `DELETE FROM daily_reports WHERE company_id = $1 AND id = $2`
	_, err := tx.ExecContext(ctx, query, companyID, id)
	return err
}
