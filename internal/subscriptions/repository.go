package subscriptions

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

func (r *Repository) Create(ctx context.Context, companyID string, req *CreateSubscriptionRequest) (*CompanySubscription, error) {
	query := `
		INSERT INTO companies_subscriptions (company_id, status, price, billing_cycle, max_projects, max_users, max_storage_mb)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, start_date, trial_end_date, created_at, updated_at`

	var s CompanySubscription
	trialEnd := time.Now().AddDate(0, 0, 14)

	err := r.db.QueryRowContext(ctx, query,
		companyID, req.Status, req.Price, req.BillingCycle,
		req.MaxProjects, req.MaxUsers, req.MaxStorageMB,
	).Scan(&s.ID, &s.StartDate, &s.TrialEndDate, &s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		return nil, err
	}

	s.TrialEndDate = &trialEnd
	s.CompanyID = companyID
	s.Status = req.Status
	s.Price = req.Price
	s.BillingCycle = req.BillingCycle
	s.MaxProjects = req.MaxProjects
	s.MaxUsers = req.MaxUsers
	s.MaxStorageMB = req.MaxStorageMB

	return &s, nil
}

func (r *Repository) GetByCompany(ctx context.Context, companyID string) (*CompanySubscription, error) {
	query := `
		SELECT id, company_id, status, start_date, end_date, trial_end_date, price,
		       billing_cycle, max_projects, max_users, max_storage_mb, features,
		       COALESCE(payment_provider, ''), COALESCE(payment_provider_subscription_id, ''),
		       COALESCE(payment_provider_customer_id, ''),
		       last_payment_date, next_billing_date, cancelled_at, created_at, updated_at
		FROM companies_subscriptions WHERE company_id = $1`

	var s CompanySubscription
	var endDate, trialEnd, lastPayment, nextBilling, cancelled sql.NullTime
	var features sql.NullString

	err := r.db.QueryRowContext(ctx, query, companyID).Scan(
		&s.ID, &s.CompanyID, &s.Status, &s.StartDate, &endDate, &trialEnd,
		&s.Price, &s.BillingCycle, &s.MaxProjects, &s.MaxUsers, &s.MaxStorageMB,
		&features, &s.PaymentProvider, &s.PaymentProviderSubscriptionID,
		&s.PaymentProviderCustomerID, &lastPayment, &nextBilling, &cancelled,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if endDate.Valid { s.EndDate = &endDate.Time }
	if trialEnd.Valid { s.TrialEndDate = &trialEnd.Time }
	if lastPayment.Valid { s.LastPaymentDate = &lastPayment.Time }
	if nextBilling.Valid { s.NextBillingDate = &nextBilling.Time }
	if cancelled.Valid { s.CancelledAt = &cancelled.Time }
	if features.Valid { s.Features = features.String }

	return &s, nil
}

func (r *Repository) Update(ctx context.Context, id string, companyID string, req *UpdateSubscriptionRequest) (*CompanySubscription, error) {
	query := `
		UPDATE companies_subscriptions SET updated_at = CURRENT_TIMESTAMP`

	var args []interface{}
	argIdx := 1

	if req.Status != nil {
		query += `, status = $` + string(rune('0'+argIdx))
		args = append(args, *req.Status)
		argIdx++
	}
	if req.Price != nil {
		query += `, price = $` + string(rune('0'+argIdx))
		args = append(args, *req.Price)
		argIdx++
	}
	if req.BillingCycle != nil {
		query += `, billing_cycle = $` + string(rune('0'+argIdx))
		args = append(args, *req.BillingCycle)
		argIdx++
	}
	if req.MaxProjects != nil {
		query += `, max_projects = $` + string(rune('0'+argIdx))
		args = append(args, *req.MaxProjects)
		argIdx++
	}
	if req.MaxUsers != nil {
		query += `, max_users = $` + string(rune('0'+argIdx))
		args = append(args, *req.MaxUsers)
		argIdx++
	}
	if req.MaxStorageMB != nil {
		query += `, max_storage_mb = $` + string(rune('0'+argIdx))
		args = append(args, *req.MaxStorageMB)
		argIdx++
	}
	if req.EndDate != nil {
		query += `, end_date = $` + string(rune('0'+argIdx))
		args = append(args, *req.EndDate)
		argIdx++
	}

	query += ` WHERE id = $` + string(rune('0'+argIdx)) + ` AND company_id = $` + string(rune('0'+argIdx+1))
	args = append(args, id, companyID)
	query += ` RETURNING updated_at`

	var updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&updatedAt)
	if err != nil {
		return nil, err
	}

	return r.GetByCompany(ctx, companyID)
}

func (r *Repository) CountActiveProjects(ctx context.Context, companyID string) (int, error) {
	query := `SELECT COUNT(*) FROM projects WHERE company_id = $1`
	var count int
	err := r.db.QueryRowContext(ctx, query, companyID).Scan(&count)
	return count, err
}

