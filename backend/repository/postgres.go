package repository

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"multicatalogo-backend/domain"
	"multicatalogo-backend/models"
)

// Opciones permite activar fallas controladas para DEMOSTRAR el ROLLBACK en clase.
// En uso normal ambas están apagadas.
type Opciones struct {
	// SimularFalla: aborta la compra después de insertar el pedido y antes de insertar los items.
	SimularFalla bool
	// PausaAntesCommit: espera justo antes del COMMIT (para tumbar el contenedor de BD a mitad de la compra).
	PausaAntesCommit time.Duration
}

// PostgresRepository implementa DBRepository sobre un pool de pgx.
type PostgresRepository struct {
	db   *pgxpool.Pool
	opts Opciones
}

// NewPostgresRepository recibe el pool ya creado (inyección) y devuelve el repositorio.
func NewPostgresRepository(db *pgxpool.Pool, opts Opciones) *PostgresRepository {
	return &PostgresRepository{db: db, opts: opts}
}

// conTimeout limita el tiempo de una consulta para que ninguna quede colgada.
func conTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, d)
}

// Ping comprueba que PostgreSQL responde.
func (r *PostgresRepository) Ping(ctx context.Context) error {
	ctx, cancel := conTimeout(ctx, 2*time.Second)
	defer cancel()
	return r.db.Ping(ctx)
}

// ---------- USUARIOS ----------

func (r *PostgresRepository) BuscarUsuarioPorEmail(ctx context.Context, email string) (*models.Usuario, error) {
	ctx, cancel := conTimeout(ctx, 5*time.Second)
	defer cancel()

	var u models.Usuario
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password, rol, COALESCE(nombre, ''), activo FROM usuarios WHERE LOWER(email) = LOWER($1)`, email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Rol, &u.Nombre, &u.Activo)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ---------- PRODUCTOS ----------

// columnasProducto: el orden coincide con escanearProducto.
const columnasProducto = `id, nombre, descripcion, precio, categoria, img, galeria, stock, activo`

func escanearProducto(row pgx.Row) (*models.Producto, error) {
	var p models.Producto
	if err := row.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &p.Galeria, &p.Stock, &p.Activo); err != nil {
		return nil, err
	}
	if p.Galeria == nil {
		p.Galeria = []string{}
	}
	return &p, nil
}

func (r *PostgresRepository) ObtenerProductos(ctx context.Context) ([]models.Producto, error) {
	ctx, cancel := conTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.Query(ctx, `SELECT `+columnasProducto+` FROM productos WHERE activo ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	productos := []models.Producto{} // slice vacío (no nil) para que el JSON sea [] y no null
	for rows.Next() {
		p, err := escanearProducto(rows)
		if err != nil {
			return nil, err
		}
		productos = append(productos, *p)
	}
	return productos, rows.Err()
}

