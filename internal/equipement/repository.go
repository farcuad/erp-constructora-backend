package equipement

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

func (r *Repository) CreateEquipment(ctx context.Context, e *Equipment) error {
	query := `INSERT INTO equipment (company_id, type_id, name, plate_number, model, brand, ownership_type) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7) 
	          RETURNING id, status, created_at, updated_at`

	var typeID interface{} = nil
	if e.TypeID != "" {
		typeID = e.TypeID
	}

	return r.db.QueryRowContext(ctx, query, e.CompanyID, typeID, e.Name, e.PlateNumber, e.Model, e.Brand, e.OwnershipType).
		Scan(&e.ID, &e.Status, &e.CreatedAt, &e.UpdatedAt)
}

func (r *Repository) GetEquipmentByCompany(ctx context.Context, companyID string) ([]Equipment, error) {
	query := `SELECT id, company_id, COALESCE(type_id::text, ''), name, COALESCE(plate_number, ''), COALESCE(model, ''), COALESCE(brand, ''), status, ownership_type, created_at, updated_at 
	          FROM equipment WHERE company_id = $1`

	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Equipment
	for rows.Next() {
		var e Equipment
		if err := rows.Scan(&e.ID, &e.CompanyID, &e.TypeID, &e.Name, &e.PlateNumber, &e.Model, &e.Brand, &e.Status, &e.OwnershipType, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, nil
}

// AssignToProject registra la asignación de obra y cambia el estado del equipo a 'Assigned'
func (r *Repository) AssignToProject(ctx context.Context, a *EquipmentAssignment) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insertar la asignación
	queryAssign := `INSERT INTO equipment_assignments (equipment_id, project_id, assigned_by, start_date, notes) 
	                 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	err = tx.QueryRowContext(ctx, queryAssign, a.EquipmentID, a.ProjectID, a.AssignedBy, a.StartDate, a.Notes).
		Scan(&a.ID, &a.CreatedAt)
	if err != nil {
		return err
	}

	// 2. Actualizar el estado del equipo
	queryUpdate := `UPDATE equipment SET status = 'Assigned', updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err = tx.ExecContext(ctx, queryUpdate, a.EquipmentID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// CreateMaintenance registra el mantenimiento e incrementa los costos u operaciones de la máquina
func (r *Repository) CreateMaintenance(ctx context.Context, m *MaintenanceRecord) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryMain := `INSERT INTO maintenance_records (equipment_id, maintenance_type, description, cost, maintenance_date, next_due_date) 
	              VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`

	var nextDue interface{} = nil
	if m.NextDueDate != "" {
		nextDue = m.NextDueDate
	}

	err = tx.QueryRowContext(ctx, queryMain, m.EquipmentID, m.MaintenanceType, m.Description, m.Cost, m.MaintenanceDate, nextDue).
		Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return err
	}

	// Cambiar estado a mantenimiento preventivo/correctivo
	queryUpdate := `UPDATE equipment SET status = 'In Maintenance', updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err = tx.ExecContext(ctx, queryUpdate, m.EquipmentID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) GetEquipmentByID(ctx context.Context, id, companyID string) (*Equipment, error) {
	query := `SELECT id, company_id, COALESCE(type_id::text, ''), name, COALESCE(plate_number, ''), COALESCE(model, ''), COALESCE(brand, ''), status, ownership_type, created_at, updated_at 
	          FROM equipment WHERE id = $1 AND company_id = $2`
	var e Equipment
	err := r.db.QueryRowContext(ctx, query, id, companyID).Scan(&e.ID, &e.CompanyID, &e.TypeID, &e.Name, &e.PlateNumber, &e.Model, &e.Brand, &e.Status, &e.OwnershipType, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repository) UpdateEquipment(ctx context.Context, e *Equipment) error {
	var typeID interface{} = nil
	if e.TypeID != "" {
		typeID = e.TypeID
	}
	query := `UPDATE equipment SET type_id = $1, name = $2, plate_number = $3, model = $4, brand = $5, status = $6, ownership_type = $7, updated_at = CURRENT_TIMESTAMP WHERE id = $8 AND company_id = $9`
	_, err := r.db.ExecContext(ctx, query, typeID, e.Name, e.PlateNumber, e.Model, e.Brand, e.Status, e.OwnershipType, e.ID, e.CompanyID)
	return err
}

func (r *Repository) DeleteEquipment(ctx context.Context, id, companyID string) error {
	query := `DELETE FROM equipment WHERE id = $1 AND company_id = $2`
	_, err := r.db.ExecContext(ctx, query, id, companyID)
	return err
}

func (r *Repository) GetEquipmentTypeByID(ctx context.Context, id, companyID string) (*EquipmentType, error) {
	query := `SELECT id, company_id, name, created_at FROM equipment_types WHERE id = $1 AND company_id = $2`
	var et EquipmentType
	err := r.db.QueryRowContext(ctx, query, id, companyID).Scan(&et.ID, &et.CompanyID, &et.Name, &et.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &et, nil
}

func (r *Repository) UpdateEquipmentType(ctx context.Context, et *EquipmentType) error {
	query := `UPDATE equipment_types SET name = $1 WHERE id = $2 AND company_id = $3`
	_, err := r.db.ExecContext(ctx, query, et.Name, et.ID, et.CompanyID)
	return err
}

func (r *Repository) DeleteEquipmentType(ctx context.Context, id, companyID string) error {
	query := `DELETE FROM equipment_types WHERE id = $1 AND company_id = $2`
	_, err := r.db.ExecContext(ctx, query, id, companyID)
	return err
}
