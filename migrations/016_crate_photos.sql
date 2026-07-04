-- Módulo 15: Evidencia Fotográfica
CREATE TABLE project_photos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    
    -- Relaciones opcionales dependientes del contexto de la foto
    task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,          -- Módulo 13 (Cronograma)
    daily_report_id UUID REFERENCES daily_reports(id) ON DELETE SET NULL, -- Módulo 14 (Avance)
    
    user_id UUID NOT NULL REFERENCES users(id), -- Quién subió la foto
    photo_url TEXT NOT NULL,                     -- URL del Storage de Supabase
    description TEXT,                           -- Nota/Pie de foto
    latitude NUMERIC(10, 8),                    -- Geolocalización opcional para certificar que es en la obra
    longitude NUMERIC(11, 8),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Índices eficientes para búsquedas rápidas en dashboards y apps móviles
CREATE INDEX idx_photos_company_project ON project_photos(company_id, project_id);
CREATE INDEX idx_photos_daily_report ON project_photos(daily_report_id) WHERE daily_report_id IS NOT NULL;
CREATE INDEX idx_photos_task ON project_photos(task_id) WHERE task_id IS NOT NULL;  