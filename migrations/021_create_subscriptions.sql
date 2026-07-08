CREATE TABLE companies_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'trial',
    start_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_date TIMESTAMPTZ,
    trial_end_date TIMESTAMPTZ,
    price NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    billing_cycle VARCHAR(20) NOT NULL DEFAULT 'monthly',
    max_projects INT NOT NULL DEFAULT 1,
    max_users INT NOT NULL DEFAULT 3,
    max_storage_mb INT NOT NULL DEFAULT 100,
    features JSONB DEFAULT '{}',
    payment_provider VARCHAR(50),
    payment_provider_subscription_id VARCHAR(255),
    payment_provider_customer_id VARCHAR(255),
    last_payment_date TIMESTAMPTZ,
    next_billing_date TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT unique_company_subscription UNIQUE (company_id)
);

CREATE INDEX idx_subscriptions_status ON companies_subscriptions(status);
CREATE INDEX idx_subscriptions_company ON companies_subscriptions(company_id);
