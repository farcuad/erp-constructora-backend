-- 1. Activar la extensión pgvector en PostgreSQL
CREATE EXTENSION IF NOT EXISTS vector;

-- 2. Tabla para almacenar los Chunks y sus Embeddings (Vectores)
CREATE TABLE document_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    document_version_id UUID NOT NULL REFERENCES document_versions(id) ON DELETE CASCADE,
    
    chunk_index INT NOT NULL,              -- Posición del fragmento (0, 1, 2...)
    chunk_content TEXT NOT NULL,           -- El fragmento de texto extraído
    embedding vector(1536) NOT NULL,       -- Vector generado (1536 dimensiones para OpenAI text-embedding-3-small)
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Índices de rendimiento para RAG
-- Búsqueda vectorial rápida (HNSW con similitud coseno)
CREATE INDEX idx_document_embeddings_vector 
ON document_embeddings 
USING hnsw (embedding vector_cosine_ops);

-- Filtro multi-tenant prioritario
CREATE INDEX idx_document_embeddings_company 
ON document_embeddings(company_id);