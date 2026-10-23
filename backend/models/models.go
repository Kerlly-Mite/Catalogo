// Declaramos el paquete models para agrupar las estructuras de datos de la aplicación.
package models

// LoginRequest define la estructura esperada para el cuerpo de la petición (JSON) al iniciar sesión.
type LoginRequest struct {
	// Email representa el correo del usuario; la etiqueta `json:"email"` indica cómo se mapea el JSON a esta variable.
	Email string `json:"email"`
	// Password representa la contraseña del usuario; se extrae del campo "password" del JSON entrante.
	Password string `json:"password"`
}

// Producto define la estructura de los datos de un artículo en nuestro catálogo.
// (Fase 2) Ahora incluye descripción, categoría y galería, que vienen de PostgreSQL.
type Producto struct {
	// ID es el identificador único numérico del producto.
	ID int `json:"id"`
	// Nombre es el nombre en texto del producto.
	Nombre string `json:"nombre"`
	// Descripcion es el detalle del producto.
	Descripcion string `json:"descripcion"`
	// Precio es el costo del producto, almacenado como un número con decimales (float64).
	Precio float64 `json:"precio"`
	// Categoria agrupa el producto (Serum, Crema, Tónico, Mascarilla, Kit).
	Categoria string `json:"categoria"`
	// Img es la URL que apunta a la fotografía principal del producto.
	Img string `json:"img"`
	// Galeria son las imágenes adicionales del producto.
	Galeria []string `json:"galeria"`
	// Stock son las unidades disponibles. Activo=false es el "borrado lógico": no se muestra en el catálogo.
	Stock  int  `json:"stock"`
	Activo bool `json:"activo"`
}

// Usuario representa una fila de la tabla usuarios.
type Usuario struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"` // nunca se envía al cliente (aquí vive el hash bcrypt)
	Rol      string `json:"rol"`
	Nombre   string `json:"nombre"`
	Activo   bool   `json:"activo"`
}

