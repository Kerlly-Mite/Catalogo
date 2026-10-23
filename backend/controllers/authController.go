package controllers

import (
	"errors"
	"log"
	"net/mail"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"multicatalogo-backend/auth"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// AuthController agrupa el login. Recibe sus dependencias por el constructor (inyección).
type AuthController struct {
	repo  repository.DBRepository
	cache repository.CacheRepository
	// jwtSecret firma los tokens (viene de JWT_SECRET en el .env, lo inyecta main.go).
	jwtSecret string
}

func NewAuthController(repo repository.DBRepository, cache repository.CacheRepository, jwtSecret string) *AuthController {
	return &AuthController{repo: repo, cache: cache, jwtSecret: jwtSecret}
}

// Login valida las credenciales contra la tabla usuarios (POST /api/login).
func (a *AuthController) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, "Cuerpo de petición inválido", nil))
	}

	// Validamos que email y password no vengan vacíos antes de consultar la BD.
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		faltantes := []string{}
		if req.Email == "" {
			faltantes = append(faltantes, "email")
		}
		if req.Password == "" {
			faltantes = append(faltantes, "password")
		}
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(
			fiber.StatusBadRequest,
			"Email y contraseña son obligatorios",
			fiber.Map{"campos_vacios": faltantes},
		))
	}

	credencialesIncorrectas := func() error {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewAPIError(fiber.StatusUnauthorized, "Credenciales incorrectas", nil))
	}

	// Buscamos por email y comparamos la contraseña contra el hash bcrypt guardado.
	usuario, err := a.repo.BuscarUsuarioPorEmail(c.UserContext(), req.Email)
	if errors.Is(err, repository.ErrNoEncontrado) {
		return credencialesIncorrectas()
	}
	if err != nil {
		log.Printf("error en login: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewAPIError(fiber.StatusInternalServerError, "Error interno del servidor", nil))
	}
	if bcrypt.CompareHashAndPassword([]byte(usuario.Password), []byte(req.Password)) != nil {
		return credencialesIncorrectas()
	}

	// Las cuentas desactivadas no pueden iniciar sesión (se comprueba después de la contraseña).
	if !usuario.Activo {
		return c.Status(fiber.StatusForbidden).JSON(models.NewAPIError(fiber.StatusForbidden, "Cuenta desactivada", nil))
	}

	// JWT real (HS256, 2 horas) con id, email y rol del usuario.
	token, err := auth.GenerarToken(a.jwtSecret, usuario.ID, usuario.Email, usuario.Rol)
	if err != nil {
		log.Printf("error al firmar el token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewAPIError(fiber.StatusInternalServerError, "Error interno del servidor", nil))
	}
	return c.JSON(fiber.Map{"token": token, "email": usuario.Email, "rol": usuario.Rol})
}

// Register crea una cuenta nueva (POST /api/register). El rol siempre es "cliente".
// El código de referido (opcional) puede venir en el body ("ref") o en la URL (?ref=CODIGO).
func (a *AuthController) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, "Cuerpo de petición inválido", nil))
	}
	req.Nombre = strings.TrimSpace(req.Nombre)
	req.Email = repository.NormalizarEmail(req.Email)
	ref := strings.ToUpper(strings.TrimSpace(req.Ref))
	if ref == "" {
		ref = strings.ToUpper(strings.TrimSpace(c.Query("ref")))
	}

	// ---- Validaciones (400) ----
	problemas := []string{}
	if n := len([]rune(req.Nombre)); n < 2 || n > 100 {
		problemas = append(problemas, "nombre debe tener entre 2 y 100 caracteres")
	}
	if addr, err := mail.ParseAddress(req.Email); err != nil || addr.Address != req.Email || len(req.Email) > 150 {
		problemas = append(problemas, "email no es válido")
	}
	if len(req.Password) < 8 {
		problemas = append(problemas, "password debe tener al menos 8 caracteres")
	}
	if len(req.Password) > 72 { // límite de bcrypt
		problemas = append(problemas, "password no puede superar 72 caracteres")
	}
	if len(problemas) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(
			fiber.StatusBadRequest, "Datos de registro inválidos: "+strings.Join(problemas, "; "), fiber.Map{"errores": problemas}))
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("register: error al hashear: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewAPIError(fiber.StatusInternalServerError, "Error interno del servidor", nil))
	}

	usuario, codigo, err := a.repo.RegistrarUsuario(c.UserContext(), req.Nombre, req.Email, string(hash), ref)
	switch {
	case errors.Is(err, repository.ErrEmailDuplicado):
		return c.Status(fiber.StatusConflict).JSON(models.NewAPIError(fiber.StatusConflict, "Ya existe una cuenta con ese email", nil))
	case errors.Is(err, repository.ErrReferidoInvalido):
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, "El código de referido no existe", fiber.Map{"ref": ref}))
	case err != nil:
		log.Printf("register: error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewAPIError(fiber.StatusInternalServerError, "Error interno del servidor", nil))
	}

	// La red cambió (hay un nodo nuevo): la caché del árbol ya no es válida.
	if err := a.cache.DeleteByPrefix(c.UserContext(), prefijoRed); err != nil {
		log.Printf("register: caché no disponible (limpieza %s*): %v", prefijoRed, err)
	}

	// Iniciamos sesión automáticamente: devolvemos el mismo formato que /api/login.
	token, err := auth.GenerarToken(a.jwtSecret, usuario.ID, usuario.Email, usuario.Rol)
	if err != nil {
		log.Printf("register: error al firmar el token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewAPIError(fiber.StatusInternalServerError, "Error interno del servidor", nil))
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"token":           token,
		"email":           usuario.Email,
		"rol":             usuario.Rol,
		"nombre":          usuario.Nombre,
		"codigo_referido": codigo,
	})
}
