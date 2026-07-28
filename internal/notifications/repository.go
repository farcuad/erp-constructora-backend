package notifications

import (
	"context"
	"database/sql"
	"encoding/json"

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
		INSERT INTO notifications (
			company_id, project_id, entity_type, entity_id, 
			type, priority, title, message, link_to_ui, metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at`

	// Si Metadata viene nulo o vacío, asignamos un JSON objeto por defecto
	metaBytes := n.Metadata
	if len(metaBytes) == 0 {
		metaBytes = json.RawMessage("{}")
	}

	return tx.QueryRowContext(
		ctx, query,
		n.CompanyID, n.ProjectID, n.EntityType, n.EntityID,
		n.Type, n.Priority, n.Title, n.Message, n.LinkToUI, metaBytes,
	).Scan(&n.ID, &n.CreatedAt)
}

// AssignToUsersBulk vincula eficientemente la notificación a múltiples usuarios en UN SOLO query
func (r *Repository) AssignToUsersBulk(ctx context.Context, tx *sql.Tx, companyID, notificationID string, userIDs []string) error {
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
		SELECT 
			n.id, n.company_id, n.project_id, n.entity_type, n.entity_id, 
			n.type, n.priority, n.title, n.message, n.link_to_ui, n.metadata, 
			n.created_at, nr.is_read
		FROM notifications n
		JOIN notification_reads nr ON n.id = nr.notification_id
		WHERE nr.company_id = $1 AND nr.user_id = $2
		ORDER BY n.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, companyID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]Notification, 0)
	for rows.Next() {
		var n Notification
		var metaBytes []byte

		err := rows.Scan(
			&n.ID, &n.CompanyID, &n.ProjectID, &n.EntityType, &n.EntityID,
			&n.Type, &n.Priority, &n.Title, &n.Message, &n.LinkToUI, &metaBytes,
			&n.CreatedAt, &n.IsRead,
		)
		if err != nil {
			return nil, err
		}

		n.Metadata = json.RawMessage(metaBytes)
		list = append(list, n)
	}
	return list, nil
}

func (r *Repository) Delete(ctx context.Context, companyID, id string) error {
	query := `DELETE FROM notifications WHERE company_id = $1 AND id = $2`
	_, err := r.db.ExecContext(ctx, query, companyID, id)
	return err
}

func (r *Repository) MarkAsRead(ctx context.Context, companyID, notificationID, userID string) error {
	query := `
		UPDATE notification_reads 
		SET is_read = TRUE, read_at = CURRENT_TIMESTAMP 
		WHERE company_id = $1 AND notification_id = $2 AND user_id = $3`

	_, err := r.db.ExecContext(ctx, query, companyID, notificationID, userID)
	return err
}
