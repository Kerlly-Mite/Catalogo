package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"multicatalogo-backend/auth"
	"multicatalogo-backend/controllers"
)

// Auditoría automática de seguridad (Fase 7): TODA ruta registrada debe estar clasificada en una de estas
// tres listas. Si agregas un endpoint nuevo y no lo clasificas aquí, el test falla: ninguna ruta nace sin decidir
// quién puede llamarla. No necesita base de datos: las pruebas se cortan en el middleware antes del controlador.

// Públicas a propósito (sin token).
var rutasPublicas = map[string]bool{
	"GET /api/health":        true,
	"POST /api/login":        true,
	"POST /api/register":     true,
	"GET /api/productos":     true,
	"GET /api/productos/:id": true,
	"GET /api/imagenes/:id":  true,
}

// Cualquier usuario con sesión (RequireAuth). Cada una filtra por la identidad del TOKEN, no por el body/query.
var rutasConSesion = map[string]bool{
	"POST /api/checkout":          true,
	"POST /api/comprobantes":      true, // el voucher queda a nombre del usuario del TOKEN
	"GET /api/red/mia":            true,
	"GET /api/comisiones/mias":    true,
	"GET /api/liquidaciones/mias": true,
	"GET /api/pedidos/mios":       true,
	"GET /api/pedidos/:id":        true, // el controlador devuelve 404 si el pedido no es del usuario (salvo admin)
}

// Solo administrador (RequireAuth + RequireAdmin).
var rutasAdmin = map[string]bool{
	"GET /api/red":                            true,
	"GET /api/pedidos":                        true,
	"PATCH /api/pedidos/:id/estado":           true,
	"POST /api/productos":                     true,
	"POST /api/imagenes":                      true,
	"GET /api/comprobantes/:id":               true, // los vouchers contienen datos bancarios: solo admin
	"PUT /api/productos/:id":                  true,
	"DELETE /api/productos/:id":               true,
	"GET /api/admin/productos":                true,
	"GET /api/liquidaciones":                  true,
	"GET /api/liquidaciones/deuda":            true,
	"POST /api/liquidaciones":                 true,
	"GET /api/reportes/resumen":               true,
	"GET /api/reportes/top-productos":         true,
	"GET /api/reportes/top-referidos":         true,
	"GET /api/reportes/comisiones-pendientes": true,
	"GET /api/usuarios":                       true,
	"PATCH /api/usuarios/:id":                 true,
}

const secretoDePrueba = "secreto-de-prueba-con-mas-de-32-caracteres-123"

func appDePrueba() *fiber.App {
	app := fiber.New()
	// Si una ruta se queda sin protección, su controlador (sin BD aquí) revienta: recover lo convierte en 500
	// y el test informa "se esperaba 401 y llegó 500" en vez de abortar con un panic.
	app.Use(recover.New())
	SetupRoutes(app, Dependencias{
		Auth:        &controllers.AuthController{},
		Producto:    &controllers.ProductoController{},
		Red:         &controllers.RedController{},
		Checkout:    &controllers.CheckoutController{},
		Health:      &controllers.HealthController{},
		Pedido:      &controllers.PedidoController{},
		Liquidacion: &controllers.LiquidacionController{},
		Reporte:     &controllers.ReporteController{},
		Usuario:     &controllers.UsuarioController{},
		JWTSecret:   secretoDePrueba,
	})
	return app
}

// rutasRegistradas devuelve "MÉTODO /ruta" de cada endpoint (ignora HEAD/OPTIONS que Fiber agrega solo).
func rutasRegistradas(app *fiber.App) []string {
	var claves []string
	for _, r := range app.GetRoutes(true) {
		if r.Method == http.MethodHead || r.Method == http.MethodOptions {
			continue
		}
		claves = append(claves, r.Method+" "+r.Path)
	}
	return claves
}

func estado(t *testing.T, app *fiber.App, clave, token string) int {
	t.Helper()
	metodo, ruta, _ := strings.Cut(clave, " ")
	ruta = strings.ReplaceAll(ruta, ":id", "1")
	req := httptest.NewRequest(metodo, ruta, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s: %v", clave, err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func TestTodasLasRutasEstanClasificadas(t *testing.T) {
	app := appDePrueba()
	vistas := map[string]bool{}
	for _, k := range rutasRegistradas(app) {
		vistas[k] = true
		n := 0
		for _, lista := range []map[string]bool{rutasPublicas, rutasConSesion, rutasAdmin} {
			if lista[k] {
				n++
			}
		}
		if n != 1 {
			t.Errorf("la ruta %q debe estar en EXACTAMENTE una lista (pública, con sesión o admin) y está en %d", k, n)
		}
	}
	for _, lista := range []map[string]bool{rutasPublicas, rutasConSesion, rutasAdmin} {
		for k := range lista {
			if !vistas[k] {
				t.Errorf("la lista menciona %q pero esa ruta no está registrada", k)
			}
		}
	}
}

func TestSinTokenSoloLasPublicasResponden(t *testing.T) {
	app := appDePrueba()
	for _, k := range rutasRegistradas(app) {
		if rutasPublicas[k] {
			continue // las públicas llegan al controlador (que aquí no tiene BD)
		}
		if got := estado(t, app, k, ""); got != fiber.StatusUnauthorized {
			t.Errorf("%s sin token: se esperaba 401 y llegó %d", k, got)
		}
	}
}

func TestClienteNoEntraARutasDeAdmin(t *testing.T) {
	app := appDePrueba()
	token, err := auth.GenerarToken(secretoDePrueba, 2, "cliente@upse.edu.ec", "cliente")
	if err != nil {
		t.Fatal(err)
	}
	for k := range rutasAdmin {
		if got := estado(t, app, k, token); got != fiber.StatusForbidden {
			t.Errorf("%s con token de cliente: se esperaba 403 y llegó %d", k, got)
		}
	}
}

func TestTokenAdulteradoEsRechazado(t *testing.T) {
	app := appDePrueba()
	token, _ := auth.GenerarToken("otro-secreto-distinto-de-mas-de-32-caracteres", 1, "admin@upse.edu.ec", "admin")
	for k := range rutasAdmin {
		if got := estado(t, app, k, token); got != fiber.StatusUnauthorized {
			t.Errorf("%s con token firmado con otro secreto: se esperaba 401 y llegó %d", k, got)
		}
	}
}