func (r *PostgresRepository) ObtenerProductoPorID(ctx context.Context, id int) (*models.Producto, error) {
	ctx, cancel := conTimeout(ctx, 5*time.Second)
	defer cancel()

	p, err := escanearProducto(r.db.QueryRow(ctx, `SELECT `+columnasProducto+` FROM productos WHERE id = $1 AND activo`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	return p, err
}

// escaparLike evita que % y _ escritos por el usuario actúen como comodines en ILIKE.
func escaparLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// ListarProductosAdmin devuelve TODOS los productos (activos e inactivos), con búsqueda ?q= (ILIKE) y paginación.
func (r *PostgresRepository) ListarProductosAdmin(ctx context.Context, q string, page, limit int) ([]models.Producto, int, error) {
	ctx, cancel := conTimeout(ctx, 5*time.Second)
	defer cancel()

	patron := "%" + escaparLike(strings.TrimSpace(q)) + "%"
	const filtro = `(nombre ILIKE $1 OR descripcion ILIKE $1 OR categoria ILIKE $1)`

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM productos WHERE `+filtro, patron).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx,
		`SELECT `+columnasProducto+` FROM productos WHERE `+filtro+` ORDER BY id LIMIT $2 OFFSET $3`,
		patron, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	productos := []models.Producto{}
	for rows.Next() {
		p, err := escanearProducto(rows)
		if err != nil {
			return nil, 0, err
		}
		productos = append(productos, *p)
	}
	return productos, total, rows.Err()
}

func (r *PostgresRepository) CrearProducto(ctx context.Context, in models.ProductoEntrada) (*models.Producto, error) {
	ctx, cancel := conTimeout(ctx, 5*time.Second)
	defer cancel()

	if in.Galeria == nil {
		in.Galeria = []string{}
	}
	return escanearProducto(r.db.QueryRow(ctx,
		`INSERT INTO productos (nombre, descripcion, precio, categoria, img, galeria, stock)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING `+columnasProducto,
		in.Nombre, in.Descripcion, in.Precio, in.Categoria, in.Img, in.Galeria, in.Stock))
}

// ActualizarProducto cambia SOLO los campos enviados (COALESCE conserva el valor actual si llega NULL).
func (r *PostgresRepository) ActualizarProducto(ctx context.Context, id int, c models.ProductoCambios) (*models.Producto, error) {
	ctx, cancel := conTimeout(ctx, 5*time.Second)
	defer cancel()

	p, err := escanearProducto(r.db.QueryRow(ctx,
		`UPDATE productos SET
			nombre      = COALESCE($2, nombre),
			descripcion = COALESCE($3, descripcion),
			precio      = COALESCE($4, precio),
			categoria   = COALESCE($5, categoria),
			img         = COALESCE($6, img),
			galeria     = COALESCE($7, galeria),
			stock       = COALESCE($8, stock),
			activo      = COALESCE($9, activo)
		 WHERE id = $1 RETURNING `+columnasProducto,
		id, c.Nombre, c.Descripcion, c.Precio, c.Categoria, c.Img, c.Galeria, c.Stock, c.Activo))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	return p, err
}

// DesactivarProducto es el "DELETE" lógico: la fila se conserva (los pedidos antiguos la referencian).
func (r *PostgresRepository) DesactivarProducto(ctx context.Context, id int) error {
	ctx, cancel := conTimeout(ctx, 5*time.Second)
	defer cancel()

	tag, err := r.db.Exec(ctx, `UPDATE productos SET activo = FALSE WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNoEncontrado
	}
	return nil
}

// ---------- PEDIDOS: consulta y gestión ----------

func numeroPedido(id int) string { return fmt.Sprintf("MC-%06d", id) }

// ListarPedidos devuelve una página de pedidos (más recientes primero) y el total que cumple el filtro.
func (r *PostgresRepository) ListarPedidos(ctx context.Context, f FiltroPedidos) ([]models.PedidoListado, int, error) {
	ctx, cancel := conTimeout(ctx, 5*time.Second)
	defer cancel()

	const filtro = `($1 = 0 OR usuario_id = $1) AND ($2 = '' OR estado = $2)`
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM pedidos WHERE `+filtro, f.UsuarioID, f.Estado).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, fecha, total, estado, cliente_nombre, cliente_email,
		        (SELECT COALESCE(SUM(cantidad), 0) FROM pedido_items WHERE pedido_id = pedidos.id)
		   FROM pedidos WHERE `+filtro+`
		  ORDER BY fecha DESC, id DESC LIMIT $3 OFFSET $4`,
		f.UsuarioID, f.Estado, f.Limit, (f.Page-1)*f.Limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	lista := []models.PedidoListado{}
	for rows.Next() {
		var p models.PedidoListado
		var fecha time.Time
		if err := rows.Scan(&p.ID, &fecha, &p.Total, &p.Estado, &p.ClienteNombre, &p.ClienteEmail, &p.NumItems); err != nil {
			return nil, 0, err
		}
		p.Numero = numeroPedido(p.ID)
		p.Fecha = fecha.Format(time.RFC3339)
		lista = append(lista, p)
	}
	return lista, total, rows.Err()
}

// ObtenerPedido devuelve la cabecera + los items (con el nombre del producto, por JOIN).
func (r *PostgresRepository) ObtenerPedido(ctx context.Context, id int) (*models.PedidoDetalle, error) {
	ctx, cancel := conTimeout(ctx, 5*time.Second)
	defer cancel()

	var d models.PedidoDetalle
	var fecha time.Time
	var detalleJSON []byte
	err := r.db.QueryRow(ctx,
		`SELECT id, usuario_id, fecha, total, estado, cliente_nombre, cliente_email, direccion, ciudad, telefono, metodo_pago, pago_detalle
		   FROM pedidos WHERE id = $1`, id,
	).Scan(&d.ID, &d.UsuarioID, &fecha, &d.Total, &d.Estado, &d.ClienteNombre, &d.ClienteEmail,
		&d.Direccion, &d.Ciudad, &d.Telefono, &d.MetodoPago, &detalleJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, err
	}
	d.Numero = numeroPedido(d.ID)
	d.Fecha = fecha.Format(time.RFC3339)
	d.PagoDetalle = map[string]string{}
	if len(detalleJSON) > 0 {
		_ = json.Unmarshal(detalleJSON, &d.PagoDetalle) // si el JSON estuviera dañado, se muestra sin detalle
	}

	rows, err := r.db.Query(ctx,
		`SELECT pi.producto_id, p.nombre, pi.cantidad, pi.precio_unitario
		   FROM pedido_items pi JOIN productos p ON p.id = pi.producto_id
		  WHERE pi.pedido_id = $1 ORDER BY pi.id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	d.Items = []models.PedidoItem{}
	for rows.Next() {
		var it models.PedidoItem
		if err := rows.Scan(&it.ProductoID, &it.Nombre, &it.Cantidad, &it.PrecioUnitario); err != nil {
			return nil, err
		}
		d.Items = append(d.Items, it)
		d.NumItems += it.Cantidad
	}
	return &d, rows.Err()
}

// transicionesPedido: de qué estado a cuáles se puede pasar. Solo se cancela antes de enviar.
var transicionesPedido = map[string][]string{
	"pendiente":  {"preparando", "cancelado"},
	"preparando": {"enviado", "cancelado"},
	"enviado":    {"entregado"},
}

func transicionPermitida(actual, nuevo string) bool {
	for _, destino := range transicionesPedido[actual] {
		if destino == nuevo {
			return true
		}
	}
	return false
}

// CambiarEstadoPedido aplica la transición en UNA transacción. Cancelar es el espejo ACID del checkout:
// devuelve el stock, revierte las comisiones aún no liquidadas (y su saldo) y resta la venta al comprador.
func (r *PostgresRepository) CambiarEstadoPedido(ctx context.Context, id int, nuevo string) error {
	ctx, cancel := conTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	// Bloqueamos el pedido: dos cambios de estado simultáneos se hacen uno tras otro.
	var actual string
	var usuarioID int
	var total float64
	var comisionesCalculadas bool
	err = tx.QueryRow(ctx,
		`SELECT estado, usuario_id, total, comisiones_calculadas FROM pedidos WHERE id = $1 FOR UPDATE`, id,
	).Scan(&actual, &usuarioID, &total, &comisionesCalculadas)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoEncontrado
	}
	if err != nil {
		return err
	}

	if !transicionPermitida(actual, nuevo) {
		return &ErrNegocio{
			Mensaje: fmt.Sprintf("Transición no permitida: de %s a %s", actual, nuevo),
			Detalle: map[string]string{"estado_actual": actual, "estado_nuevo": nuevo},
		}
	}

	if _, err := tx.Exec(ctx, `UPDATE pedidos SET estado = $2 WHERE id = $1`, id, nuevo); err != nil {
		return err
	}

	if nuevo == "cancelado" {
		// 1) Devolver el stock de cada item.
		if _, err := tx.Exec(ctx,
			`UPDATE productos p SET stock = p.stock + x.cant
			   FROM (SELECT producto_id, SUM(cantidad) AS cant FROM pedido_items WHERE pedido_id = $1 GROUP BY producto_id) x
			  WHERE p.id = x.producto_id`, id); err != nil {
			return err
		}

		// 2) Revertir las comisiones que todavía no se pagaron (las 'pagada' ya salieron en una liquidación).
		rows, err := tx.Query(ctx,
			`UPDATE comisiones SET estado = 'revertida'
			  WHERE pedido_id = $1 AND estado IN ('pendiente', 'aprobada')
			  RETURNING referido_id, monto`, id)
		if err != nil {
			return err
		}
		type reversa struct {
			referidoID int
			monto      float64
		}
		var reversas []reversa
		for rows.Next() {
			var rv reversa
			if err := rows.Scan(&rv.referidoID, &rv.monto); err != nil {
				rows.Close()
				return err
			}
			reversas = append(reversas, rv)
		}
		rows.Close() // hay que cerrar el cursor antes de usar la misma transacción otra vez
		if err := rows.Err(); err != nil {
			return err
		}
		// El saldo solo se resta si la goroutine ya lo había sumado (si no, nunca se sumó y no hay nada que restar).
		if comisionesCalculadas {
			for _, rv := range reversas {
				if _, err := tx.Exec(ctx, `UPDATE referidos SET saldo = GREATEST(saldo - $1, 0) WHERE id = $2`, rv.monto, rv.referidoID); err != nil {
					return err
				}
			}
		}

		// 3) La venta ya no cuenta en las "ventas" del comprador (solo si las comisiones ya se habían calculado).
		if comisionesCalculadas {
			if _, err := tx.Exec(ctx,
				`UPDATE referidos SET ventas = GREATEST(ventas - $1, 0)
				  WHERE id = (SELECT referido_id FROM usuarios WHERE id = $2)`, total, usuarioID); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

// alfabeto de los códigos de referido: sin 0/O ni 1/I para que no se confundan al dictarlos.
const alfabetoCodigo = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generarCodigo() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = alfabetoCodigo[int(b[i])%len(alfabetoCodigo)]
	}
	return string(b)
}

// RegistrarUsuario crea la cuenta y su nodo en la red en UNA transacción:
//  1. busca al referente (por código; sin código, el nodo raíz)
//  2. INSERT en usuarios (rol cliente)
//  3. INSERT en referidos (nivel = nivel del referente + 1, parent_id = referente, código nuevo)
//  4. enlaza usuarios.referido_id con el nodo nuevo
// Si algo falla, ROLLBACK: nunca queda un usuario sin nodo ni un nodo sin usuario.
func (r *PostgresRepository) RegistrarUsuario(ctx context.Context, nombre, email, passwordHash, refCodigo string) (*models.Usuario, string, error) {
	ctx, cancel := conTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	// 1) Referente
	var padreID, padreNivel int
	if refCodigo == "" {
		err = tx.QueryRow(ctx, `SELECT id, nivel FROM referidos WHERE parent_id IS NULL ORDER BY id LIMIT 1`).Scan(&padreID, &padreNivel)
	} else {
		err = tx.QueryRow(ctx, `SELECT id, nivel FROM referidos WHERE codigo = $1`, refCodigo).Scan(&padreID, &padreNivel)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		if refCodigo != "" {
			return nil, "", ErrReferidoInvalido
		}
		return nil, "", errors.New("la red no tiene nodo raíz")
	}
	if err != nil {
		return nil, "", err
	}

	// 2) Usuario
	u := &models.Usuario{Email: email, Nombre: nombre, Rol: "cliente", Activo: true}
	err = tx.QueryRow(ctx,
		`INSERT INTO usuarios (email, password, rol, nombre) VALUES ($1, $2, 'cliente', $3) RETURNING id`,
		email, passwordHash, nombre).Scan(&u.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, "", ErrEmailDuplicado
		}
		return nil, "", err
	}

	// 3) Código único para su link de invitación
	codigo := ""
	for intento := 0; intento < 5 && codigo == ""; intento++ {
		c := generarCodigo()
		var existe bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM referidos WHERE codigo = $1)`, c).Scan(&existe); err != nil {
			return nil, "", err
		}
		if !existe {
			codigo = c
		}
	}
	if codigo == "" {
		return nil, "", errors.New("no se pudo generar un código de referido único")
	}

	// 4) Nodo en la red y vínculo con el usuario
	var nodoID int
	if err := tx.QueryRow(ctx,
		`INSERT INTO referidos (nombre, nivel, parent_id, usuario_id, codigo) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		nombre, padreNivel+1, padreID, u.ID, codigo).Scan(&nodoID); err != nil {
		return nil, "", fmt.Errorf("no se pudo crear el nodo de la red: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE usuarios SET referido_id = $1 WHERE id = $2`, nodoID, u.ID); err != nil {
		return nil, "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, "", err
	}
	return u, codigo, nil
}

