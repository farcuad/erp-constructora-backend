-- Tabla de Reportes Diarios (Bitácora de la obra)
CREATE TABLE daily_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id), -- Ingeniero/Supervisor que reporta
    report_date DATE NOT NULL,
    weather_condition VARCHAR(100), -- Estado del clima (Soleado, Lluvia, etc.)
    observations TEXT, -- Notas generales del día
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Evitamos duplicidad de reportes para el mismo proyecto el mismo día
    CONSTRAINT unique_project_date_report UNIQUE (project_id, report_date)
);

-- Tabla de Entradas de Avance (Progreso físico de las tareas)
CREATE TABLE progress_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    daily_report_id UUID NOT NULL REFERENCES daily_reports(id) ON DELETE CASCADE,
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE, -- Tarea del cronograma (Módulo 13)
    progress_percentage NUMERIC(5, 2) NOT NULL CHECK (progress_percentage >= 0 AND progress_percentage <= 100), -- % acumulado o del día
    quantity_executed NUMERIC(12, 2) DEFAULT 0.00, -- Metraje o cantidad ejecutada si aplica
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_task_per_report UNIQUE (daily_report_id, task_id)
);

-- Índices optimizados para las consultas multi-tenant y por proyecto
CREATE INDEX idx_daily_reports_company_project ON daily_reports(company_id, project_id);
CREATE INDEX idx_progress_entries_report ON progress_entries(daily_report_id);