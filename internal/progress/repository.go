package progress

import (
	"context"
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateReport(ctx context.Context, report *DailyReport) error {
	query := `
		INSERT INTO daily_reports (company_id, project_id, user_id, report_date, weather_condition, observations)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	return r.db.QueryRowContext(ctx, query,
		report.CompanyID,
		report.ProjectID,
		report.UserID,
		report.ReportDate,
		report.WeatherCondition,
		report.Observations,
	).Scan(&report.ID, &report.CreatedAt)
}

func (r *Repository) CreateProgressEntry(ctx context.Context, entry *ProgressEntry) error {
	query := `
		INSERT INTO progress_entries (company_id, project_id, daily_report_id, task_id, progress_percentage, quantity_executed, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		entry.CompanyID,
		entry.ProjectID,
		entry.DailyReportID,
		entry.TaskID,
		entry.ProgressPercentage,
		entry.QuantityExecuted,
		entry.Notes,
	).Scan(&entry.ID)
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
		return nil, err
	}

	// Si no se encontró ningún registro, retornamos un error controlado de "no encontrado"
	if report == nil {
		return nil, sql.ErrNoRows
	}

	return report, nil
}