// ---------- RED MULTINIVEL ----------

// ObtenerRed lee la tabla plana referidos y reconstruye el árbol multinivel en memoria.
func (r *PostgresRepository) ObtenerRed(ctx context.Context) (*models.Referido, error) {
	ctx, cancel := conTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.Query(ctx,
		`SELECT id, nombre, nivel, ventas, saldo, parent_id FROM referidos ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodos := map[int]*models.Referido{}
	padres := map[int]int{}
	orden := []int{}

	for rows.Next() {
		var n models.Referido
		var parent *int // puntero: acepta NULL de la base de datos
		if err := rows.Scan(&n.ID, &n.Nombre, &n.Nivel, &n.Ventas, &n.Saldo, &parent); err != nil {
			return nil, err
		}
		nodos[n.ID] = &n
		orden = append(orden, n.ID)
		if parent != nil {
			padres[n.ID] = *parent
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var raiz *models.Referido
	for _, id := range orden {
		padreID, tienePadre := padres[id]
		if !tienePadre {
			raiz = nodos[id]
			continue
		}
		if padre, ok := nodos[padreID]; ok {
			padre.Hijos = append(padre.Hijos, nodos[id])
		}
	}
	if raiz == nil {
		return nil, ErrNoEncontrado
	}
	return raiz, nil
}

// ---------- CHECKOUT: TRANSACCIÓN ACID ----------

// GuardarPedido guarda el pedido completo o no guarda nada:
//
//	BEGIN
//	  1. INSERT en pedidos
//	  2. por cada item: descuenta stock + INSERT en pedido_items
//	  3. calcula el total
//	COMMIT      (si cualquier paso falla -> ROLLBACK automático por el defer)
func (r *PostgresRepository) GuardarPedido(ctx context.Context, req models.CheckoutRequest, usuario *models.Usuario) (*models.Pedido, error) {
	// Una cantidad <= 0 se rechaza ANTES de abrir la transacción (sumaría stock en vez de restarlo).
	for _, it := range req.Items {
		if it.Cantidad <= 0 {
			return nil, ErrCantidadInvalida
		}
	}

	ctx, cancel := conTimeout(ctx, 60*time.Second)
	defer cancel()

	// BEGIN
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("no se pudo iniciar la transacción: %w", err)
	}
	// Si salimos por cualquier error sin llegar al Commit, esto hace ROLLBACK.
	// Después de un Commit exitoso no hace nada (devuelve ErrTxClosed, que ignoramos).
	defer func() { _ = tx.Rollback(context.Background()) }()

	// Paso 1: registro principal del pedido (total en 0; se calcula con los items).
	pedido := &models.Pedido{Estado: "pendiente", Cliente: req.Cliente.Nombre, Items: []models.PedidoItem{}}
	detalle := req.PagoDetalle
	if detalle == nil {
		detalle = map[string]string{}
	}
	detalleJSON, err := json.Marshal(detalle)
	if err != nil {
		return nil, fmt.Errorf("no se pudo preparar el detalle del pago: %w", err)
	}
	var fecha time.Time
	err = tx.QueryRow(ctx,
		`INSERT INTO pedidos (usuario_id, cliente_nombre, cliente_email, direccion, ciudad, telefono, metodo_pago, total, estado, pago_detalle)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 0, 'pendiente', $8::jsonb)
		 RETURNING id, fecha`,
		usuario.ID, req.Cliente.Nombre, req.Cliente.Email, req.Cliente.Direccion,
		req.Cliente.Ciudad, req.Cliente.Telefono, req.MetodoPago, string(detalleJSON),
	).Scan(&pedido.ID, &fecha)
	if err != nil {
		return nil, fmt.Errorf("no se pudo insertar el pedido: %w", err)
	}

	// >>> PUNTO DE FALLA PARA LA CLASE (SIMULAR_FALLA_CHECKOUT=true): el pedido ya se insertó,
	// pero los items no. El defer hace ROLLBACK y en la BD no queda ningún "pedido fantasma".
	if r.opts.SimularFalla {
		return nil, errors.New("falla simulada: caída del sistema antes de procesar los items")
	}

	// Unimos líneas repetidas del mismo producto y las ordenamos por id: así dos compras simultáneas
	// bloquean las filas siempre en el mismo orden y no se produce un deadlock.
	cantidades := map[int]int{}
	for _, it := range req.Items {
		cantidades[it.ID] += it.Cantidad
	}
	ids := make([]int, 0, len(cantidades))
	for id := range cantidades {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	// Paso 2: detalle del pedido restando stock.
	for _, id := range ids {
		cant := cantidades[id]
		var nombre string
		var precio float64

		// Resta stock solo si alcanza, y devuelve nombre y precio REALES (nunca los del cliente).
		err := tx.QueryRow(ctx,
			`UPDATE productos SET stock = stock - $2 WHERE id = $1 AND activo AND stock >= $2 RETURNING nombre, precio`,
			id, cant,
		).Scan(&nombre, &precio)
		if errors.Is(err, pgx.ErrNoRows) {
			// No se actualizó nada: o el producto no existe o no hay stock suficiente.
			var stock int
			errStock := tx.QueryRow(ctx, `SELECT stock FROM productos WHERE id = $1`, id).Scan(&stock)
			if errors.Is(errStock, pgx.ErrNoRows) {
				return nil, &ErrNegocio{Mensaje: fmt.Sprintf("El producto %d no existe", id), Detalle: map[string]int{"producto_id": id}}
			}
			if errStock != nil {
				return nil, errStock
			}
			return nil, &ErrNegocio{
				Mensaje: fmt.Sprintf("Stock insuficiente para el producto %d", id),
				Detalle: map[string]int{"producto_id": id, "solicitado": cant, "disponible": stock},
			}
		}
		if err != nil {
			return nil, fmt.Errorf("no se pudo descontar el stock del producto %d: %w", id, err)
		}

		if _, err := tx.Exec(ctx,
			`INSERT INTO pedido_items (pedido_id, producto_id, cantidad, precio_unitario) VALUES ($1, $2, $3, $4)`,
			pedido.ID, id, cant, precio,
		); err != nil {
			return nil, fmt.Errorf("no se pudo insertar el item del producto %d: %w", id, err)
		}
		pedido.Items = append(pedido.Items, models.PedidoItem{ProductoID: id, Nombre: nombre, Cantidad: cant, PrecioUnitario: precio})
	}

	// Paso 3: el total lo calcula PostgreSQL con los precios guardados (evita errores de decimales).
	if err := tx.QueryRow(ctx,
		`UPDATE pedidos SET total = (SELECT COALESCE(SUM(cantidad * precio_unitario), 0) FROM pedido_items WHERE pedido_id = $1)
		 WHERE id = $1 RETURNING total`,
		pedido.ID,
	).Scan(&pedido.Total); err != nil {
		return nil, fmt.Errorf("no se pudo calcular el total: %w", err)
	}

	// Ledger de comisiones: NACE con el pedido, dentro de la MISMA transacción. Si algo falla, no queda
	// ni el pedido ni sus comisiones. La goroutine posterior solo actualiza saldos y ventas (contadores).
	if err := registrarComisiones(ctx, tx, pedido.ID, usuario.ID, pedido.Total); err != nil {
		return nil, fmt.Errorf("no se pudo registrar el ledger de comisiones: %w", err)
	}

	// >>> PAUSA PARA LA CLASE (CHECKOUT_PAUSA_SEGUNDOS=10): da tiempo de hacer "docker kill catalogo_postgres"
	// mientras la transacción está abierta y todavía sin COMMIT.
	if r.opts.PausaAntesCommit > 0 {
		select {
		case <-time.After(r.opts.PausaAntesCommit):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// COMMIT: recién aquí el pedido, los items y el stock se vuelven permanentes.
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("no se pudo confirmar la transacción: %w", err)
	}

	pedido.Numero = numeroPedido(pedido.ID)
	pedido.Fecha = fecha.Format(time.RFC3339)
	return pedido, nil
}

// ---------- COMISIONES: ledger + contadores ----------

type consultador interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type ascendente struct{ id, nivel int }

// ascendentes devuelve hasta maxNivel ascendentes de un nodo, con una consulta recursiva (CTE) sobre parent_id.
func ascendentes(ctx context.Context, q consultador, nodoID, maxNivel int) ([]ascendente, error) {
	rows, err := q.Query(ctx, `
		WITH RECURSIVE ascendentes AS (
			SELECT parent_id AS id, 1 AS nivel FROM referidos WHERE id = $1 AND parent_id IS NOT NULL
			UNION ALL
			SELECT r.parent_id, a.nivel + 1
			FROM referidos r JOIN ascendentes a ON r.id = a.id
			WHERE r.parent_id IS NOT NULL AND a.nivel < $2
		)
		SELECT id, nivel FROM ascendentes ORDER BY nivel`, nodoID, maxNivel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []ascendente
	for rows.Next() {
		var a ascendente
		if err := rows.Scan(&a.id, &a.nivel); err != nil {
			return nil, err
		}
		lista = append(lista, a)
	}
	return lista, rows.Err()
}

// registrarComisiones escribe el LEDGER (tabla comisiones, estado "pendiente") dentro de la transacción del pedido.
// El reparto lo calcula la función pura domain.RepartirComisiones. Si el comprador no tiene nodo en la red, no hay comisiones.
func registrarComisiones(ctx context.Context, tx pgx.Tx, pedidoID, usuarioID int, total float64) error {
	var nodo *int
	if err := tx.QueryRow(ctx, `SELECT referido_id FROM usuarios WHERE id = $1`, usuarioID).Scan(&nodo); err != nil {
		return err
	}
	if nodo == nil {
		return nil
	}

	reparto := domain.RepartirComisiones(total, domain.TasasComision)
	asc, err := ascendentes(ctx, tx, *nodo, len(reparto))
	if err != nil {
		return err
	}
	for _, a := range asc {
		if a.nivel < 1 || a.nivel > len(reparto) {
			continue
		}
		c := reparto[a.nivel-1]
		// ON CONFLICT + UNIQUE(pedido, referido, nivel): una comisión nunca se registra dos veces.
		if _, err := tx.Exec(ctx,
			`INSERT INTO comisiones (pedido_id, referido_id, nivel, porcentaje, monto) VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (pedido_id, referido_id, nivel) DO NOTHING`,
			pedidoID, a.id, c.Nivel, c.Porcentaje, c.Monto); err != nil {
			return err
		}
	}
	return nil
}

// AplicarComisiones (se ejecuta en una Goroutine DESPUÉS del COMMIT del pedido) actualiza los CONTADORES:
// las ventas del comprador y el saldo de cada ascendente, a partir de lo que ya está en el ledger.
// Es idempotente (la marca comisiones_calculadas evita sumar dos veces) y no hace nada si el pedido se canceló.
func (r *PostgresRepository) AplicarComisiones(ctx context.Context, pedidoID, usuarioID int, total float64) error {
	ctx, cancel := conTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	var estado string
	var yaCalculadas bool
	if err := tx.QueryRow(ctx, `SELECT estado, comisiones_calculadas FROM pedidos WHERE id = $1 FOR UPDATE`, pedidoID).Scan(&estado, &yaCalculadas); err != nil {
		return err
	}
	if estado == "cancelado" || yaCalculadas {
		return nil // cancelado antes de correr, o ya aplicado: no hay nada que hacer (ROLLBACK por el defer)
	}

	// La venta cuenta en las "ventas" del comprador (si tiene nodo en la red).
	if _, err := tx.Exec(ctx,
		`UPDATE referidos SET ventas = ventas + $1 WHERE id = (SELECT referido_id FROM usuarios WHERE id = $2)`,
		total, usuarioID); err != nil {
		return err
	}
	// El saldo de cada ascendente sube lo que dice el ledger de ESTE pedido.
	if _, err := tx.Exec(ctx,
		`UPDATE referidos r SET saldo = r.saldo + c.monto
		   FROM comisiones c
		  WHERE c.pedido_id = $1 AND c.estado = 'pendiente' AND c.referido_id = r.id`, pedidoID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE pedidos SET comisiones_calculadas = TRUE WHERE id = $1`, pedidoID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ---------- RED PROPIA Y ESTADO DE CUENTA ----------

func buscarNodo(n *models.Referido, id int) *models.Referido {
	if n.ID == id {
		return n
	}
	for _, h := range n.Hijos {
		if encontrado := buscarNodo(h, id); encontrado != nil {
			return encontrado
		}
	}
	return nil
}

// ajustarNiveles vuelve relativos los niveles: el nodo del usuario pasa a 0 y sus directos a 1.
func ajustarNiveles(n *models.Referido, base int) {
	n.Nivel -= base
	for _, h := range n.Hijos {
		ajustarNiveles(h, base)
	}
}

// ObtenerRedUsuario devuelve el subárbol desde el nodo del usuario (NO el árbol global) y su código de invitación.
func (r *PostgresRepository) ObtenerRedUsuario(ctx context.Context, usuarioID int) (*models.RedUsuario, error) {
	ctx, cancel := conTimeout(ctx, 5*time.Second)
	defer cancel()

	var nodoID, nodoNivel int
	var codigo string
	err := r.db.QueryRow(ctx,
		`SELECT r.id, r.nivel, COALESCE(r.codigo, '')
		   FROM usuarios u JOIN referidos r ON r.id = u.referido_id WHERE u.id = $1`, usuarioID,
	).Scan(&nodoID, &nodoNivel, &codigo)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, err
	}

	raiz, err := r.ObtenerRed(ctx)
	if err != nil {
		return nil, err
	}
	nodo := buscarNodo(raiz, nodoID)
	if nodo == nil {
		return nil, ErrNoEncontrado
	}
	ajustarNiveles(nodo, nodoNivel)
	return &models.RedUsuario{Codigo: codigo, Red: nodo}, nil
}

