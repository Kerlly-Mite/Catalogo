package controllers

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// Máximo por comprobante (el frontend reduce las fotos antes de subirlas; un PDF debe pesar menos de 2 MB).
const maxComprobanteBytes = 2 << 20 // 2 MB

// Formatos de voucher admitidos. Se detectan por el CONTENIDO del archivo, no por su nombre.
var tiposComprobante = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
	"application/pdf": true,
}

// ComprobanteController recibe los vouchers de transferencia y los muestra solo al administrador.
type ComprobanteController struct {
	repo repository.DBRepository
}

func NewComprobanteController(repo repository.DBRepository) *ComprobanteController {
	return &ComprobanteController{repo: repo}
}

// Subir (POST /api/comprobantes, con sesión): recibe un formulario con el campo "comprobante" y devuelve
// {"id": 7}. Ese id viaja luego en pagoDetalle.comprobante_id al confirmar el pedido.
func (cc *ComprobanteController) Subir(c *fiber.Ctx) error {
	archivo, err := c.FormFile("comprobante")
	if err != nil {
		return imagenInvalida(c, `Falta el archivo del comprobante (campo "comprobante")`)
	}
	if archivo.Size > maxComprobanteBytes {
		return imagenInvalida(c, "El comprobante no puede superar 2 MB")
	}
	f, err := archivo.Open()
	if err != nil {
		return errorInterno(c, "abrir comprobante", err)
	}
	defer f.Close()

	datos, err := io.ReadAll(io.LimitReader(f, maxComprobanteBytes+1))
	if err != nil {
		return errorInterno(c, "leer comprobante", err)
	}
	if len(datos) == 0 || len(datos) > maxComprobanteBytes {
		return imagenInvalida(c, "El comprobante está vacío o supera 2 MB")
	}
	tipo := http.DetectContentType(datos)
	if !tiposComprobante[tipo] {
		return imagenInvalida(c, "Formato no permitido: usa una imagen (JPG, PNG, WebP) o un PDF")
	}

	id, err := cc.repo.GuardarComprobante(c.UserContext(), middleware.UsuarioID(c), tipo, datos)
	if err != nil {
		return errorInterno(c, "guardar comprobante", err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

// Ver (GET /api/comprobantes/:id, solo admin): sirve el voucher. Es privado: nunca se guarda en cachés compartidas.
func (cc *ComprobanteController) Ver(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return imagenInvalida(c, "El id del comprobante debe ser un número entero")
	}
	tipo, datos, err := cc.repo.ObtenerComprobante(c.UserContext(), id)
	if errors.Is(err, repository.ErrNoEncontrado) {
		return c.Status(fiber.StatusNotFound).JSON(models.NewAPIError(fiber.StatusNotFound, "Comprobante no encontrado", nil))
	}
	if err != nil {
		return errorInterno(c, "ver comprobante", err)
	}
	c.Set(fiber.HeaderContentType, tipo)
	c.Set(fiber.HeaderCacheControl, "private, no-store")
	c.Set("X-Content-Type-Options", "nosniff")
	return c.Send(datos)
}
