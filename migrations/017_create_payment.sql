-- Módulo 16: Facturación
CREATE TYPE invoice_type AS ENUM ('EMITTED', 'RECEIVED');
CREATE TYPE invoice_status AS ENUM ('DRAFT', 'PENDING', 'PAID', 'PARTIALLY_PAID', 'CANCELLED');

CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    invoice_number VARCHAR(50) NOT NULL, -- Número de factura legal
    type invoice_type NOT NULL,          -- EMITTED (Cliente) o RECEIVED (Proveedor/Contratista)
    status invoice_status NOT NULL DEFAULT 'PENDING',
    
    -- Relaciones opcionales dependiendo del tipo de factura
    client_id UUID REFERENCES clients(id) ON DELETE SET NULL,         -- Si es EMITTED
    supplier_id UUID REFERENCES suppliers(id) ON DELETE SET NULL,     -- Si es RECEIVED (Proveedor)
    contractor_id UUID REFERENCES contractors(id) ON DELETE SET NULL, -- Si es RECEIVED (Contratista)
    
    issue_date DATE NOT NULL,
    due_date DATE NOT NULL,
    subtotal NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    tax_amount NUMERIC(15, 2) NOT NULL DEFAULT 0.00, -- IVA / Impuestos
    total_amount NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    remaining_amount NUMERIC(15, 2) NOT NULL DEFAULT 0.00, -- Saldo pendiente por pagar
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- El número de factura debe ser único por empresa/tipo (ej. no repetir mismo número de proveedor)
    CONSTRAINT unique_invoice_per_company UNIQUE (company_id, invoice_number, type)
);

CREATE TABLE invoice_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    description TEXT NOT NULL, -- Ej: "Valuación N° 3 - Vaciado de losas"
    quantity NUMERIC(12, 2) NOT NULL DEFAULT 1.00,
    unit_price NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    total NUMERIC(15, 2) NOT NULL DEFAULT 0.00
);

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    payment_date DATE NOT NULL,
    amount NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
    payment_method VARCHAR(50) NOT NULL, -- Ej: Transferencia, Cheque, Efectivo
    reference VARCHAR(100),              -- Número de transferencia o boucher
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Índices optimizados
CREATE INDEX idx_invoices_lookup ON invoices(company_id, project_id, type, status);
CREATE INDEX idx_invoice_items ON invoice_items(invoice_id);
CREATE INDEX idx_payments_invoice ON payments(invoice_id);