// RegisterRequest es el cuerpo de POST /api/register. Ref (opcional) es el código de referido de quien invita.
type RegisterRequest struct {
	Nombre   string `json:"nombre"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Ref      string `json:"ref"`
}

// Referido representa un nodo del árbol multinivel (tabla referidos).
type Referido struct {
	ID     int     `json:"id"`
	Nombre string  `json:"nombre"`
	Nivel  int     `json:"nivel"`
	Ventas float64 `json:"ventas"`
	// Saldo son las comisiones acumuladas (se actualizan al comprar un referido descendiente).
	Saldo float64     `json:"saldo"`
	Hijos []*Referido `json:"hijos,omitempty"`
}

// APIError estandariza TODOS los mensajes de error HTTP de la API.
type APIError struct {
	// Status es el código HTTP del error (400, 401, 404...).
	Status int `json:"status"`
	// Message es el mensaje legible que describe el error.
	Message string `json:"message"`
	// Details es información adicional opcional (ej. campos inválidos). Se omite si es nil.
	Details interface{} `json:"details,omitempty"`
	// Error repite Message: el frontend actual lee `data.error` en el login,
	// así mantenemos compatibilidad sin tocar React (regla 3 del DRS).
	Error string `json:"error"`
}

// ExponerDetalles controla si los errores incluyen el campo Details. Solo se activa con APP_ENV=dev
// (lo configura main.go): en producción los detalles internos no se filtran al cliente.
var ExponerDetalles bool

// NewAPIError construye un APIError rellenando también el campo de compatibilidad.
func NewAPIError(status int, message string, details interface{}) APIError {
	if !ExponerDetalles {
		details = nil
	}
	return APIError{Status: status, Message: message, Details: details, Error: message}
}

// ===== Checkout (Fase 4) =====

// ClienteCheckout son los datos del formulario de envío de Checkout.tsx.
type ClienteCheckout struct {
	Nombre    string `json:"nombre"`
	Email     string `json:"email"`
	Direccion string `json:"direccion"`
	Ciudad    string `json:"ciudad"`
	Telefono  string `json:"telefono"`
}

// ItemCheckout es una línea del carrito. Solo se confía en id y cantidad: el precio SIEMPRE se lee de la BD.
type ItemCheckout struct {
	ID       int `json:"id"`
	Cantidad int `json:"cantidad"`
}

// CheckoutRequest es el JSON que recibe POST /api/checkout.
type CheckoutRequest struct {
	Cliente    ClienteCheckout `json:"cliente"`
	MetodoPago string          `json:"metodoPago"`
	// PagoDetalle son los datos del método elegido (tarjeta: últimos 4; transferencia: banco, referencia y
	// comprobante; efectivo: quién recibe, franja, con cuánto paga). El controlador los valida y los limpia.
	PagoDetalle map[string]string `json:"pagoDetalle"`
	Items       []ItemCheckout    `json:"items"`
}

// PedidoItem es una línea ya guardada del pedido (con nombre y precio reales de la BD).
type PedidoItem struct {
	ProductoID     int     `json:"id"`
	Nombre         string  `json:"nombre"`
	Cantidad       int     `json:"cantidad"`
	PrecioUnitario float64 `json:"precio"`
}

// Pedido es la respuesta de un checkout exitoso (lo que muestra Confirmacion.tsx).
type Pedido struct {
	ID      int          `json:"pedido_id"`
	Numero  string       `json:"numero"`
	Total   float64      `json:"total"`
	Estado  string       `json:"estado"`
	Fecha   string       `json:"fecha"`
	Cliente string       `json:"cliente"`
	Items   []PedidoItem `json:"items"`
}

// ProductoEntrada es el cuerpo de POST /api/productos (crear).
type ProductoEntrada struct {
	Nombre      string   `json:"nombre"`
	Descripcion string   `json:"descripcion"`
	Precio      float64  `json:"precio"`
	Categoria   string   `json:"categoria"`
	Img         string   `json:"img"`
	Galeria     []string `json:"galeria"`
	Stock       int      `json:"stock"`
}

// ProductoCambios es el cuerpo de PUT /api/productos/:id (edición PARCIAL): solo se cambian los campos
// que vienen en el JSON (los punteros nil o el slice nil significan "no tocar").
type ProductoCambios struct {
	Nombre      *string  `json:"nombre"`
	Descripcion *string  `json:"descripcion"`
	Precio      *float64 `json:"precio"`
	Categoria   *string  `json:"categoria"`
	Img         *string  `json:"img"`
	Galeria     []string `json:"galeria"`
	Stock       *int     `json:"stock"`
	Activo      *bool    `json:"activo"`
}

// Pagina es la respuesta estándar de los listados: {data, page, limit, total}.
type Pagina[T any] struct {
	Data  []T `json:"data"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// PedidoListado es una fila de los listados de pedidos (sin los items).
type PedidoListado struct {
	ID            int     `json:"id"`
	Numero        string  `json:"numero"`
	Fecha         string  `json:"fecha"`
	Total         float64 `json:"total"`
	Estado        string  `json:"estado"`
	ClienteNombre string  `json:"cliente_nombre"`
	ClienteEmail  string  `json:"cliente_email"`
	NumItems      int     `json:"num_items"` // unidades en total
}

// PedidoDetalle es un pedido con sus items (nombre del producto por JOIN) y los datos de envío.
type PedidoDetalle struct {
	PedidoListado
	Direccion   string            `json:"direccion"`
	Ciudad      string            `json:"ciudad"`
	Telefono    string            `json:"telefono"`
	MetodoPago  string            `json:"metodo_pago"`
	PagoDetalle map[string]string `json:"pago_detalle"`
	UsuarioID   int               `json:"-"` // dueño del pedido: se usa para autorizar, no se envía
	Items       []PedidoItem      `json:"items"`
}

// CambioEstado es el cuerpo de PATCH /api/pedidos/:id/estado.
type CambioEstado struct {
	Estado string `json:"estado"`
}

// RedUsuario es la red del usuario con sesión: el subárbol desde SU nodo (los niveles son relativos a él:
// 1 = referido directo) y su código para invitar (/registro?ref=CODIGO).
type RedUsuario struct {
	Codigo string    `json:"codigo_referido"`
	Red    *Referido `json:"red"`
}

// ComisionLedger es una línea del ledger de comisiones (estado de cuenta del distribuidor).
type ComisionLedger struct {
	ID         int     `json:"id"`
	PedidoID   int     `json:"pedido_id"`
	Numero     string  `json:"numero"`
	Nivel      int     `json:"nivel"`
	Porcentaje float64 `json:"porcentaje"`
	Monto      float64 `json:"monto"`
	Estado     string  `json:"estado"` // pendiente | aprobada | pagada | revertida
	Fecha      string  `json:"fecha"`
}

// ComisionesMias es la página del ledger del usuario + el total por estado.
type ComisionesMias struct {
	Data    []ComisionLedger   `json:"data"`
	Page    int                `json:"page"`
	Limit   int                `json:"limit"`
	Total   int                `json:"total"`
	Totales map[string]float64 `json:"totales"`
}

// Liquidacion es un pago del admin a un distribuidor: agrupa sus comisiones pendientes.
type Liquidacion struct {
	ID         int     `json:"id"`
	ReferidoID int     `json:"referido_id"`
	Referido   string  `json:"referido"`
	Total      float64 `json:"total"`
	Estado     string  `json:"estado"`
	CreadoEn   string  `json:"creado_en"`
	PagadaEn   string  `json:"pagada_en"`
}

// DeudaReferido es lo que se le debe a un distribuidor (comisiones pendientes de pago).
type DeudaReferido struct {
	ReferidoID int     `json:"referido_id"`
	Nombre     string  `json:"nombre"`
	Pendiente  float64 `json:"pendiente"`
	Comisiones int     `json:"comisiones"`
}

// LiquidarRequest es el cuerpo de POST /api/liquidaciones.
type LiquidarRequest struct {
	ReferidoID int `json:"referido_id"`
}

// ===== Reportes (Fase 6) =====

// ReporteResumen son los KPIs del negocio en un periodo (GET /api/reportes/resumen). Los pedidos
// cancelados no cuentan como venta; las comisiones revertidas no cuentan como devengadas.
type ReporteResumen struct {
	Desde                string  `json:"desde,omitempty"`
	Hasta                string  `json:"hasta,omitempty"`
	TotalVendido         float64 `json:"total_vendido"`
	Pedidos              int     `json:"pedidos"`
	PedidosCancelados    int     `json:"pedidos_cancelados"`
	TicketPromedio       float64 `json:"ticket_promedio"`
	ComisionesDevengadas float64 `json:"comisiones_devengadas"`
	ComisionesPendientes float64 `json:"comisiones_pendientes"`
	ComisionesPagadas    float64 `json:"comisiones_pagadas"`
	Distribuidores       int     `json:"distribuidores"` // nodos de la red sin contar la raíz
}

// TopProducto es una fila del ranking de productos por ingresos.
type TopProducto struct {
	ID       int     `json:"id"`
	Nombre   string  `json:"nombre"`
	Unidades int     `json:"unidades"`
	Ingresos float64 `json:"ingresos"`
}

// TopReferido es una fila del ranking de distribuidores: lo que su red vendió y lo que él ganó (ledger).
type TopReferido struct {
	ReferidoID int     `json:"referido_id"`
	Nombre     string  `json:"nombre"`
	Pedidos    int     `json:"pedidos"`
	VentasRed  float64 `json:"ventas_red"`
	Comisiones float64 `json:"comisiones"`
}

// ComisionesPendientesReporte es la deuda de comisiones por distribuidor + el total adeudado.
type ComisionesPendientesReporte struct {
	Data           []DeudaReferido `json:"data"`
	TotalPendiente float64         `json:"total_pendiente"`
}

// ===== Usuarios (Fase 7, solo admin) =====

// UsuarioListado es una fila del listado de usuarios (nunca incluye el hash de la contraseña).
type UsuarioListado struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Nombre   string `json:"nombre"`
	Rol      string `json:"rol"`
	Activo   bool   `json:"activo"`
	CreadoEn string `json:"creado_en"`
	Codigo   string `json:"codigo_referido"`
}

// UsuarioCambios es el cuerpo de PATCH /api/usuarios/:id: solo se cambia lo que viene (nil = no tocar).
type UsuarioCambios struct {
	Rol    *string `json:"rol"`
	Activo *bool   `json:"activo"`
}
