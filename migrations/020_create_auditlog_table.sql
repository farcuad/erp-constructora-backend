CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL, -- SET NULL por si el usuario es eliminado, el log persiste
    action VARCHAR(100) NOT NULL, -- Ej: "CREATE_BUDGET", "UPDATE_PURCHASE_ORDER", "DELETE_EMPLOYEE"
    table_name VARCHAR(100) NOT NULL, -- Ej: "budgets", "purchase_orders"
    row_id UUID, -- El ID del registro afectado (opcional pero muy útil)
    ip_address VARCHAR(45) NOT NULL, -- Soporta IPv4 e IPv6
    old_values JSONB, -- Estado anterior del registro (NULL si es creación)[cite: 2]
    new_values JSONB, -- Estado nuevo del registro (NULL si es eliminación)[cite: 2]
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP -- Fecha y hora exacta[cite: 2]
);

-- Índice para búsquedas rápidas por empresa y acción (muy común en ERP)
CREATE INDEX idx_audit_logs_company_action ON audit_logs(company_id, action);