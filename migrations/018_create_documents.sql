-- Módulo 18: Gestión Documental

-- 1. Categorías o Tipos de documentos (Ej: Planos, Contratos, Permisos, Presupuestos)
CREATE TABLE document_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_type_per_company UNIQUE (company_id, name)
);

-- 2. Registro Principal del Documento (Contenedor lógico)
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    document_type_id UUID NOT NULL REFERENCES document_types(id) ON DELETE RESTRICT,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    current_version INT NOT NULL DEFAULT 1,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, ARCHIVED, DEPRECATED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Historial de Versiones del Documento (Cada actualización física del archivo)
CREATE TABLE document_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    version_number INT NOT NULL, -- 1, 2, 3, etc.
    file_url TEXT NOT NULL,       -- URL pública o privada del Storage de Supabase
    file_size INT,               -- En bytes
    file_extension VARCHAR(10),  -- pdf, dwg, xlsx, etc.
    change_log TEXT,             -- Nota de qué cambió en esta versión (Ej: "Ajuste de vigas en eje B")
    user_id UUID NOT NULL REFERENCES users(id), -- Quién subió esta versión específica
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_version_per_document UNIQUE (document_id, version_number)
);

-- Índices optimizados
CREATE INDEX idx_documents_lookup ON documents(company_id, project_id, document_type_id);
CREATE INDEX idx_document_versions ON document_versions(document_id, version_number);