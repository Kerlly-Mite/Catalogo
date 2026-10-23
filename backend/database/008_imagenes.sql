-- 008_imagenes.sql: fotos de productos subidas desde el dispositivo.
-- Se guardan en la BASE DE DATOS (no en disco): en Render el disco de la API se borra en cada despliegue.
-- Idempotente: se puede ejecutar varias veces.
CREATE TABLE IF NOT EXISTS imagenes (
    id        SERIAL PRIMARY KEY,
    tipo      VARCHAR(50)  NOT NULL CHECK (tipo IN ('image/jpeg', 'image/png', 'image/webp')),
    datos     BYTEA        NOT NULL,
    tamano    INTEGER      NOT NULL,
    creado_en TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
