package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// Prefijo y duración de la caché del árbol de referidos. La clave completa es red:tree:{email}.
const (
	prefijoRed = "red:tree:"
	ttlRed     = 10 * time.Minute
)

// RedController agrupa el endpoint del árbol multinivel.
type RedController struct {
	repo  repository.DBRepository
	cache repository.CacheRepository
}

func NewRedController(repo repository.DBRepository, cache repository.CacheRepository) *RedController {
	return &RedController{repo: repo, cache: cache}
}

// GetRed devuelve el árbol de referidos con Cache-Aside (GET /api/red, requiere sesión).
// Armar el árbol es una operación costosa, por eso se guarda 10 minutos en Redis bajo red:tree:{email}.
// Si por algún motivo no hay email se usa la clave "red:tree:general".
func (r *RedController) GetRed(c *fiber.Ctx) error {
	ctx := c.UserContext()

	email := repository.NormalizarEmail(middleware.Email(c)) // identidad del token, no del query string
	if email == "" {
		email = "general"
	}
	clave := prefijoRed + email

	cached, hit, err := r.cache.Get(ctx, clave)
	if err != nil {
		log.Printf("caché no disponible (GET %s): %v", clave, err)
	}
	if hit {
		c.Set("X-Cache", "HIT")
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		return c.Send(cached)
	}

	red, err := r.repo.ObtenerRed(ctx)
	if err != nil {
		log.Printf("error al obtener la red: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewAPIError(fiber.StatusInternalServerError, "No se pudo obtener la red de referidos", nil))
	}

	if data, err := json.Marshal(red); err == nil {
		if err := r.cache.Set(ctx, clave, data, ttlRed); err != nil {
			log.Printf("caché no disponible (SET %s): %v", clave, err)
		}
	}
	c.Set("X-Cache", "MISS")
	return c.JSON(red)
}

// Mia devuelve la red del USUARIO del token (GET /api/red/mia): el subárbol desde su nodo y su código de
// invitación. Cache-Aside bajo red:tree:mia:{email} (el checkout y el registro borran todo el prefijo red:tree:).
func (r *RedController) Mia(c *fiber.Ctx) error {
	ctx := c.UserContext()
	clave := prefijoRed + "mia:" + repository.NormalizarEmail(middleware.Email(c))

	cached, hit, err := r.cache.Get(ctx, clave)
	if err != nil {
		log.Printf("caché no disponible (GET %s): %v", clave, err)
	}
	if hit {
		c.Set("X-Cache", "HIT")
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		return c.Send(cached)
	}

	datos, err := r.repo.ObtenerRedUsuario(ctx, middleware.UsuarioID(c))
	if errors.Is(err, repository.ErrNoEncontrado) {
		return c.Status(fiber.StatusNotFound).JSON(models.NewAPIError(fiber.StatusNotFound, "Tu cuenta todavía no tiene un nodo en la red", nil))
	}
	if err != nil {
		return errorInterno(c, "red/mia", err)
	}

	if data, err := json.Marshal(datos); err == nil {
		if err := r.cache.Set(ctx, clave, data, ttlRed); err != nil {
			log.Printf("caché no disponible (SET %s): %v", clave, err)
		}
	}
	c.Set("X-Cache", "MISS")
	return c.JSON(datos)
}

// ComisionesMias devuelve el ledger de comisiones del usuario del token (GET /api/comisiones/mias).
func (r *RedController) ComisionesMias(c *fiber.Ctx) error {
	page, limit := paginacion(c)
	res, err := r.repo.ComisionesMias(c.UserContext(), middleware.UsuarioID(c), page, limit)
	if errors.Is(err, repository.ErrNoEncontrado) {
		return c.Status(fiber.StatusNotFound).JSON(models.NewAPIError(fiber.StatusNotFound, "Usuario no encontrado", nil))
	}
	if err != nil {
		return errorInterno(c, "comisiones/mias", err)
	}
	return c.JSON(res)
}
