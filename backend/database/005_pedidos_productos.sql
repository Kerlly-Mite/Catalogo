-- 005_pedidos_productos.sql (Fases 2 y 3 del PLAN_IMPLEMENTACION_U2). Idempotente.
-- Estados del pedido: pendiente -> preparando -> enviado -> entregado, o cancelado (solo antes de enviar).

-- El checkout ya no crea el pedido como 'pagado': nace 'pendiente'. Los pedidos viejos se convierten.
ALTER TABLE pedidos ALTER COLUMN estado SET DEFAULT 'pendiente';
UPDATE pedidos SET estado = 'pendiente' WHERE estado = 'pagado';
ALTER TABLE pedidos DROP CONSTRAINT IF EXISTS pedidos_estado_check;
ALTER TABLE pedidos ADD CONSTRAINT pedidos_estado_check
    CHECK (estado IN ('pendiente', 'preparando', 'enviado', 'entregado', 'cancelado'));

-- Historial de pedidos por usuario (GET /api/pedidos/mios).
CREATE INDEX IF NOT EXISTS idx_pedidos_usuario_fecha ON pedidos(usuario_id, fecha DESC);
CREATE INDEX IF NOT EXISTS idx_pedidos_estado        ON pedidos(estado);
