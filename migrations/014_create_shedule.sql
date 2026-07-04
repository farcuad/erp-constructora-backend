-- 1. Tabla de Tareas / Actividades
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,       -- Ej: "Excavación de fundaciones"
    description TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    progress NUMERIC(5, 2) DEFAULT 0.00, -- Porcentaje de avance (0.00 a 100.00)
    status VARCHAR(50) DEFAULT 'To Do',  -- 'To Do', 'In Progress', 'Blocked', 'Done'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Tabla de Dependencias entre Tareas (Relaciones Sucesor-Predecesor)
CREATE TABLE task_dependencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,             -- Tarea dependiente (Sucesora)
    depends_on_uuid UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,     -- Tarea de la que depende (Predecesora)
    dependency_type VARCHAR(20) DEFAULT 'FS', -- 'FS' (Finish-to-Start), 'SS' (Start-to-Start), etc.
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_task_dependency UNIQUE (task_id, depends_on_uuid)
);

-- 3. Tabla de Hitos (Milestones)
CREATE TABLE milestones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,       -- Ej: "Finalización de Estructura Principal"
    due_date DATE NOT NULL,           -- Fecha del hito
    is_achieved BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);