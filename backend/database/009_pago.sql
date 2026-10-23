-- 009_pago.sql: datos del pago por método (tarjeta / transferencia / efectivo) y comprobantes de transferencia.
-- Idempotente: se puede ejecutar varias veces.
-- NUNCA se guarda el número completo de la tarjeta ni el CVV: solo los últimos 4 dígitos.
ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS pago_detalle JSONB NOT NULL DEFAULT '{}'::jsonb;

-- Vouchers de transferencia (imagen o PDF). Se guardan en la BASE DE DATOS (el disco de Render se borra en cada despliegue).
-- No son públicos: solo el admin puede verlos (GET /api/comprobantes/:id).
CREATE TABLE IF NOT EXISTS comprobantes (
    id         SERIAL PRIMARY KEY,
    usuario_id INTEGER      NOT NULL REFERENCES usuarios(id),
    tipo       VARCHAR(50)  NOT NULL CHECK (tipo IN ('image/jpeg', 'image/png', 'image/webp', 'application/pdf')),
    datos      BYTEA        NOT NULL,
    tamano     INTEGER      NOT NULL,
    creado_en  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
