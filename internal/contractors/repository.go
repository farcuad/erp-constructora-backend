package contractors

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

func (r *Repository) CreateContractor(ctx context.Context, c *Contractor) error {
	query := `INSERT INTO contractors (company_id, name, nit, representative, phone, email) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, is_active, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, c.CompanyID, c.Name, c.NIT, c.Representative, c.Phone, c.Email).
		Scan(&c.ID, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
}

func (r *Repository) CreateContract(ctx context.Context, cc *ContractorContract) error {
	query := `INSERT INTO contractor_contracts (company_id, contractor_id, project_id, title, total_amount, balance, start_date, end_date) 
	          VALUES ($1, $2, $3, $4, $5, $5, $6, $7) -- Inicializamos balance igual al total_amount
	          RETURNING id, balance, status, created_at, updated_at`

	var endDate interface{} = nil
	if cc.EndDate != "" {
		endDate = cc.EndDate
	}

	return r.db.QueryRowContext(ctx, query, cc.CompanyID, cc.ContractorID, cc.ProjectID, cc.Title, cc.TotalAmount, cc.StartDate, endDate).
		Scan(&cc.ID, &cc.Balance, &cc.Status, &cc.CreatedAt, &cc.UpdatedAt)
}

// Procesa el pago y deduce de forma atómica el saldo remanente del subcontrato
func (r *Repository) RegisterPayment(ctx context.Context, p *ContractorPayment) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Validar que el saldo del contrato cubra el pago
	var currentBalance float64
	err = tx.QueryRowContext(ctx, "SELECT balance FROM contractor_contracts WHERE id = $1 For UPDATE", p.ContractID).Scan(&currentBalance)
	if err != nil {
		return err
	}
	if currentBalance < p.Amount {
		return errors.New("el monto del pago excede el saldo restante del subcontrato")
	}

	// 2. Registrar el pago
	queryPay := `INSERT INTO contractor_payments (contract_id, user_id, amount, payment_date, reference_number, notes) 
	             VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	err = tx.QueryRowContext(ctx, queryPay, p.ContractID, p.UserID, p.Amount, p.PaymentDate, p.ReferenceNumber, p.Notes).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return err
	}

	// 3. Descontar del balance del contrato principal
	queryUpdate := `UPDATE contractor_contracts SET balance = balance - $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err = tx.ExecContext(ctx, queryUpdate, p.ContractID, p.Amount)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) GetContractsByProject(ctx context.Context, projectID string) ([]ContractorContract, error) {
	query := `SELECT id, company_id, contractor_id, project_id, title, total_amount, balance, start_date, COALESCE(end_date::text, ''), status, created_at, updated_at 
	          FROM contractor_contracts WHERE project_id = $1`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ContractorContract
	for rows.Next() {
		var cc ContractorContract
		if err := rows.Scan(&cc.ID, &cc.CompanyID, &cc.ContractorID, &cc.ProjectID, &cc.Title, &cc.TotalAmount, &cc.Balance, &cc.StartDate, &cc.EndDate, &cc.Status, &cc.CreatedAt, &cc.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, cc)
	}
	return list, nil
}