func (r *Repository) CountActiveUsers(ctx context.Context, companyID string) (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE company_id = $1 AND is_active = true`
	var count int
	err := r.db.QueryRowContext(ctx, query, companyID).Scan(&count)
	return count, err
}

func (r *Repository) GetAllWithCompany(ctx context.Context) ([]SubscriptionWithCompany, error) {
	query := `
		SELECT cs.id, cs.company_id, cs.status, cs.start_date, cs.end_date, cs.trial_end_date,
		       cs.price, cs.billing_cycle, cs.max_projects, cs.max_users, cs.max_storage_mb, cs.features,
		       COALESCE(cs.payment_provider, ''), COALESCE(cs.payment_provider_subscription_id, ''),
		       COALESCE(cs.payment_provider_customer_id, ''),
		       cs.last_payment_date, cs.next_billing_date, cs.cancelled_at, cs.created_at, cs.updated_at,
		       c.name, c.nit, COALESCE(c.email, '')
		FROM companies_subscriptions cs
		JOIN companies c ON c.id = cs.company_id
		ORDER BY cs.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []SubscriptionWithCompany
	for rows.Next() {
		var s SubscriptionWithCompany
		var endDate, trialEnd, lastPayment, nextBilling, cancelled sql.NullTime
		var features sql.NullString

		err := rows.Scan(
			&s.ID, &s.CompanyID, &s.Status, &s.StartDate, &endDate, &trialEnd,
			&s.Price, &s.BillingCycle, &s.MaxProjects, &s.MaxUsers, &s.MaxStorageMB,
			&features, &s.PaymentProvider, &s.PaymentProviderSubscriptionID,
			&s.PaymentProviderCustomerID, &lastPayment, &nextBilling, &cancelled,
			&s.CreatedAt, &s.UpdatedAt,
			&s.CompanyName, &s.CompanyNIT, &s.CompanyEmail,
		)
		if err != nil {
			return nil, err
		}

		if endDate.Valid { s.EndDate = &endDate.Time }
		if trialEnd.Valid { s.TrialEndDate = &trialEnd.Time }
		if lastPayment.Valid { s.LastPaymentDate = &lastPayment.Time }
		if nextBilling.Valid { s.NextBillingDate = &nextBilling.Time }
		if cancelled.Valid { s.CancelledAt = &cancelled.Time }
		if features.Valid { s.Features = features.String }

		userCount, _ := r.CountActiveUsers(ctx, s.CompanyID)
		projectCount, _ := r.CountActiveProjects(ctx, s.CompanyID)
		s.UserCount = userCount
		s.ProjectCount = projectCount

		subs = append(subs, s)
	}
	return subs, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*CompanySubscription, error) {
	query := `
		SELECT id, company_id, status, start_date, end_date, trial_end_date, price,
		       billing_cycle, max_projects, max_users, max_storage_mb, features,
		       COALESCE(payment_provider, ''), COALESCE(payment_provider_subscription_id, ''),
		       COALESCE(payment_provider_customer_id, ''),
		       last_payment_date, next_billing_date, cancelled_at, created_at, updated_at
		FROM companies_subscriptions WHERE id = $1`

	var s CompanySubscription
	var endDate, trialEnd, lastPayment, nextBilling, cancelled sql.NullTime
	var features sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.CompanyID, &s.Status, &s.StartDate, &endDate, &trialEnd,
		&s.Price, &s.BillingCycle, &s.MaxProjects, &s.MaxUsers, &s.MaxStorageMB,
		&features, &s.PaymentProvider, &s.PaymentProviderSubscriptionID,
		&s.PaymentProviderCustomerID, &lastPayment, &nextBilling, &cancelled,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if endDate.Valid { s.EndDate = &endDate.Time }
	if trialEnd.Valid { s.TrialEndDate = &trialEnd.Time }
	if lastPayment.Valid { s.LastPaymentDate = &lastPayment.Time }
	if nextBilling.Valid { s.NextBillingDate = &nextBilling.Time }
	if cancelled.Valid { s.CancelledAt = &cancelled.Time }
	if features.Valid { s.Features = features.String }

	return &s, nil
}

func (r *Repository) GetPaymentsByCompany(ctx context.Context, companyID string) ([]PaymentRecord, error) {
	query := `
		SELECT p.id, p.invoice_id, i.invoice_number, p.payment_date, p.amount, p.payment_method, COALESCE(p.reference, '')
		FROM payments p
		JOIN invoices i ON i.id = p.invoice_id
		WHERE p.company_id = $1
		ORDER BY p.payment_date DESC`

	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []PaymentRecord
	for rows.Next() {
		var p PaymentRecord
		if err := rows.Scan(&p.ID, &p.InvoiceID, &p.InvoiceNumber, &p.PaymentDate, &p.Amount, &p.PaymentMethod, &p.Reference); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, nil
}