// ComisionesMias devuelve el ledger del usuario (paginado) y el total por estado.
func (r *PostgresRepository) ComisionesMias(ctx context.Context, usuarioID, page, limit int) (*models.ComisionesMias, error) {
	ctx, cancel := conTimeout(ctx, 5*time.Second)
	defer cancel()

	res := &models.ComisionesMias{
		Data: []models.ComisionLedger{}, Page: page, Limit: limit,
		Totales: map[string]float64{"pendiente": 0, "aprobada": 0, "pagada": 0, "revertida": 0},
	}

	var nodo *int
	if err := r.db.QueryRow(ctx, `SELECT referido_id FROM usuarios WHERE id = $1`, usuarioID).Scan(&nodo); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoEncontrado
		}
		return nil, err
	}
	if nodo == nil {
		return res, nil
	}

	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM comisiones WHERE referido_id = $1`, *nodo).Scan(&res.Total); err != nil {
		return nil, err
	}

	filas, err := r.db.Query(ctx,
		`SELECT estado, COALESCE(SUM(monto), 0) FROM comisiones WHERE referido_id = $1 GROUP BY estado`, *nodo)
	if err != nil {
		return nil, err
	}
	for filas.Next() {
		var estado string
		var suma float64
		if err := filas.Scan(&estado, &suma); err != nil {
			filas.Close()
			return nil, err
		}
		res.Totales[estado] = suma
	}
	filas.Close()
	if err := filas.Err(); err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, pedido_id, nivel, COALESCE(porcentaje, 0), monto, estado, fecha
		   FROM comisiones WHERE referido_id = $1 ORDER BY fecha DESC, id DESC LIMIT $2 OFFSET $3`,
		*nodo, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var c models.ComisionLedger
		var fecha time.Time
		if err := rows.Scan(&c.ID, &c.PedidoID, &c.Nivel, &c.Porcentaje, &c.Monto, &c.Estado, &fecha); err != nil {
			return nil, err
		}
		c.Numero = numeroPedido(c.PedidoID)
		c.Fecha = fecha.Format(time.RFC3339)
		res.Data = append(res.Data, c)
	}
	return res, rows.Err()
}

