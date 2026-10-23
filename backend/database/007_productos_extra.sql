-- 007_productos_extra.sql
-- Agrega los productos 5 a 8 a una base de datos que YA existe (local o Render).
-- Es seguro ejecutarlo varias veces: si el id ya existe, no hace nada.

INSERT INTO productos (id, nombre, descripcion, precio, categoria, img, galeria) VALUES
(5, 'Serum Anti-Edad Retinol',
 'Serum de retinol encapsulado que suaviza líneas de expresión y mejora la firmeza. Uso nocturno recomendado.',
 58.00, 'Serum', 'https://picsum.photos/seed/retinol/150',
 ARRAY['https://picsum.photos/seed/retinol1/600','https://picsum.photos/seed/retinol2/600','https://picsum.photos/seed/retinol3/600']),
(6, 'Crema Contorno de Ojos',
 'Contorno de ojos con cafeína y péptidos. Reduce bolsas y ojeras, hidrata la zona más delicada del rostro.',
 26.00, 'Crema', 'https://picsum.photos/seed/ojos/150',
 ARRAY['https://picsum.photos/seed/ojos1/600','https://picsum.photos/seed/ojos2/600','https://picsum.photos/seed/ojos3/600']),
(7, 'Tónico Exfoliante AHA-BHA',
 'Exfoliación química suave con ácidos AHA y BHA. Renueva la textura de la piel y desobstruye los poros.',
 34.00, 'Tónico', 'https://picsum.photos/seed/exfoliante/150',
 ARRAY['https://picsum.photos/seed/exfoliante1/600','https://picsum.photos/seed/exfoliante2/600','https://picsum.photos/seed/exfoliante3/600']),
(8, 'Kit Rutina Completa',
 'Kit integral con serum, crema, tónico y mascarilla. La rutina perfecta para comenzar: limpieza, tratamiento e hidratación.',
 129.00, 'Kit', 'https://picsum.photos/seed/kit/150',
 ARRAY['https://picsum.photos/seed/kit1/600','https://picsum.photos/seed/kit2/600','https://picsum.photos/seed/kit3/600'])
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('productos', 'id'), (SELECT MAX(id) FROM productos));
