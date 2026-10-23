package controllers

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// Métodos de pago que ofrece Checkout.tsx.
var metodosPago = map[string]bool{
	"Tarjeta de crédito":      true,
	"Transferencia bancaria":  true,
	"Efectivo contra entrega": true,
}

// CheckoutController procesa las compras. Recibe repositorio y caché por interfaz (inyección).
type CheckoutController struct {
	repo  repository.DBRepository
	cache repository.CacheRepository
	// wg cuenta las Goroutines de comisiones en curso para esperarlas al apagar el servidor.
	wg sync.WaitGroup
}

func NewCheckoutController(repo repository.DBRepository, cache repository.CacheRepository) *CheckoutController {
	return &CheckoutController{repo: repo, cache: cache}
}

// Esperar bloquea hasta que terminen las comisiones pendientes (se llama al apagar el servidor).
func (cc *CheckoutController) Esperar() { cc.wg.Wait() }

// Checkout procesa la compra (POST /api/checkout).
func (cc *CheckoutController) Checkout(c *fiber.Ctx) error {
	var req models.CheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, "Cuerpo de petición inválido", nil))
	}

	// ---- Validaciones (HTTP 400) ----
	req.Cliente.Nombre = strings.TrimSpace(req.Cliente.Nombre)
	// Identidad: SIEMPRE sale del TOKEN (lo puso RequireAuth), nunca del body. Si el cliente manda otro
	// email en el JSON, se ignora: ni se guarda en el pedido ni se le atribuyen comisiones.
	emailToken := repository.NormalizarEmail(middleware.Email(c))
	req.Cliente.Email = emailToken
	req.Cliente.Direccion = strings.TrimSpace(req.Cliente.Direccion)
	req.Cliente.Ciudad = strings.TrimSpace(req.Cliente.Ciudad)
	req.Cliente.Telefono = strings.TrimSpace(req.Cliente.Telefono)

	problemas := []string{}
	if req.Cliente.Nombre == "" {
		problemas = append(problemas, "cliente.nombre es obligatorio")
	}
	if !strings.Contains(req.Cliente.Email, "@") {
		problemas = append(problemas, "cliente.email no es válido")
	}
	if req.Cliente.Direccion == "" {
		problemas = append(problemas, "cliente.direccion es obligatoria")
	}
	if req.Cliente.Ciudad == "" {
		problemas = append(problemas, "cliente.ciudad es obligatoria")
	}
	if !metodosPago[req.MetodoPago] {
		problemas = append(problemas, "metodoPago no es válido")
	}
	if len(req.Items) == 0 {
		problemas = append(problemas, "items no puede estar vacío")
	}
	for i, it := range req.Items {
		if it.ID <= 0 {
			problemas = append(problemas, "items["+strconv.Itoa(i)+"].id inválido")
		}
		if it.Cantidad <= 0 || it.Cantidad > 100 {
			problemas = append(problemas, "items["+strconv.Itoa(i)+"].cantidad debe estar entre 1 y 100")
		}
	}
	if len(problemas) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, "Datos de compra inválidos", fiber.Map{"errores": problemas}))
	}

	ctx := c.UserContext()

	// ¿Quién compra? El usuario del token.
	usuario, err := cc.repo.BuscarUsuarioPorEmail(ctx, emailToken)
	if errors.Is(err, repository.ErrNoEncontrado) {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewAPIError(fiber.StatusUnauthorized, "Sesión inválida: el usuario ya no existe", nil))
	}
	if err != nil {
		log.Printf("checkout: error buscando usuario: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewAPIError(fiber.StatusInternalServerError, "Error interno del servidor", nil))
	}

	// ---- Transacción ACID: todo o nada ----
	pedido, err := cc.repo.GuardarPedido(ctx, req, usuario)
	if err != nil {
		var neg *repository.ErrNegocio
		if errors.As(err, &neg) {
			return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, neg.Mensaje, neg.Detalle))
		}
		log.Printf("checkout: ROLLBACK, no se guardó el pedido: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewAPIError(
			fiber.StatusInternalServerError,
			"No se pudo procesar el pedido. No se realizó ningún cargo; intenta nuevamente.",
			nil,
		))
	}

	// ---- Concurrencia: las comisiones se calculan EN SEGUNDO PLANO ----
	// Respondemos al cliente de inmediato; la Goroutine sigue trabajando después.
	cc.wg.Add(1)
	go cc.calcularComisiones(pedido.ID, pedido.Total, usuario.ID, usuario.Email)

	return c.Status(fiber.StatusCreated).JSON(pedido)
}

// calcularComisiones corre en una Goroutine DESPUÉS del COMMIT. Actualiza los saldos de los 3 niveles
// ascendentes y luego invalida la caché de la red para que el Dashboard muestre los KPIs nuevos.
func (cc *CheckoutController) calcularComisiones(pedidoID int, total float64, usuarioID int, email string) {
	defer cc.wg.Done()
	// Una falla aquí nunca debe tumbar el servidor.
	defer func() {
		if r := recover(); r != nil {
			log.Printf("comisiones: pánico recuperado (pedido %d): %v", pedidoID, r)
		}
	}()

	// Contexto propio: el de la petición HTTP ya terminó cuando respondimos al cliente.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := cc.repo.AplicarComisiones(ctx, pedidoID, usuarioID, total); err != nil {
		// El pedido queda con comisiones_calculadas = FALSE, así se puede reintentar luego.
		log.Printf("comisiones: error en pedido %d: %v", pedidoID, err)
		return
	}

	// Paso 5: invalidar la caché. Se borra red:tree:{email del comprador} y también la de sus
	// ascendentes (la red es compartida, así que se limpian todas las claves red:tree:*).
	if err := cc.cache.Delete(ctx, prefijoRed+email); err != nil {
		log.Printf("caché no disponible (DEL red:tree:%s): %v", email, err)
	}
	if err := cc.cache.DeleteByPrefix(ctx, prefijoRed); err != nil {
		log.Printf("caché no disponible (limpieza %s*): %v", prefijoRed, err)
	}
	log.Printf("comisiones: pedido %d procesado y caché de la red invalidada", pedidoID)
}
