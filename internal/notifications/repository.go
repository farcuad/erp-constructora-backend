package notifications

import (
	"context"
	"database/sql"

	"github.com/lib/pq" // Requerido para pasar arrays nativos a PostgreSQL
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

// CreateTx inserta la notificación usando la transacción activa
func (r *Repository) CreateTx(ctx context.Context, tx *sql.Tx, n *Notification) error {
	query := `
        INSERT INTO notifications (company_id, project_id, title, message, link_to_ui)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at`

	return tx.QueryRowContext(ctx, query, n.CompanyID, n.ProjectID, n.Title, n.Message, n.LinkToUI).
		Scan(&n.ID, &n.CreatedAt)
}

// AssignToUsersBulk vincula eficientemente la notificación a múltiples usuarios en UN SOLO query
func (r *Repository) AssignToUsersBulk(ctx context.Context, tx *sql.Tx, companyID, notificationID string, userIDs []string) error {
	// Usamos UNNEST para expandir los arrays enviados desde Go en filas de Postgres en una sola operación
	query := `
        INSERT INTO notification_reads (company_id, notification_id, user_id, is_read)
        SELECT $1, $2, unnested_user_id, FALSE
        FROM UNNEST($3::uuid[]) AS unnested_user_id
        ON CONFLICT (notification_id, user_id) DO NOTHING`

	_, err := tx.ExecContext(ctx, query, companyID, notificationID, pq.Array(userIDs))
	return err
}

func (r *Repository) GetUserNotifications(ctx context.Context, companyID, userID string) ([]Notification, error) {
	query := `
        SELECT n.id, n.company_id, n.project_id, n.title, n.message, n.link_to_ui, n.created_at, nr.is_read
        FROM notifications n
        JOIN notification_reads nr ON n.id = nr.notification_id
        WHERE nr.company_id = $1 AND nr.user_id = $2
        ORDER BY n.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, companyID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Notification = make([]Notification, 0) // Evita retornar un json 'null', retorna '[]' vacío
	for rows.Next() {
		var n Notification
		err := rows.Scan(&n.ID, &n.CompanyID, &n.ProjectID, &n.Title, &n.Message, &n.LinkToUI, &n.CreatedAt, &n.IsRead)
		if err != nil {
			return nil, err
		}
		list = append(list, n)
	}
	return list, nil
}

func (r *Repository) MarkAsRead(ctx context.Context, companyID, notificationID, userID string) error {
	query := `
        UPDATE notification_reads 
        SET is_read = TRUE, read_at = CURRENT_TIMESTAMP 
        WHERE company_id = $1 AND notification_id = $2 AND user_id = $3`

	_, err := r.db.ExecContext(ctx, query, companyID, notificationID, userID)
	return err
}
