package controllers

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"

	"multicatalogo-backend/repository"
)

// HealthController comprueba que las dependencias reales responden.
type HealthController struct {
	repo  repository.DBRepository
	cache repository.CacheRepository
}

func NewHealthController(repo repository.DBRepository, cache repository.CacheRepository) *HealthController {
	return &HealthController{repo: repo, cache: cache}
}

// Health (GET /api/health):
//   - 200 "ok"        -> PostgreSQL y Redis responden.
//   - 200 "degraded"  -> PostgreSQL responde pero Redis no (la API sigue funcionando sin caché).
//   - 503 "error"     -> PostgreSQL no responde.
func (h *HealthController) Health(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
	defer cancel()

	estado, codigo, mensaje := "ok", fiber.StatusOK, "Servidor Go/Fiber operativo"
	postgres, redis := "ok", "ok"

	if err := h.repo.Ping(ctx); err != nil {
		log.Printf("health: PostgreSQL no responde: %v", err)
		postgres, estado, codigo, mensaje = "down", "error", fiber.StatusServiceUnavailable, "PostgreSQL no responde"
	}
	if err := h.cache.Ping(ctx); err != nil {
		redis = "down"
		if estado == "ok" {
			estado, mensaje = "degraded", "Redis no responde: la API funciona sin caché"
		}
	}

	return c.Status(codigo).JSON(fiber.Map{
		"status":   estado,
		"message":  mensaje,
		"postgres": postgres,
		"redis":    redis,
	})
}
