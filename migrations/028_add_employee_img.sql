-- Módulo 28: Foto del empleado
-- Agrega una columna para guardar la URL de la fotografía del empleado
-- (debe apuntar a una URL pública, ej: Supabase Storage / S3 / CDN).

ALTER TABLE employees
    ADD COLUMN IF NOT EXISTS img_url TEXT;
