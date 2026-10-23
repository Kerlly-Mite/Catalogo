package controllers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// UsuarioController agrupa la administración de cuentas (solo admin).
type UsuarioController struct {
	repo repository.DBRepository
}

func NewUsuarioController(repo repository.DBRepository) *UsuarioController {
	return &UsuarioController{repo: repo}
}

// Listar (GET /api/usuarios?q=&page=&limit=): usuarios paginados con el formato estándar {data,page,limit,total}.
func (u *UsuarioController) Listar(c *fiber.Ctx) error {
	page, limit := paginacion(c)
	lista, total, err := u.repo.ListarUsuarios(c.UserContext(), c.Query("q"), page, limit)
	if err != nil {
		return errorInterno(c, "listar usuarios", err)
	}
	return c.JSON(models.Pagina[models.UsuarioListado]{Data: lista, Page: page, Limit: limit, Total: total})
}

// Actualizar (PATCH /api/usuarios/:id, body {"rol"?, "activo"?}): cambia el rol o activa/desactiva la cuenta.
// Un admin no puede desactivarse ni quitarse el rol a sí mismo: así siempre queda al menos un admin activo.
// Nota: un token ya emitido sigue valiendo hasta que venza (2 h); la cuenta desactivada no podrá volver a iniciar sesión.
func (u *UsuarioController) Actualizar(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, "El id del usuario debe ser un número entero", fiber.Map{"id_recibido": c.Params("id")}))
	}
	var cambios models.UsuarioCambios
	if err := c.BodyParser(&cambios); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, "Cuerpo de petición inválido", nil))
	}
	if cambios.Rol == nil && cambios.Activo == nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, "No hay cambios: envía \"rol\" y/o \"activo\"", nil))
	}
	if cambios.Rol != nil && *cambios.Rol != "admin" && *cambios.Rol != "cliente" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, "El rol debe ser \"admin\" o \"cliente\"", nil))
	}
	if id == middleware.UsuarioID(c) &&
		((cambios.Activo != nil && !*cambios.Activo) || (cambios.Rol != nil && *cambios.Rol != "admin")) {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, "No puedes desactivarte ni quitarte el rol de administrador a ti mismo", nil))
	}

	usuario, err := u.repo.ActualizarUsuario(c.UserContext(), id, cambios)
	if errors.Is(err, repository.ErrNoEncontrado) {
		return c.Status(fiber.StatusNotFound).JSON(models.NewAPIError(fiber.StatusNotFound, "Usuario no encontrado", nil))
	}
	if err != nil {
		return errorInterno(c, "actualizar usuario", err)
	}
	return c.JSON(usuario)
}