// ---------- LIQUIDACIONES ----------

// Liquidar paga TODAS las comisiones pendientes de un distribuidor en una transacción:
// bloquea sus filas (FOR UPDATE), crea la liquidación con el total y las marca "pagada".
// Si dos admins liquidan a la vez, el segundo ve 0 pendientes y recibe un error: nunca se paga dos veces.
func (r *PostgresRepository) Liquidar(ctx context.Context, referidoID int) (*models.Liquidacion, error) {
	ctx, cancel := conTimeout(ctx, 15*time.Second)
	defer cancel()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	var nombre string
	err = tx.QueryRow(ctx, `SELECT nombre FROM referidos WHERE id = $1`, referidoID).Scan(&nombre)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx,
		`SELECT id, monto FROM comisiones WHERE referido_id = $1 AND estado = 'pendiente' ORDER BY id FOR UPDATE`, referidoID)
	if err != nil {
		return nil, err
	}
	ids := []int{}
	var centavos int64
	for rows.Next() {
		var id int
		var monto float64
		if err := rows.Scan(&id, &monto); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
		centavos += int64(math.Round(monto * 100)) // se suma en centavos para no acumular errores de decimales
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, &ErrNegocio{Mensaje: "No hay comisiones pendientes para liquidar", Detalle: map[string]int{"referido_id": referidoID}}
	}

	liq := &models.Liquidacion{ReferidoID: referidoID, Referido: nombre, Total: float64(centavos) / 100, Estado: "pagada"}
	var creado, pagada time.Time
	if err := tx.QueryRow(ctx,
		`INSERT INTO liquidaciones (referido_id, total, estado, pagada_en) VALUES ($1, $2, 'pagada', NOW())
		 RETURNING id, creado_en, pagada_en`, referidoID, liq.Total,
	).Scan(&liq.ID, &creado, &pagada); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx,
		`UPDATE comisiones SET estado = 'pagada', liquidacion_id = $1 WHERE id = ANY($2)`, liq.ID, ids); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	liq.CreadoEn = creado.Format(time.RFC3339)
	liq.PagadaEn = pagada.Format(time.RFC3339)
	return liq, nil
}

