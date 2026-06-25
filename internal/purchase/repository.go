package purchase

import (
	"context"
	"database/sql"
)

type Repository interface {
	CreateOrder(ctx context.Context, companyID string, userID string, po *CreatePurchaseOrderRequest) (*PurchaseOrder, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateOrder(ctx context.Context, companyID string, userID string, po *CreatePurchaseOrderRequest) (*PurchaseOrder, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Calcular total acumulado de la orden
	var totalOrder float64
	for _, item := range po.Items {
		totalOrder += item.Quantity * item.UnitPrice
	}

	// 2. Insertar cabecera de la Orden de Compra
	orderQuery := `
		INSERT INTO purchase_orders (company_id, project_id, supplier_id, user_id, total_amount, delivery_date, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, order_number, status, created_at, updated_at`

	var order PurchaseOrder
	order.CompanyID = companyID
	order.ProjectID = po.ProjectID
	order.SupplierID = po.SupplierID
	order.UserID = userID
	order.TotalAmount = totalOrder
	order.DeliveryDate = po.DeliveryDate
	order.Notes = po.Notes

	err = tx.QueryRowContext(ctx, orderQuery, companyID, po.ProjectID, po.SupplierID, userID, totalOrder, po.DeliveryDate, po.Notes).Scan(
		&order.ID, &order.OrderNumber, &order.Status, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// 3. Insertar ítems asociados
	itemQuery := `
		INSERT INTO purchase_order_items (purchase_order_id, description, unit, quantity, unit_price)
		VALUES ($1, $2, $3, $4, $5)`

	for _, item := range po.Items {
		_, err = tx.ExecContext(ctx, itemQuery, order.ID, item.Description, item.Unit, item.Quantity, item.UnitPrice)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &order, nil
}
