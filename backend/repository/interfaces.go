// Paquete repository: capa de acceso a datos. Los controladores dependen de estas INTERFACES,
// no de PostgreSQL ni de Redis directamente (Inyección de Dependencias).
package repository

import (
	"context"
	"errors"
	"time"

	"multicatalogo-backend/models"
)

// ErrNoEncontrado se devuelve cuando la consulta no encuentra ninguna fila.
var ErrNoEncontrado = errors.New("registro no encontrado")

// ErrCantidadInvalida: la cantidad de un item es cero o negativa (400). Es distinto del error de stock.
var ErrCantidadInvalida = &ErrNegocio{Mensaje: "cantidad debe ser mayor que cero"}

// Errores del registro de usuarios: el controlador los traduce a 409 y 400.
var (
	ErrEmailDuplicado   = errors.New("el email ya está registrado")
	ErrReferidoInvalido = errors.New("el código de referido no existe")
)

// ErrNegocio es un error de regla de negocio o de datos del cliente (stock insuficiente, producto
// inexistente...). El controlador lo traduce a HTTP 400; cualquier otro error se traduce a HTTP 500.
type ErrNegocio struct {
	Mensaje string
	Detalle interface{}
}

func (e *ErrNegocio) Error() string { return e.Mensaje }

// Estados válidos de un pedido.
var EstadosPedido = []string{"pendiente", "preparando", "enviado", "entregado", "cancelado"}

// FiltroPedidos acota un listado de pedidos. UsuarioID 0 = de todos los usuarios; Estado "" = todos.
type FiltroPedidos struct {
	UsuarioID int
	Estado    string
	Page      int
	Limit     int
}

// DBRepository define TODO lo que la aplicación necesita de la base de datos relacional.
type DBRepository interface {
	// BuscarUsuarioPorEmail devuelve el usuario con su hash bcrypt en Password (el login lo compara).
	BuscarUsuarioPorEmail(ctx context.Context, email string) (*models.Usuario, error)

	// RegistrarUsuario crea el usuario (rol cliente) y su nodo en la red EN UNA TRANSACCIÓN.
	// Devuelve el usuario y su código de referido. Con refCodigo vacío cuelga del nodo raíz.
	// Errores: ErrEmailDuplicado (409) y ErrReferidoInvalido (400).
	RegistrarUsuario(ctx context.Context, nombre, email, passwordHash, refCodigo string) (*models.Usuario, string, error)

	// Catálogo público: solo productos activos.
	ObtenerProductos(ctx context.Context) ([]models.Producto, error)
	ObtenerProductoPorID(ctx context.Context, id int) (*models.Producto, error)

	// Administración de productos (solo admin). El borrado es LÓGICO (activo=false) para no romper pedidos antiguos.
	ListarProductosAdmin(ctx context.Context, q string, page, limit int) ([]models.Producto, int, error)
	CrearProducto(ctx context.Context, p models.ProductoEntrada) (*models.Producto, error)
	ActualizarProducto(ctx context.Context, id int, cambios models.ProductoCambios) (*models.Producto, error)
	DesactivarProducto(ctx context.Context, id int) error

	// Red y comisiones del usuario con sesión.
	ObtenerRedUsuario(ctx context.Context, usuarioID int) (*models.RedUsuario, error)
	ComisionesMias(ctx context.Context, usuarioID, page, limit int) (*models.ComisionesMias, error)

	// Liquidaciones: el admin paga lo que el ledger debe. Liquidar es una transacción; si no hay nada
	// pendiente devuelve *ErrNegocio (400), por eso no se puede liquidar dos veces lo mismo.
	Liquidar(ctx context.Context, referidoID int) (*models.Liquidacion, error)
	ListarLiquidaciones(ctx context.Context, usuarioID, page, limit int) ([]models.Liquidacion, int, error)
	DeudaPorReferido(ctx context.Context) ([]models.DeudaReferido, error)

	// Fotos de productos: se guardan en la base de datos (el disco de Render no es permanente).
	GuardarImagen(ctx context.Context, tipo string, datos []byte) (int, error)
	ObtenerImagen(ctx context.Context, id int) (tipo string, datos []byte, err error)

	// Comprobantes (vouchers) de transferencia: privados, solo el admin los consulta.
	GuardarComprobante(ctx context.Context, usuarioID int, tipo string, datos []byte) (int, error)
	ObtenerComprobante(ctx context.Context, id int) (tipo string, datos []byte, err error)

	// Consulta y gestión de pedidos.
	ListarPedidos(ctx context.Context, f FiltroPedidos) ([]models.PedidoListado, int, error)
	ObtenerPedido(ctx context.Context, id int) (*models.PedidoDetalle, error)
	// CambiarEstadoPedido valida la transición. Al CANCELAR (en la misma transacción) devuelve el stock
	// y revierte las comisiones aún no liquidadas. Errores: ErrNoEncontrado (404) y *ErrNegocio (400).
	CambiarEstadoPedido(ctx context.Context, id int, nuevo string) error

	ObtenerRed(ctx context.Context) (*models.Referido, error)

	// Reportes (solo admin). desde/hasta son el intervalo [desde, hasta) ya calculado (nil = sin límite).
	ResumenVentas(ctx context.Context, desde, hasta *time.Time) (*models.ReporteResumen, error)
	TopProductos(ctx context.Context, desde, hasta *time.Time, limite int) ([]models.TopProducto, error)
	TopReferidos(ctx context.Context, desde, hasta *time.Time, limite int) ([]models.TopReferido, error)

	// Administración de usuarios (solo admin): listado paginado con búsqueda ?q= y cambio de rol/estado.
	ListarUsuarios(ctx context.Context, q string, page, limit int) ([]models.UsuarioListado, int, error)
	ActualizarUsuario(ctx context.Context, id int, c models.UsuarioCambios) (*models.UsuarioListado, error)

	// GuardarPedido ejecuta toda la compra en UNA transacción ACID (BEGIN ... COMMIT / ROLLBACK).
	GuardarPedido(ctx context.Context, req models.CheckoutRequest, usuario *models.Usuario) (*models.Pedido, error)
	// AplicarComisiones reparte las comisiones a los 3 niveles ascendentes del comprador.
	AplicarComisiones(ctx context.Context, pedidoID, usuarioID int, total float64) error

	// Ping comprueba que PostgreSQL responde (lo usa /api/health).
	Ping(ctx context.Context) error
}

// CacheRepository define las operaciones de caché que usa la aplicación (patrón Cache-Aside).
// Se implementa con Redis, pero los controladores no lo saben.
type CacheRepository interface {
	// Get devuelve el valor guardado; el bool indica si existía (hit) o no (miss).
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	// DeleteByPrefix borra todas las claves que empiecen con el prefijo (ej. "red:tree:").
	DeleteByPrefix(ctx context.Context, prefix string) error
	// Ping comprueba que la caché responde (lo usa /api/health).
	Ping(ctx context.Context) error
}
