-- Módulo 27: Publicación de versiones de la App (APK / IPK)
-- Guarda cada compilación de la aplicación móvil (Flutter) para descargarse desde la web
-- y para que la app detecte si hay una versión más nueva instalada.

CREATE TABLE app_releases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version VARCHAR(50) NOT NULL,          -- Versión de la app (ej: 1.0.0, 2.1.5)
    app_url TEXT NOT NULL,                 -- URL pública del .apk (Supabase Storage / S3)
    description TEXT,                      -- Notas del release (ej: "Se agregó X")
    file_size BIGINT,                      -- Tamaño en bytes del archivo
    checksum VARCHAR(128),                 -- Hash (SHA-256) opcional para verificar integridad
    is_mandatory BOOLEAN NOT NULL DEFAULT FALSE, -- TRUE obliga al usuario a actualizar
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE push_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    platform VARCHAR(20) NOT NULL, -- 'android' / 'ios'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_push_tokens_user_id ON push_tokens(user_id);

-- Índice para traer eficientemente "el último release publicado"
CREATE INDEX idx_app_releases_created_at
ON app_releases(created_at DESC);