package purchase

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

// --- MÉTODOS DE ÓRDENES DE COMPRA ---

func (r *Repository) CreatePurchaseOrder(ctx context.Context, po *PurchaseOrder) error {
	// Iniciamos transacción para insertar la cabecera y los ítems de manera segura
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insertar la Orden de Compra Principal
	queryOrder := `INSERT INTO purchase_orders (company_id, project_id, supplier_id, user_id, status, total_amount, delivery_date, notes) 
	               VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
	               RETURNING id, order_number, created_at, updated_at`

	var deliveryDate interface{} = nil
	if po.DeliveryDate != "" {
		deliveryDate = po.DeliveryDate
	}

	err = tx.QueryRowContext(ctx, queryOrder, po.CompanyID, po.ProjectID, po.SupplierID, po.UserID, po.Status, po.TotalAmount, deliveryDate, po.Notes).
		Scan(&po.ID, &po.OrderNumber, &po.CreatedAt, &po.UpdatedAt)
	if err != nil {
		return err
	}

	// 2. Insertar los ítems asociados utilizando la ID devuelta de la cabecera
	queryItem := `INSERT INTO purchase_order_items (purchase_order_id, description, unit, quantity, unit_price) 
	              VALUES ($1, $2, $3, $4, $5) 
	              RETURNING id, total_price`

	for i := range po.Items {
		po.Items[i].PurchaseOrderID = po.ID
		err = tx.QueryRowContext(ctx, queryItem, po.ID, po.Items[i].Description, po.Items[i].Unit, po.Items[i].Quantity, po.Items[i].UnitPrice).
			Scan(&po.Items[i].ID, &po.Items[i].TotalPrice)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) GetOrdersByProject(ctx context.Context, projectID string) ([]PurchaseOrder, error) {
	query := `SELECT id, company_id, project_id, supplier_id, user_id, order_number, status, total_amount, COALESCE(delivery_date::text, ''), notes, created_at, updated_at 
	          FROM purchase_orders WHERE project_id = $1`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []PurchaseOrder
	for rows.Next() {
		var po PurchaseOrder
		if err := rows.Scan(&po.ID, &po.CompanyID, &po.ProjectID, &po.SupplierID, &po.UserID, &po.OrderNumber, &po.Status, &po.TotalAmount, &po.DeliveryDate, &po.Notes, &po.CreatedAt, &po.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, po)
	}
	return orders, nil
}
