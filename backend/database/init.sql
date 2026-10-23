-- Estructura relacional + datos semilla del proyecto Catalogo (Fases 2, 4).
-- Se ejecuta automáticamente la primera vez que se levanta el contenedor de PostgreSQL.
-- Los datos son EXACTAMENTE los que estaban quemados en el código:
--   usuarios  -> authController.go (login de admin)
--   productos -> prodController.go (4 productos)
--   referidos -> MiRed.tsx (3 referidos del frontend)
-- (Fase 4) Se agregan: usuario 'cliente' (compra desde el checkout), stock, saldo de referidos
-- y las tablas pedidos, pedido_items y comisiones.

-- pgcrypto permite guardar contraseñas con bcrypt (crypt / gen_salt) directamente desde SQL.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ========== USUARIOS ==========
CREATE TABLE IF NOT EXISTS usuarios (
    id       SERIAL PRIMARY KEY,
    email    VARCHAR(150) NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL,           -- hash bcrypt (nunca texto plano)
    rol      VARCHAR(20)  NOT NULL CHECK (rol IN ('admin', 'cliente'))
);

INSERT INTO usuarios (email, password, rol) VALUES
    ('admin@upse.edu.ec',   crypt('123456', gen_salt('bf', 10)), 'admin'),
    ('cliente@upse.edu.ec', crypt('123456', gen_salt('bf', 10)), 'cliente');   -- Fase 4: quien compra desde el checkout

-- ========== PRODUCTOS ==========
CREATE TABLE IF NOT EXISTS productos (
    id          SERIAL PRIMARY KEY,
    nombre      VARCHAR(150)  NOT NULL,
    descripcion TEXT          NOT NULL DEFAULT '',
    precio      NUMERIC(10,2) NOT NULL CHECK (precio >= 0),
    categoria   VARCHAR(50)   NOT NULL,
    img         TEXT          NOT NULL,
    galeria     TEXT[]        NOT NULL DEFAULT '{}',
    stock       INTEGER       NOT NULL DEFAULT 100 CHECK (stock >= 0)   -- Fase 4: stock simulado
);

INSERT INTO productos (id, nombre, descripcion, precio, categoria, img, galeria) VALUES
(1, 'Serum Revitalizante',
 'Serum concentrado con vitamina C y ácido hialurónico. Ilumina la piel, reduce manchas y aporta hidratación profunda desde la primera aplicación.',
 45.00, 'Serum', 'https://picsum.photos/seed/serum/150',
 ARRAY['https://picsum.photos/seed/serum1/600','https://picsum.photos/seed/serum2/600','https://picsum.photos/seed/serum3/600']),
(2, 'Crema Hidratante Pro',
 'Crema facial de textura ligera con niacinamida y manteca de karité. Hidratación de 24 horas y barrera cutánea fortalecida para todo tipo de piel.',
 32.50, 'Crema', 'https://picsum.photos/seed/crema/150',
 ARRAY['https://picsum.photos/seed/crema1/600','https://picsum.photos/seed/crema2/600','https://picsum.photos/seed/crema3/600']),
(3, 'Tónico Purificante',
 'Tónico sin alcohol con hamamelis y agua de rosas. Limpia poros, equilibra el pH y prepara la piel para el resto de la rutina diaria.',
 28.00, 'Tónico', 'https://picsum.photos/seed/tonico/150',
 ARRAY['https://picsum.photos/seed/tonico1/600','https://picsum.photos/seed/tonico2/600','https://picsum.photos/seed/tonico3/600']),
(4, 'Mascarilla Nocturna',
 'Mascarilla de noche con colágeno y vitamina E. Repara la piel mientras duermes y recupera la luminosidad al despertar.',
 50.00, 'Mascarilla', 'https://picsum.photos/seed/mascarilla/150',
 ARRAY['https://picsum.photos/seed/mascarilla1/600','https://picsum.photos/seed/mascarilla2/600','https://picsum.photos/seed/mascarilla3/600']);

-- Como insertamos ids a mano, sincronizamos la secuencia para que los próximos INSERT no choquen.
SELECT setval(pg_get_serial_sequence('productos', 'id'), (SELECT MAX(id) FROM productos));

-- ========== REFERIDOS (árbol multinivel con auto-referencia) ==========
CREATE TABLE IF NOT EXISTS referidos (
    id        INTEGER PRIMARY KEY,
    nombre    VARCHAR(100)  NOT NULL,
    nivel     INTEGER       NOT NULL CHECK (nivel >= 0),
    ventas    NUMERIC(10,2) NOT NULL DEFAULT 0,
    parent_id INTEGER REFERENCES referidos(id),   -- NULL = raíz del árbol ("Tú")
    saldo     NUMERIC(12,2) NOT NULL DEFAULT 0,   -- Fase 4: comisiones acumuladas
    usuario_id INTEGER UNIQUE REFERENCES usuarios(id)  -- Fase 4: qué usuario es este nodo (si tiene cuenta)
);

