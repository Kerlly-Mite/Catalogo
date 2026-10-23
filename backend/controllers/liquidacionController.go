package controllers

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// LiquidacionController agrupa los pagos de comisiones a los distribuidores.
type LiquidacionController struct {
	repo repository.DBRepository
}

func NewLiquidacionController(repo repository.DBRepository) *LiquidacionController {
	return &LiquidacionController{repo: repo}
}

// Liquidar (POST /api/liquidaciones, solo admin): paga todas las comisiones pendientes del distribuidor.
// 201 con la liquidación; 400 si no tiene nada pendiente (así no se paga dos veces); 404 si no existe.
func (l *LiquidacionController) Liquidar(c *fiber.Ctx) error {
	var req models.LiquidarRequest
	if err := c.BodyParser(&req); err != nil || req.ReferidoID < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, "Cuerpo de petición inválido", nil))
	}

	liq, err := l.repo.Liquidar(c.UserContext(), req.ReferidoID)
	var neg *repository.ErrNegocio
	switch {
	case errors.Is(err, repository.ErrNoEncontrado):
		return c.Status(fiber.StatusNotFound).JSON(models.NewAPIError(fiber.StatusNotFound, "Distribuidor no encontrado", nil))
	case errors.As(err, &neg):
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, neg.Mensaje, neg.Detalle))
	case err != nil:
		return errorInterno(c, "liquidar", err)
	}
	return c.Status(fiber.StatusCreated).JSON(liq)
}

// Listar (GET /api/liquidaciones, solo admin): historial de pagos, paginado.
func (l *LiquidacionController) Listar(c *fiber.Ctx) error {
	page, limit := paginacion(c)
	lista, total, err := l.repo.ListarLiquidaciones(c.UserContext(), 0, page, limit)
	if err != nil {
		return errorInterno(c, "liquidaciones", err)
	}
	return c.JSON(models.Pagina[models.Liquidacion]{Data: lista, Page: page, Limit: limit, Total: total})
}

// Mias (GET /api/liquidaciones/mias): los cobros del distribuidor del token.
func (l *LiquidacionController) Mias(c *fiber.Ctx) error {
	page, limit := paginacion(c)
	lista, total, err := l.repo.ListarLiquidaciones(c.UserContext(), middleware.UsuarioID(c), page, limit)
	if err != nil {
		return errorInterno(c, "liquidaciones/mias", err)
	}
	return c.JSON(models.Pagina[models.Liquidacion]{Data: lista, Page: page, Limit: limit, Total: total})
}

// Deuda (GET /api/liquidaciones/deuda, solo admin): cuánto se le debe a cada distribuidor.
func (l *LiquidacionController) Deuda(c *fiber.Ctx) error {
	lista, err := l.repo.DeudaPorReferido(c.UserContext())
	if err != nil {
		return errorInterno(c, "liquidaciones/deuda", err)
	}
	return c.JSON(lista)
}
