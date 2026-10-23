-- 003_seguridad.sql: contraseñas con bcrypt + índices.
-- Para volúmenes YA creados (el init.sql solo corre la primera vez). Es idempotente: se puede ejecutar varias veces.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Convierte a bcrypt solo las contraseñas que todavía están en texto plano (los hashes bcrypt empiezan con $2).
UPDATE usuarios
   SET password = crypt(password, gen_salt('bf', 10))
 WHERE password NOT LIKE '$2%';

-- Índices para las consultas más frecuentes.
CREATE INDEX IF NOT EXISTS idx_referidos_parent_id    ON referidos(parent_id);
CREATE INDEX IF NOT EXISTS idx_pedidos_cliente_email  ON pedidos(cliente_email);
CREATE INDEX IF NOT EXISTS idx_pedido_items_pedido_id ON pedido_items(pedido_id);
CREATE INDEX IF NOT EXISTS idx_productos_categoria    ON productos(categoria);