-- id 0 = el administrador (raíz). Ana y Luis son nivel 1 (directos); Marta es nivel 2,
-- así que la colgamos de Ana. (Fase 4) 'Cliente Demo' (nivel 3, cuelga de Marta) es el nodo del usuario
-- cliente@upse.edu.ec: cuando compra, cobran comisión sus 3 ascendentes (Marta, Ana y el administrador).
INSERT INTO referidos (id, nombre, nivel, ventas, parent_id) VALUES
(0, 'Tú',            0,    0, NULL),
(1, 'Ana García',    1, 1200, 0),
(2, 'Luis Poveda',   1,  850, 0),
(3, 'Marta Sánchez', 2,  430, 1),
(4, 'Cliente Demo',  3,    0, 3);

UPDATE referidos SET usuario_id = (SELECT id FROM usuarios WHERE email = 'admin@upse.edu.ec')   WHERE id = 0;
UPDATE referidos SET usuario_id = (SELECT id FROM usuarios WHERE email = 'cliente@upse.edu.ec') WHERE id = 4;

-- ========== PEDIDOS (Fase 4: transacción ACID) ==========
CREATE TABLE IF NOT EXISTS pedidos (
    id                    SERIAL PRIMARY KEY,
    usuario_id            INTEGER       NOT NULL REFERENCES usuarios(id),
    cliente_nombre        VARCHAR(150)  NOT NULL,
    cliente_email         VARCHAR(150)  NOT NULL,
    direccion             VARCHAR(250)  NOT NULL,
    ciudad                VARCHAR(100)  NOT NULL,
    telefono              VARCHAR(30)   NOT NULL DEFAULT '',
    metodo_pago           VARCHAR(50)   NOT NULL,
    total                 NUMERIC(12,2) NOT NULL DEFAULT 0,
    estado                VARCHAR(20)   NOT NULL DEFAULT 'pagado',
    comisiones_calculadas BOOLEAN       NOT NULL DEFAULT FALSE,  -- la Goroutine la pasa a TRUE
    fecha                 TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pedido_items (
    id              SERIAL PRIMARY KEY,
    pedido_id       INTEGER       NOT NULL REFERENCES pedidos(id) ON DELETE CASCADE,
    producto_id     INTEGER       NOT NULL REFERENCES productos(id),  -- id inexistente => viola la clave foránea
    cantidad        INTEGER       NOT NULL CHECK (cantidad > 0),
    precio_unitario NUMERIC(10,2) NOT NULL
);

-- Registro de cada comisión pagada. El UNIQUE evita pagar dos veces la misma comisión del mismo pedido.
CREATE TABLE IF NOT EXISTS comisiones (
    id          SERIAL PRIMARY KEY,
    pedido_id   INTEGER       NOT NULL REFERENCES pedidos(id),
    referido_id INTEGER       NOT NULL REFERENCES referidos(id),
    nivel       INTEGER       NOT NULL CHECK (nivel BETWEEN 1 AND 3),
    monto       NUMERIC(12,2) NOT NULL,
    fecha       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE (pedido_id, referido_id, nivel)
);

-- ========== ÍNDICES (consultas frecuentes) ==========
CREATE INDEX IF NOT EXISTS idx_referidos_parent_id    ON referidos(parent_id);
CREATE INDEX IF NOT EXISTS idx_pedidos_cliente_email  ON pedidos(cliente_email);
CREATE INDEX IF NOT EXISTS idx_pedido_items_pedido_id ON pedido_items(pedido_id);
CREATE INDEX IF NOT EXISTS idx_productos_categoria    ON productos(categoria);

-- ========== FASE 0: BACK-OFFICE (misma migración 004_backoffice.sql) ==========
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

-- ========== FASES 2 y 3: PEDIDOS Y PRODUCTOS (misma migración 005_pedidos_productos.sql) ==========
ALTER TABLE pedidos ALTER COLUMN estado SET DEFAULT 'pendiente';
UPDATE pedidos SET estado = 'pendiente' WHERE estado = 'pagado';
ALTER TABLE pedidos DROP CONSTRAINT IF EXISTS pedidos_estado_check;
ALTER TABLE pedidos ADD CONSTRAINT pedidos_estado_check
    CHECK (estado IN ('pendiente', 'preparando', 'enviado', 'entregado', 'cancelado'));

-- Historial de pedidos por usuario (GET /api/pedidos/mios).
CREATE INDEX IF NOT EXISTS idx_pedidos_usuario_fecha ON pedidos(usuario_id, fecha DESC);
CREATE INDEX IF NOT EXISTS idx_pedidos_estado        ON pedidos(estado);

-- ========== FASE 6: REPORTES (misma migración 006_reportes.sql) ==========
CREATE INDEX IF NOT EXISTS idx_pedidos_fecha ON pedidos(fecha);
