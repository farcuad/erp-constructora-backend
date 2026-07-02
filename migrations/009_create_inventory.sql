-- 1. Categorías de Materiales
CREATE TABLE material_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_company_category UNIQUE (company_id, name)
);

-- 2. Catálogo de Materiales General
CREATE TABLE materials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    category_id UUID REFERENCES material_categories(id) ON DELETE SET NULL,
    name VARCHAR(150) NOT NULL,
    code VARCHAR(50), -- Código interno de inventario (SKU)
    unit VARCHAR(20) NOT NULL, -- Ej: Bolsa, Metro, Tonelada
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_company_material_code UNIQUE (company_id, code)
);

-- 3. Almacenes / Bodegas (Vinculados a un Proyecto)
CREATE TABLE warehouses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL, -- Ej: "Bodega Principal - Torre A"
    location TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_project_warehouse UNIQUE (project_id, name)
);

-- 4. Stock Actual por Almacén (Tabla de relación física de existencias)
CREATE TABLE warehouse_stock (
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    material_id UUID NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
    quantity NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (warehouse_id, material_id)
);

-- 5. Historial de Movimientos de Inventario
CREATE TABLE stock_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    material_id UUID NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    movement_type VARCHAR(20) NOT NULL, -- 'INPUT' (Entrada), 'OUTPUT' (Salida)
    quantity NUMERIC(12, 2) NOT NULL,
    reference_id UUID, -- ID de la Orden de Compra (si es INPUT) o Reporte Diario (si es OUTPUT)
    description TEXT, -- Ej: "Recepción de Orden de Compra #104"
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);