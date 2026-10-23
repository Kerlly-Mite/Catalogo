package controllers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"

	"multicatalogo-backend/domain"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// ReporteController agrupa los reportes del negocio (solo admin). Se calculan con SQL de agregación en cada
// consulta (sin caché) para que el Dashboard siempre refleje los pedidos y comisiones reales.
type ReporteController struct {
	repo repository.DBRepository
}

func NewReporteController(repo repository.DBRepository) *ReporteController {
	return &ReporteController{repo: repo}
}

// rangoDeReporte lee ?desde=AAAA-MM-DD&hasta=AAAA-MM-DD (ambos opcionales e inclusivos). Si no son válidos
// ya respondió 400 y devuelve ok=false.
func rangoDeReporte(c *fiber.Ctx) (desde, hasta *time.Time, ok bool) {
	desde, hasta, err := domain.RangoFechas(c.Query("desde"), c.Query("hasta"))
	if errors.Is(err, domain.ErrRangoInvalido) {
		_ = c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest,
			"Rango de fechas inválido: usa el formato AAAA-MM-DD y que \"desde\" no sea posterior a \"hasta\"",
			fiber.Map{"desde": c.Query("desde"), "hasta": c.Query("hasta")}))
		return nil, nil, false
	}
	return desde, hasta, true
}

// limiteTop lee ?limit= para los rankings (por defecto 5, máximo 50).
func limiteTop(c *fiber.Ctx) int {
	limite := c.QueryInt("limit", 5)
	if limite < 1 {
		limite = 5
	}
	if limite > 50 {
		limite = 50
	}
	return limite
}

// Resumen (GET /api/reportes/resumen?desde&hasta): total vendido, nº de pedidos, ticket promedio y comisiones.
func (r *ReporteController) Resumen(c *fiber.Ctx) error {
	desde, hasta, ok := rangoDeReporte(c)
	if !ok {
		return nil
	}
	res, err := r.repo.ResumenVentas(c.UserContext(), desde, hasta)
	if err != nil {
		return errorInterno(c, "reportes/resumen", err)
	}
	res.Desde, res.Hasta = c.Query("desde"), c.Query("hasta")
	return c.JSON(res)
}

// TopProductos (GET /api/reportes/top-productos?desde&hasta&limit): los más vendidos por ingresos.
func (r *ReporteController) TopProductos(c *fiber.Ctx) error {
	desde, hasta, ok := rangoDeReporte(c)
	if !ok {
		return nil
	}
	lista, err := r.repo.TopProductos(c.UserContext(), desde, hasta, limiteTop(c))
	if err != nil {
		return errorInterno(c, "reportes/top-productos", err)
	}
	return c.JSON(lista)
}

// TopReferidos (GET /api/reportes/top-referidos?desde&hasta&limit): distribuidores con más comisión devengada.
func (r *ReporteController) TopReferidos(c *fiber.Ctx) error {
	desde, hasta, ok := rangoDeReporte(c)
	if !ok {
		return nil
	}
	lista, err := r.repo.TopReferidos(c.UserContext(), desde, hasta, limiteTop(c))
	if err != nil {
		return errorInterno(c, "reportes/top-referidos", err)
	}
	return c.JSON(lista)
}

// ComisionesPendientes (GET /api/reportes/comisiones-pendientes): la deuda de comisiones por distribuidor
// (la misma consulta que usa la pantalla de liquidaciones) más el total adeudado.
func (r *ReporteController) ComisionesPendientes(c *fiber.Ctx) error {
	lista, err := r.repo.DeudaPorReferido(c.UserContext())
	if err != nil {
		return errorInterno(c, "reportes/comisiones-pendientes", err)
	}
	var total float64
	for _, d := range lista {
		total += d.Pendiente
	}
	return c.JSON(models.ComisionesPendientesReporte{Data: lista, TotalPendiente: total})
}
