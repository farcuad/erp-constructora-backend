-- 1. Tabla principal de Presupuestos
CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title VARCHAR(150) NOT NULL,
    description TEXT,
    total_amount NUMERIC(15, 2) DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_project_budget UNIQUE (project_id)
);

-- 2. Tabla para el historial de Versiones del Presupuesto
CREATE TABLE budget_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    version_number INT NOT NULL, -- Ej: 1, 2, 3...
    status VARCHAR(50) DEFAULT 'Draft', -- Draft, Sent, Approved, Rejected
    total_amount NUMERIC(15, 2) DEFAULT 0.00,
    changed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    description TEXT, -- Motivo del cambio o la revisión
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_budget_version UNIQUE (budget_id, version_number)
);

-- 3. Tabla para los Ítems/Rubros del Presupuesto (Asociados a una versión específica)
CREATE TABLE budget_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_version_id UUID NOT NULL REFERENCES budget_versions(id) ON DELETE CASCADE,
    category VARCHAR(100) NOT NULL, -- Ej: "Cimentación", "Estructura", "Acabados"
    description TEXT NOT NULL,      -- Ej: "Concreto de 3000 PSI para zapatas"
    unit VARCHAR(20) NOT NULL,      -- Ej: "m3", "m2", "kg", "Global"
    quantity NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    unit_price NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    total_price NUMERIC(15, 2) GENERATED ALWAYS AS (quantity * unit_price) STORED,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);s