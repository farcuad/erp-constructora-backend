package payments

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

func (r *Repository) CreateInvoice(ctx context.Context, inv *Invoice) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryInvoice := `
		INSERT INTO invoices (company_id, project_id, invoice_number, type, status, client_id, supplier_id, contractor_id, issue_date, due_date, subtotal, tax_amount, total_amount, remaining_amount, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at`

	err = tx.QueryRowContext(ctx, queryInvoice,
		inv.CompanyID, inv.ProjectID, inv.InvoiceNumber, inv.Type, inv.Status,
		inv.ClientID, inv.SupplierID, inv.ContractorID, inv.IssueDate, inv.DueDate,
		inv.Subtotal, inv.TaxAmount, inv.TotalAmount, inv.TotalAmount, inv.Notes, // remaining_amount inicia igual al total
	).Scan(&inv.ID, &inv.CreatedAt)
	if err != nil {
		return err
	}

	queryItem := `
		INSERT INTO invoice_items (company_id, invoice_id, description, quantity, unit_price, total)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	for i := range inv.Items {
		inv.Items[i].CompanyID = inv.CompanyID
		inv.Items[i].InvoiceID = inv.ID
		inv.Items[i].Total = inv.Items[i].Quantity * inv.Items[i].UnitPrice

		err = tx.QueryRowContext(ctx, queryItem,
			inv.Items[i].CompanyID, inv.Items[i].InvoiceID, inv.Items[i].Description,
			inv.Items[i].Quantity, inv.Items[i].UnitPrice, inv.Items[i].Total,
		).Scan(&inv.Items[i].ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) UpdateInvoice(ctx context.Context, companyID, id string, req UpdateInvoiceRequest) error {
	query := `
		UPDATE invoices
		SET status = COALESCE($1, status),
		    notes = COALESCE($2, notes),
		    due_date = COALESCE($3, due_date)
		WHERE company_id = $4 AND id = $5`

	var status, notes interface{}
	if req.Status != nil {
		status = *req.Status
	} else {
		status = nil
	}
	if req.Notes != nil {
		notes = *req.Notes
	} else {
		notes = nil
	}

	_, err := r.db.ExecContext(ctx, query, status, notes, req.DueDate, companyID, id)
	return err
}

func (r *Repository) DeleteInvoice(ctx context.Context, companyID, id string) error {
	query := `DELETE FROM invoices WHERE company_id = $1 AND id = $2`
	_, err := r.db.ExecContext(ctx, query, companyID, id)
	return err
}

func (r *Repository) CancelInvoice(ctx context.Context, companyID, id string) error {
	query := `UPDATE invoices SET status = 'CANCELLED' WHERE company_id = $1 AND id = $2`
	_, err := r.db.ExecContext(ctx, query, companyID, id)
	return err
}

func (r *Repository) RegisterPayment(ctx context.Context, p *Payment) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insertar pago
	queryPayment := `
		INSERT INTO payments (company_id, project_id, invoice_id, payment_date, amount, payment_method, reference, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`

	err = tx.QueryRowContext(ctx, queryPayment,
		p.CompanyID, p.ProjectID, p.InvoiceID, p.PaymentDate, p.Amount, p.PaymentMethod, p.Reference, p.Notes,
	).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return err
	}

	// 2. Actualizar el saldo restante (remaining_amount) y estatus de la factura
	queryUpdateInvoice := `
		UPDATE invoices 
		SET remaining_amount = remaining_amount - $1,
			status = CASE 
				WHEN (remaining_amount - $1) <= 0 THEN 'PAID'::invoice_status 
				ELSE 'PARTIALLY_PAID'::invoice_status 
			END
		WHERE id = $2 AND company_id = $3`

	_, err = tx.ExecContext(ctx, queryUpdateInvoice, p.Amount, p.InvoiceID, p.CompanyID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
