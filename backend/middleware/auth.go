// Paquete middleware: protege rutas con JWT y expone la identidad del usuario autenticado.
package middleware

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"multicatalogo-backend/auth"
	"multicatalogo-backend/models"
)

// Claves con las que se guarda la identidad en c.Locals (solo se leen con los helpers de abajo).
const (
	claveID    = "auth_usuario_id"
	claveEmail = "auth_email"
	claveRol   = "auth_rol"
)

func noAutorizado(c *fiber.Ctx, mensaje string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(models.NewAPIError(fiber.StatusUnauthorized, mensaje, nil))
}

// RequireAuth exige un header "Authorization: Bearer <token>" válido. Sin token, con token
// adulterado o vencido responde 401. Si es válido, guarda id, email y rol del TOKEN en c.Locals.
func RequireAuth(secreto string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		h := c.Get(fiber.HeaderAuthorization)
		if len(h) <= 7 || !strings.EqualFold(h[:7], "Bearer ") {
			return noAutorizado(c, "Se requiere iniciar sesión")
		}
		claims, err := auth.ValidarToken(secreto, strings.TrimSpace(h[7:]))
		if err != nil {
			return noAutorizado(c, "Sesión inválida o vencida")
		}
		id, err := strconv.Atoi(claims.Subject)
		if err != nil || id <= 0 {
			return noAutorizado(c, "Sesión inválida o vencida")
		}
		c.Locals(claveID, id)
		c.Locals(claveEmail, claims.Email)
		c.Locals(claveRol, claims.Rol)
		return c.Next()
	}
}

// RequireAdmin se usa DESPUÉS de RequireAuth: solo deja pasar al rol "admin" (si no, 403).
func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if Rol(c) != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(models.NewAPIError(fiber.StatusForbidden, "Acceso solo para administradores", nil))
		}
		return c.Next()
	}
}

// UsuarioID devuelve el id del usuario autenticado (0 si la ruta no pasó por RequireAuth).
func UsuarioID(c *fiber.Ctx) int {
	v, _ := c.Locals(claveID).(int)
	return v
}

// Email devuelve el email del usuario autenticado ("" si la ruta no pasó por RequireAuth).
func Email(c *fiber.Ctx) string {
	v, _ := c.Locals(claveEmail).(string)
	return v
}

// Rol devuelve el rol del usuario autenticado ("admin" o "cliente").
func Rol(c *fiber.Ctx) string {
	v, _ := c.Locals(claveRol).(string)
	return v
}
