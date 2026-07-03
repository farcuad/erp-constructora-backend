-- 1. Tabla de Contratistas (Subcontratistas)
CREATE TABLE contractors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,       -- Razón social o nombre del contratista
    nit VARCHAR(50) NOT NULL,        -- Identificación fiscal
    representative VARCHAR(100),      -- Representante legal / Contacto
    phone VARCHAR(20),
    email VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_company_contractor_nit UNIQUE (company_id, nit)
);

-- 2. Contratos con Contratistas (Vinculados a un Proyecto/Obra)
CREATE TABLE contractor_contracts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    contractor_id UUID NOT NULL REFERENCES contractors(id) ON DELETE RESTRICT,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title VARCHAR(150) NOT NULL,       -- Ej: "Subcontrato de Redes Eléctricas e Iluminación"
    total_amount NUMERIC(15, 2) NOT NULL DEFAULT 0.00, -- Valor total acordado
    balance NUMERIC(15, 2) NOT NULL DEFAULT 0.00,      -- Saldo pendiente por pagar
    start_date DATE NOT NULL,
    end_date DATE,
    status VARCHAR(50) DEFAULT 'Active', -- 'Active', 'Completed', 'Suspended'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Pagos / Valuaciones de Contratistas
CREATE TABLE contractor_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id UUID NOT NULL REFERENCES contractor_contracts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id), -- Quién aprueba/registra el pago
    amount NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    payment_date DATE NOT NULL,
    reference_number VARCHAR(50),      -- Número de transferencia, cheque o recibo
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);