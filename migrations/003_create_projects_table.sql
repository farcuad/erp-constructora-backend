CREATE TABLE project_statuses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL -- Planeación, En ejecución, Suspendido, Finalizado
);

CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    status_id INT NOT NULL REFERENCES project_statuses(id),
    name VARCHAR(150) NOT NULL,
    client_name VARCHAR(150), -- O el UUID de tu tabla de clientes más adelante
    location TEXT,
    start_date DATE,
    end_date DATE,
    budget NUMERIC(15, 2) DEFAULT 0.00, -- Usamos NUMERIC para dinero, jamás FLOAT o REAL
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Para saber qué ingenieros/supervisores están asignados a qué obra
CREATE TABLE project_members (
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (project_id, user_id)
);
