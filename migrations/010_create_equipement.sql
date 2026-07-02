-- 1. Tipos / Categorías de Maquinaria
CREATE TABLE equipment_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL, -- Ej: "Excavadora", "Generador Eléctrico"
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_company_equipment_type UNIQUE (company_id, name)
);

-- 2. Inventario de Maquinaria y Equipos
CREATE TABLE equipment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    type_id UUID REFERENCES equipment_types(id) ON DELETE SET NULL,
    name VARCHAR(150) NOT NULL,      -- Ej: "Caterpillar 320D"
    plate_number VARCHAR(50),       -- Placa o serial de identificación interna
    model VARCHAR(100),
    brand VARCHAR(100),
    status VARCHAR(50) DEFAULT 'Available', -- 'Available', 'Assigned', 'In Maintenance', 'Out of Service'
    ownership_type VARCHAR(20) DEFAULT 'Owned', -- 'Owned' (Propia), 'Rented' (Alquilada)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_company_equipment_plate UNIQUE (company_id, plate_number)
);

-- 3. Asignaciones de Maquinaria a Proyectos / Obras
CREATE TABLE equipment_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    equipment_id UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    assigned_by UUID NOT NULL REFERENCES users(id),
    start_date DATE NOT NULL,
    end_date DATE, -- Puede ser NULL si sigue asignada indefinidamente
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. Registro de Mantenimientos
CREATE TABLE maintenance_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    equipment_id UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    maintenance_type VARCHAR(50) NOT NULL, -- 'Preventive', 'Corrective'
    description TEXT NOT NULL,             -- Detalle del trabajo realizado
    cost NUMERIC(15, 2) DEFAULT 0.00,       -- Costo del mantenimiento (alimenta el control de costos)
    maintenance_date DATE NOT NULL,
    next_due_date DATE,                    -- Próximo mantenimiento preventivo programado
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);