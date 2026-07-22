package subscriptions

import "time"

type CompanySubscription struct {
	ID                            string     `json:"id"`
	CompanyID                     string     `json:"company_id"`
	Status                        string     `json:"status"`
	StartDate                     time.Time  `json:"start_date"`
	EndDate                       *time.Time `json:"end_date,omitempty"`
	TrialEndDate                  *time.Time `json:"trial_end_date,omitempty"`
	Price                         float64    `json:"price"`
	BillingCycle                  string     `json:"billing_cycle"`
	MaxProjects                   int        `json:"max_projects"`
	MaxUsers                      int        `json:"max_users"`
	MaxStorageMB                  int        `json:"max_storage_mb"`
	Features                      string     `json:"features"`
	PaymentProvider               string     `json:"payment_provider,omitempty"`
	PaymentProviderSubscriptionID string     `json:"payment_provider_subscription_id,omitempty"`
	PaymentProviderCustomerID     string     `json:"payment_provider_customer_id,omitempty"`
	LastPaymentDate               *time.Time `json:"last_payment_date,omitempty"`
	NextBillingDate               *time.Time `json:"next_billing_date,omitempty"`
	CancelledAt                   *time.Time `json:"cancelled_at,omitempty"`
	CreatedAt                     time.Time  `json:"created_at"`
	UpdatedAt                     time.Time  `json:"updated_at"`
}

type SubscriptionWithCompany struct {
	CompanySubscription
	CompanyName  string `json:"company_name"`
	CompanyNIT   string `json:"company_nit"`
	CompanyEmail string `json:"company_email,omitempty"`
	UserCount    int    `json:"user_count"`
	ProjectCount int    `json:"project_count"`
}

type PaymentRecord struct {
	ID            string    `json:"id"`
	InvoiceID     string    `json:"invoice_id"`
	InvoiceNumber string    `json:"invoice_number"`
	PaymentDate   time.Time `json:"payment_date"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	Reference     string    `json:"reference,omitempty"`
}

type CreateSubscriptionRequest struct {
	Status       string  `json:"status"`
	Price        float64 `json:"price"`
	BillingCycle string  `json:"billing_cycle"`
	MaxProjects  int     `json:"max_projects"`
	MaxUsers     int     `json:"max_users"`
	MaxStorageMB int     `json:"max_storage_mb"`
}

type UpdateSubscriptionRequest struct {
	Status       *string  `json:"status,omitempty"`
	Price        *float64 `json:"price,omitempty"`
	BillingCycle *string  `json:"billing_cycle,omitempty"`
	MaxProjects  *int     `json:"max_projects,omitempty"`
	MaxUsers     *int     `json:"max_users,omitempty"`
	MaxStorageMB *int     `json:"max_storage_mb,omitempty"`
	EndDate      *string  `json:"end_date,omitempty"`
}
