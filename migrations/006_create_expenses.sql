-- 1. Tabla de Categorías de Gastos
CREATE TABLE expense_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL -- Materiales, Personal, Transporte, Maquinaria, Imprevistos
);

-- Inserción inicial de las categorías base especificadas en tu flujo
INSERT INTO expense_categories (name) VALUES 
('Materiales'),
('Personal'),
('Transporte'),
('Maquinaria'),
('Imprevistos');

-- 2. Tabla Principal de Gastos (Expenses)
CREATE TABLE expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    category_id INT NOT NULL REFERENCES expense_categories(id),
    user_id UUID NOT NULL REFERENCES users(id), -- Usuario que registra el gasto (Auditoría)
    title VARCHAR(150) NOT NULL,                -- Descripción corta del gasto
    amount NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    expense_date DATE NOT NULL DEFAULT CURRENT_DATE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Tabla para Documentos de Soporte (Facturas de proveedor, recibos, etc.)
CREATE TABLE expense_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_id UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL, -- URL del archivo subido en el Storage de Supabase
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);