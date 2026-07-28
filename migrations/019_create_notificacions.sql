-- Módulo 19: Notificaciones

-- 1. Tabla principal de notificaciones (El evento generado)
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE, -- Opcional, por si es una alerta global de la empresa
    title VARCHAR(150) NOT NULL,
    message TEXT NOT NULL,
    link_to_ui VARCHAR(255), -- Ruta de redirección en React o Flutter (Ej: "/dashboard/projects/4f/billing")
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Tabla pivote para el estado de lectura por usuario
CREATE TABLE notification_reads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT unique_user_notification UNIQUE (notification_id, user_id)
);

-- Agregamos los campos para soportar el patrón polimórfico y metadata en notifications
ALTER TABLE notifications 
ADD COLUMN entity_type VARCHAR(50) NOT NULL DEFAULT 'SYSTEM',
ADD COLUMN entity_id UUID,
ADD COLUMN type VARCHAR(50) NOT NULL DEFAULT 'GENERAL_ALERT',
ADD COLUMN priority VARCHAR(20) NOT NULL DEFAULT 'medium',
ADD COLUMN metadata JSONB DEFAULT '{}'::jsonb;

-- Índice para acelerar la consulta de notificaciones no leídas por usuario en el Dashboard/Sidebar
CREATE INDEX IF NOT EXISTS idx_notification_reads_user_unread 
ON notification_reads (company_id, user_id, is_read);

-- Índices eficientes para obtener las notificaciones no leídas de un usuario en tiempo real
CREATE INDEX idx_notifications_company ON notifications(company_id);
CREATE INDEX idx_notification_reads_user ON notification_reads(user_id, is_read);