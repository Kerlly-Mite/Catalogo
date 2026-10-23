package controllers

import (
	"errors"
	"log"
	"slices"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// PedidoController agrupa la consulta y gestión de pedidos.
type PedidoController struct {
	repo  repository.DBRepository
	cache repository.CacheRepository
}

func NewPedidoController(repo repository.DBRepository, cache repository.CacheRepository) *PedidoController {
	return &PedidoController{repo: repo, cache: cache}
}

// paginacion lee ?page y ?limit (por defecto 1 y 10; máximo 100 por página).
func paginacion(c *fiber.Ctx) (page, limit int) {
	page, limit = c.QueryInt("page", 1), c.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

func errorInterno(c *fiber.Ctx, contexto string, err error) error {
	log.Printf("%s: %v", contexto, err)
	return c.Status(fiber.StatusInternalServerError).JSON(models.NewAPIError(fiber.StatusInternalServerError, "Error interno del servidor", nil))
}

// Mios (GET /api/pedidos/mios): los pedidos del usuario del TOKEN, paginados.
func (p *PedidoController) Mios(c *fiber.Ctx) error {
	page, limit := paginacion(c)
	lista, total, err := p.repo.ListarPedidos(c.UserContext(), repository.FiltroPedidos{
		UsuarioID: middleware.UsuarioID(c), Page: page, Limit: limit,
	})
	if err != nil {
		return errorInterno(c, "pedidos/mios", err)
	}
	return c.JSON(models.Pagina[models.PedidoListado]{Data: lista, Page: page, Limit: limit, Total: total})
}

// Listar (GET /api/pedidos, solo admin): todos los pedidos, paginados y con filtro ?estado=.
func (p *PedidoController) Listar(c *fiber.Ctx) error {
	page, limit := paginacion(c)
	estado := c.Query("estado")
	if estado != "" && !slices.Contains(repository.EstadosPedido, estado) {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, "El estado no es válido", fiber.Map{"validos": repository.EstadosPedido}))
	}
	lista, total, err := p.repo.ListarPedidos(c.UserContext(), repository.FiltroPedidos{Estado: estado, Page: page, Limit: limit})
	if err != nil {
		return errorInterno(c, "pedidos (admin)", err)
	}
	return c.JSON(models.Pagina[models.PedidoListado]{Data: lista, Page: page, Limit: limit, Total: total})
}

func idPedido(c *fiber.Ctx) (int, error) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return 0, c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, "El id del pedido debe ser un número entero", fiber.Map{"id_recibido": c.Params("id")}))
	}
	return id, nil
}

// Detalle (GET /api/pedidos/:id): cabecera + items. Lo ve su dueño o un admin; para cualquier otro
// usuario responde 404 (no se revela que el pedido existe).
func (p *PedidoController) Detalle(c *fiber.Ctx) error {
	id, err := idPedido(c)
	if err != nil {
		return err
	}
	pedido, err := p.repo.ObtenerPedido(c.UserContext(), id)
	if errors.Is(err, repository.ErrNoEncontrado) {
		return c.Status(fiber.StatusNotFound).JSON(models.NewAPIError(fiber.StatusNotFound, "Pedido no encontrado", nil))
	}
	if err != nil {
		return errorInterno(c, "pedido detalle", err)
	}
	if middleware.Rol(c) != "admin" && pedido.UsuarioID != middleware.UsuarioID(c) {
		return c.Status(fiber.StatusNotFound).JSON(models.NewAPIError(fiber.StatusNotFound, "Pedido no encontrado", nil))
	}
	return c.JSON(pedido)
}

// CambiarEstado (PATCH /api/pedidos/:id/estado, solo admin). Body: {"estado": "preparando"}.
// Transición no permitida -> 400. Al cancelar se devuelve el stock y se revierten comisiones, así que
// también se invalidan las cachés del catálogo y de la red.
func (p *PedidoController) CambiarEstado(c *fiber.Ctx) error {
	id, err := idPedido(c)
	if err != nil {
		return err
	}
	var req models.CambioEstado
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, "Cuerpo de petición inválido", nil))
	}
	if !slices.Contains(repository.EstadosPedido, req.Estado) {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, "El estado no es válido", fiber.Map{"validos": repository.EstadosPedido}))
	}

	ctx := c.UserContext()
	err = p.repo.CambiarEstadoPedido(ctx, id, req.Estado)
	var neg *repository.ErrNegocio
	switch {
	case errors.Is(err, repository.ErrNoEncontrado):
		return c.Status(fiber.StatusNotFound).JSON(models.NewAPIError(fiber.StatusNotFound, "Pedido no encontrado", nil))
	case errors.As(err, &neg):
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, neg.Mensaje, neg.Detalle))
	case err != nil:
		return errorInterno(c, "cambiar estado de pedido", err)
	}

	if req.Estado == "cancelado" {
		if err := p.cache.Delete(ctx, claveProductos); err != nil {
			log.Printf("caché no disponible (DEL %s): %v", claveProductos, err)
		}
		if err := p.cache.DeleteByPrefix(ctx, prefijoRed); err != nil {
			log.Printf("caché no disponible (limpieza %s*): %v", prefijoRed, err)
		}
	}

	pedido, err := p.repo.ObtenerPedido(ctx, id)
	if err != nil {
		return errorInterno(c, "pedido tras cambiar estado", err)
	}
	return c.JSON(pedido)
}
