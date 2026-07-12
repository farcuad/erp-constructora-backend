package inventory

import (
	"context"
	"database/sql"
	"errors"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateWarehouse(ctx context.Context, w *Warehouse) error {
	query := `INSERT INTO warehouses (company_id, project_id, name, location) 
	          VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, w.CompanyID, w.ProjectID, w.Name, w.Location).Scan(&w.ID, &w.CreatedAt)
}

func (r *Repository) CreateMaterial(ctx context.Context, m *Material) error {
	query := `INSERT INTO materials (company_id, category_id, name, code, unit) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`

	var categoryID interface{} = nil
	if m.CategoryID != "" {
		categoryID = m.CategoryID
	}

	return r.db.QueryRowContext(ctx, query, m.CompanyID, categoryID, m.Name, m.Code, m.Unit).
		Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
}

// ExecuteStockMovement procesa la entrada/salida y modifica el inventario real en una transacción
func (r *Repository) ExecuteStockMovement(ctx context.Context, m *StockMovement) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Validar existencias si es una salida (OUTPUT)
	if m.MovementType == "OUTPUT" {
		var currentStock float64
		err := tx.QueryRowContext(ctx,
			"SELECT quantity FROM warehouse_stock WHERE warehouse_id = $1 AND material_id = $2",
			m.WarehouseID, m.MaterialID).Scan(&currentStock)

		if err == sql.ErrNoRows || currentStock < m.Quantity {
			return errors.New("inventario insuficiente para realizar la salida de material")
		}
	}

	// 2. Registrar el movimiento en el historial
	queryMovement := `INSERT INTO stock_movements (warehouse_id, material_id, user_id, movement_type, quantity, reference_id, description) 
	                  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at`

	var refID interface{} = nil
	if m.ReferenceID != "" {
		refID = m.ReferenceID
	}

	err = tx.QueryRowContext(ctx, queryMovement, m.WarehouseID, m.MaterialID, m.UserID, m.MovementType, m.Quantity, refID, m.Description).
		Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return err
	}

	// 3. Modificar el stock actual (Upsert con ON CONFLICT)
	var stockQuery string
	if m.MovementType == "INPUT" {
		stockQuery = `INSERT INTO warehouse_stock (warehouse_id, material_id, quantity) 
		              VALUES ($1, $2, $3) 
		              ON CONFLICT (warehouse_id, material_id) 
		              DO UPDATE SET quantity = warehouse_stock.quantity + EXCLUDED.quantity, updated_at = CURRENT_TIMESTAMP`
	} else {
		stockQuery = `UPDATE warehouse_stock 
		              SET quantity = quantity - $3, updated_at = CURRENT_TIMESTAMP 
		              WHERE warehouse_id = $1 AND material_id = $2`
	}

	_, err = tx.ExecContext(ctx, stockQuery, m.WarehouseID, m.MaterialID, m.Quantity)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) GetStockByWarehouse(ctx context.Context, warehouseID string) ([]MaterialStock, error) {
	query := `SELECT ws.material_id, m.name, COALESCE(m.code, ''), m.unit, ws.quantity 
	          FROM warehouse_stock ws
	          JOIN materials m ON ws.material_id = m.id
	          WHERE ws.warehouse_id = $1 AND ws.quantity > 0`

	rows, err := r.db.QueryContext(ctx, query, warehouseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []MaterialStock
	for rows.Next() {
		var s MaterialStock
		if err := rows.Scan(&s.MaterialID, &s.MaterialName, &s.Code, &s.Unit, &s.Quantity); err != nil {
			return nil, err
		}
		stocks = append(stocks, s)
	}
	return stocks, nil
}

func (r *Repository) GetMaterials(ctx context.Context, companyID string) (*Material, error) {
	query := `SELECT id, company_id, COALESCE(category_id::text, ''), name, code, unit, created_at, updated_at 
	          FROM materials WHERE company_id = $1`
	var m Material
	err := r.db.QueryRowContext(ctx, query, companyID).Scan(&m.ID, &m.CompanyID, &m.CategoryID, &m.Name, &m.Code, &m.Unit, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetMaterialByID(ctx context.Context, id, companyID string) (*Material, error) {
	query := `SELECT id, company_id, COALESCE(category_id::text, ''), name, COALESCE(code, ''), unit, created_at, updated_at 
	          FROM materials WHERE id = $1 AND company_id = $2`
	var m Material
	err := r.db.QueryRowContext(ctx, query, id, companyID).Scan(&m.ID, &m.CompanyID, &m.CategoryID, &m.Name, &m.Code, &m.Unit, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) UpdateMaterial(ctx context.Context, m *Material) error {
	var categoryID interface{} = nil
	if m.CategoryID != "" {
		categoryID = m.CategoryID
	}
	query := `UPDATE materials SET name = $1, code = $2, unit = $3, category_id = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $5 AND company_id = $6`
	_, err := r.db.ExecContext(ctx, query, m.Name, m.Code, m.Unit, categoryID, m.ID, m.CompanyID)
	return err
}

func (r *Repository) DeleteMaterial(ctx context.Context, id, companyID string) error {
	query := `DELETE FROM materials WHERE id = $1 AND company_id = $2`
	_, err := r.db.ExecContext(ctx, query, id, companyID)
	return err
}

func (r *Repository) GetWarehouseByID(ctx context.Context, id, companyID string) (*Warehouse, error) {
	query := `SELECT id, company_id, project_id, name, COALESCE(location, ''), created_at 
	          FROM warehouses WHERE id = $1 AND company_id = $2`
	var w Warehouse
	err := r.db.QueryRowContext(ctx, query, id, companyID).Scan(&w.ID, &w.CompanyID, &w.ProjectID, &w.Name, &w.Location, &w.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *Repository) UpdateWarehouse(ctx context.Context, w *Warehouse) error {
	query := `UPDATE warehouses SET name = $1, location = $2 WHERE id = $3 AND company_id = $4`
	_, err := r.db.ExecContext(ctx, query, w.Name, w.Location, w.ID, w.CompanyID)
	return err
}

func (r *Repository) DeleteWarehouse(ctx context.Context, id, companyID string) error {
	query := `DELETE FROM warehouses WHERE id = $1 AND company_id = $2`
	_, err := r.db.ExecContext(ctx, query, id, companyID)
	return err
}
