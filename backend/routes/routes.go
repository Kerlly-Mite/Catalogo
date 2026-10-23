package routes

import (
	"github.com/gofiber/fiber/v2"

	"multicatalogo-backend/controllers"
	"multicatalogo-backend/middleware"
)

// Dependencias agrupa los controladores que main.go crea e inyecta.
type Dependencias struct {
	Auth     *controllers.AuthController
	Producto *controllers.ProductoController
	Red      *controllers.RedController
	Checkout *controllers.CheckoutController
	Health   *controllers.HealthController
	Pedido   *controllers.PedidoController

	Liquidacion *controllers.LiquidacionController
	Imagen      *controllers.ImagenController
	Comprobante *controllers.ComprobanteController
	Reporte     *controllers.ReporteController
	Usuario     *controllers.UsuarioController

	// JWTSecret es el secreto con el que se validan los tokens en las rutas protegidas.
	JWTSecret string
}

// SetupRoutes registra todos los endpoints de la API.
func SetupRoutes(app *fiber.App, d Dependencias) {
	api := app.Group("/api")

	// Seguridad por defecto: cada ruta protegida lleva su middleware explícito.
	autenticado := middleware.RequireAuth(d.JWTSecret)
	soloAdmin := middleware.RequireAdmin()

	// ---- Públicas ----
	api.Get("/health", d.Health.Health)
	api.Post("/login", d.Auth.Login)
	api.Post("/register", d.Auth.Register)
	api.Get("/imagenes/:id", d.Imagen.Ver) // fotos de productos (públicas, como el catálogo)
	api.Get("/productos", d.Producto.GetProductos)
	api.Get("/productos/:id", d.Producto.GetProductoPorID)

	// ---- Requieren sesión (401 sin token válido) ----
	// El árbol GLOBAL de la red es información del negocio: solo admin (cada usuario ve lo suyo en /red/mia).
	api.Get("/red", autenticado, soloAdmin, d.Red.GetRed)
	api.Post("/checkout", autenticado, d.Checkout.Checkout)
	api.Post("/comprobantes", autenticado, d.Comprobante.Subir)             // voucher de transferencia (cualquier usuario con sesión)
	api.Get("/comprobantes/:id", autenticado, soloAdmin, d.Comprobante.Ver) // verlo: solo admin

	// ---- Red propia, ledger de comisiones y liquidaciones ----
	api.Get("/red/mia", autenticado, d.Red.Mia)
	api.Get("/comisiones/mias", autenticado, d.Red.ComisionesMias)
	api.Get("/liquidaciones/mias", autenticado, d.Liquidacion.Mias)
	api.Get("/liquidaciones/deuda", autenticado, soloAdmin, d.Liquidacion.Deuda)
	api.Get("/liquidaciones", autenticado, soloAdmin, d.Liquidacion.Listar)
	api.Post("/liquidaciones", autenticado, soloAdmin, d.Liquidacion.Liquidar)

	// ---- Pedidos (Rutas fijas ANTES de /:id) ----
	api.Get("/pedidos/mios", autenticado, d.Pedido.Mios)
	api.Get("/pedidos", autenticado, soloAdmin, d.Pedido.Listar)
	api.Get("/pedidos/:id", autenticado, d.Pedido.Detalle)
	api.Patch("/pedidos/:id/estado", autenticado, soloAdmin, d.Pedido.CambiarEstado)

	// ---- Administración de productos (RequireAuth + RequireAdmin) ----
	api.Post("/imagenes", autenticado, soloAdmin, d.Imagen.Subir)
	api.Post("/productos", autenticado, soloAdmin, d.Producto.CrearProducto)
	api.Put("/productos/:id", autenticado, soloAdmin, d.Producto.ActualizarProducto)
	api.Delete("/productos/:id", autenticado, soloAdmin, d.Producto.EliminarProducto)
	api.Get("/admin/productos", autenticado, soloAdmin, d.Producto.ListarAdmin)

	// ---- Reportes (Fase 6, solo admin) ----
	api.Get("/reportes/resumen", autenticado, soloAdmin, d.Reporte.Resumen)
	api.Get("/reportes/top-productos", autenticado, soloAdmin, d.Reporte.TopProductos)
	api.Get("/reportes/top-referidos", autenticado, soloAdmin, d.Reporte.TopReferidos)
	api.Get("/reportes/comisiones-pendientes", autenticado, soloAdmin, d.Reporte.ComisionesPendientes)

	// ---- Administración de usuarios (Fase 7, solo admin) ----
	api.Get("/usuarios", autenticado, soloAdmin, d.Usuario.Listar)
	api.Patch("/usuarios/:id", autenticado, soloAdmin, d.Usuario.Actualizar)
}
