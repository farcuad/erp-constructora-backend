-- 1. Cabecera de Asistencia (Controla el día y el proyecto)
CREATE TABLE attendance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- No puede haber dos planillas de asistencia para el mismo proyecto el mismo día
    CONSTRAINT unique_project_date_attendance UNIQUE (project_id, date)
);

-- 2. Registro Detallado de Asistencia por Empleado
CREATE TABLE attendance_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attendance_id UUID NOT NULL REFERENCES attendance(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL, -- 'Present', 'Absent', 'Late', 'Justified Absence', 'Medical Leave'
    hours_worked NUMERIC(4, 2) DEFAULT 0.00, -- Ej: 8.00, 4.50, 10.00 (horas extra)
    notes TEXT, -- Observaciones como "Llegó tarde por transporte" o "Incapacidad médica"
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Un empleado solo puede tener un registro en una planilla de asistencia específica
    CONSTRAINT unique_attendance_employee UNIQUE (attendance_id, employee_id)
);