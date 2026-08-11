package audit

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Insert(ctx context.Context, log *AuditLog) error {
	query := `
        INSERT INTO audit_logs (company_id, user_id, action, table_name, row_id, ip_address, old_values, new_values)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, created_at`

	return r.db.QueryRowContext(ctx, query,
		log.CompanyID,
		log.UserID,
		log.Action,
		log.TableName,
		log.RowID,
		log.IPAddress,
		log.OldValues,
		log.NewValues,
	).Scan(&log.ID, &log.CreatedAt)
}

func (r *Repository) GetByCompany(ctx context.Context, companyID string) ([]AuditLog, error) {
	query := `
        SELECT id, company_id, user_id, action, table_name, row_id, ip_address, old_values, new_values, created_at
        FROM audit_logs
        WHERE company_id = $1
        ORDER BY created_at DESC LIMIT 100` // Limitado por seguridad/rendimiento

	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		log.Printf("[DB QUERY ERROR] audit.GetByCompany (company_id=%s): %v", companyID, err)
		return nil, err
	}
	defer rows.Close()

	logs := make([]AuditLog, 0)
	for rows.Next() {
		var l AuditLog
		err := rows.Scan(&l.ID, &l.CompanyID, &l.UserID, &l.Action, &l.TableName, &l.RowID, &l.IPAddress, &l.OldValues, &l.NewValues, &l.CreatedAt)
		if err != nil {
			log.Printf("[DB SCAN ERROR] audit.GetByCompany (company_id=%s): %v", companyID, err)
			return nil, err
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[DB ROWS ITERATION ERROR] audit.GetByCompany (company_id=%s): %v", companyID, err)
		return nil, fmt.Errorf("error iterando filas: %w", err)
	}
	return logs, nil
}
