-- 006_reportes.sql (Fase 6): los reportes filtran pedidos por rango de fechas (GET /api/reportes/*?desde&hasta).
-- Idempotente: se puede ejecutar varias veces sobre el volumen existente.
CREATE INDEX IF NOT EXISTS idx_pedidos_fecha ON pedidos(fecha);
