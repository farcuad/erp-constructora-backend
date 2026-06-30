
-- 1. Tabla Principal de Órdenes de Compra
CREATE TABLE purchase_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    supplier_id UUID NOT NULL REFERENCES suppliers(id) ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES users(id), -- Quien genera la orden
    order_number SERIAL,                        -- Consecutivo automático interno
    status VARCHAR(50) DEFAULT 'Pending',       -- Pending, Approved, Received, Cancelled
    total_amount NUMERIC(15, 2) DEFAULT 0.00,
    delivery_date DATE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Tabla de Detalles de la Orden de Compra (Muchos a Muchos / Items)
CREATE TABLE purchase_order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_order_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    description TEXT NOT NULL,       -- Ej: Varilla corrugada de 1/2 pulgada
    unit VARCHAR(20) NOT NULL,       -- Ej: Unidad, Tonelada, Global
    quantity NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    unit_price NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    total_price NUMERIC(15, 2) GENERATED ALWAYS AS (quantity * unit_price) STORED,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);