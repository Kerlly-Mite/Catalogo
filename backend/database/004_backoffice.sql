-- 004_backoffice.sql (Fase 0 del PLAN_IMPLEMENTACION_U2): columnas y tablas que necesitan las fases 1 a 7.
-- Idempotente (IF NOT EXISTS): se puede ejecutar varias veces sobre el volumen existente.
-- Adaptada a ESTE proyecto: pedidos.estado y pedidos.usuario_id ya existían desde la Fase 4 del DRS.

-- ========== USUARIOS ==========
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS nombre     VARCHAR(100);
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS activo     BOOLEAN     NOT NULL DEFAULT TRUE;
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS creado_en  TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- ========== REFERIDOS ==========
ALTER TABLE referidos ADD COLUMN IF NOT EXISTS codigo VARCHAR(20) UNIQUE;   -- código del link de invitación

-- Los referidos se insertaban con id manual; ahora los nuevos (registro) toman el id de una secuencia.
CREATE SEQUENCE IF NOT EXISTS referidos_id_seq OWNED BY referidos.id;
ALTER TABLE referidos ALTER COLUMN id SET DEFAULT nextval('referidos_id_seq');
SELECT setval('referidos_id_seq', (SELECT COALESCE(MAX(id), 0) FROM referidos) + 1, false);

-- Códigos de invitación del seed.
UPDATE referidos SET codigo = CASE id
        WHEN 0 THEN 'TU'
        WHEN 1 THEN 'ANA'
        WHEN 2 THEN 'LUIS'
        WHEN 3 THEN 'MARTA'
        WHEN 4 THEN 'CLIENTE'
    END
 WHERE codigo IS NULL AND id IN (0, 1, 2, 3, 4);

-- FK real usuario -> nodo de la red (antes el vínculo era solo referidos.usuario_id).
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS referido_id INTEGER REFERENCES referidos(id);
UPDATE usuarios u SET referido_id = r.id FROM referidos r WHERE r.usuario_id = u.id AND u.referido_id IS NULL;
UPDATE usuarios u SET nombre = r.nombre FROM referidos r WHERE r.id = u.referido_id AND u.nombre IS NULL;

-- ========== PRODUCTOS ==========
ALTER TABLE productos ADD COLUMN IF NOT EXISTS activo BOOLEAN NOT NULL DEFAULT TRUE;   -- borrado lógico

-- ========== LIQUIDACIONES ==========
CREATE TABLE IF NOT EXISTS liquidaciones (
    id          SERIAL PRIMARY KEY,
    referido_id INTEGER       NOT NULL REFERENCES referidos(id),
    total       NUMERIC(12,2) NOT NULL,
    estado      VARCHAR(20)   NOT NULL DEFAULT 'pagada',
    creado_en   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    pagada_en   TIMESTAMPTZ
);

-- ========== COMISIONES (ledger) ==========
ALTER TABLE comisiones ADD COLUMN IF NOT EXISTS porcentaje     NUMERIC(5,2);
ALTER TABLE comisiones ADD COLUMN IF NOT EXISTS estado         VARCHAR(20) NOT NULL DEFAULT 'pendiente'
    CHECK (estado IN ('pendiente', 'aprobada', 'pagada', 'revertida'));
ALTER TABLE comisiones ADD COLUMN IF NOT EXISTS liquidacion_id INTEGER REFERENCES liquidaciones(id);
UPDATE comisiones SET porcentaje = CASE nivel WHEN 1 THEN 10 WHEN 2 THEN 5 WHEN 3 THEN 2 END WHERE porcentaje IS NULL;
CREATE INDEX IF NOT EXISTS idx_comisiones_referido_pedido ON comisiones(referido_id, pedido_id);