// ListarLiquidaciones: usuarioID 0 = todas (admin); si no, solo las del nodo de ese usuario.
func (r *PostgresRepository) ListarLiquidaciones(ctx context.Context, usuarioID, page, limit int) ([]models.Liquidacion, int, error) {
	ctx, cancel := conTimeout(ctx, 5*time.Second)
	defer cancel()

	const filtro = `($1 = 0 OR l.referido_id = (SELECT referido_id FROM usuarios WHERE id = $1))`
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM liquidaciones l WHERE `+filtro, usuarioID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx,
		`SELECT l.id, l.referido_id, r.nombre, l.total, l.estado, l.creado_en, l.pagada_en
		   FROM liquidaciones l JOIN referidos r ON r.id = l.referido_id
		  WHERE `+filtro+` ORDER BY l.creado_en DESC, l.id DESC LIMIT $2 OFFSET $3`,
		usuarioID, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	lista := []models.Liquidacion{}
	for rows.Next() {
		var l models.Liquidacion
		var creado time.Time
		var pagada *time.Time
		if err := rows.Scan(&l.ID, &l.ReferidoID, &l.Referido, &l.Total, &l.Estado, &creado, &pagada); err != nil {
			return nil, 0, err
		}
		l.CreadoEn = creado.Format(time.RFC3339)
		if pagada != nil {
			l.PagadaEn = pagada.Format(time.RFC3339)
		}
		lista = append(lista, l)
	}
	return lista, total, rows.Err()
}

// DeudaPorReferido: cuánto se le debe a cada distribuidor (solo los que tienen comisiones pendientes).
func (r *PostgresRepository) DeudaPorReferido(ctx context.Context) ([]models.DeudaReferido, error) {
	ctx, cancel := conTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.Query(ctx,
		`SELECT r.id, r.nombre, SUM(c.monto), COUNT(c.id)
		   FROM referidos r JOIN comisiones c ON c.referido_id = r.id AND c.estado = 'pendiente'
		  GROUP BY r.id, r.nombre ORDER BY SUM(c.monto) DESC, r.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lista := []models.DeudaReferido{}
	for rows.Next() {
		var d models.DeudaReferido
		if err := rows.Scan(&d.ReferidoID, &d.Nombre, &d.Pendiente, &d.Comisiones); err != nil {
			return nil, err
		}
		lista = append(lista, d)
	}
	return lista, rows.Err()
}

// NormalizarEmail deja el email en minúsculas y sin espacios (se usa también para las claves de caché).
func NormalizarEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
