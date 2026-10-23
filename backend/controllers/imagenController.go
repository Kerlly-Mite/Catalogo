package controllers

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// Máximo por imagen (el frontend ya la reduce antes de subirla, así que normalmente pesa mucho menos).
const maxImagenBytes = 2 << 20 // 2 MB

// Solo formatos que el navegador muestra como imagen. SVG NO se admite: puede contener scripts.
var tiposImagen = map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true}

// ImagenController sube y sirve las fotos de los productos.
type ImagenController struct {
	repo repository.DBRepository
}

func NewImagenController(repo repository.DBRepository) *ImagenController {
	return &ImagenController{repo: repo}
}

func imagenInvalida(c *fiber.Ctx, mensaje string) error {
	return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, mensaje, nil))
}

// Subir (POST /api/imagenes, solo admin): recibe un formulario con el campo "imagen" y devuelve
// {"id": 12, "url": "/api/imagenes/12"}. El tipo se detecta por el CONTENIDO del archivo, no por su nombre.
func (i *ImagenController) Subir(c *fiber.Ctx) error {
	archivo, err := c.FormFile("imagen")
	if err != nil {
		return imagenInvalida(c, `Falta el archivo de imagen (campo "imagen")`)
	}
	if archivo.Size > maxImagenBytes {
		return imagenInvalida(c, "La imagen no puede superar 2 MB")
	}
	f, err := archivo.Open()
	if err != nil {
		return errorInterno(c, "abrir imagen", err)
	}
	defer f.Close()

	datos, err := io.ReadAll(io.LimitReader(f, maxImagenBytes+1))
	if err != nil {
		return errorInterno(c, "leer imagen", err)
	}
	if len(datos) == 0 || len(datos) > maxImagenBytes {
		return imagenInvalida(c, "La imagen está vacía o supera 2 MB")
	}
	tipo := http.DetectContentType(datos)
	if !tiposImagen[tipo] {
		return imagenInvalida(c, "Formato no permitido: usa JPG, PNG o WebP")
	}

	id, err := i.repo.GuardarImagen(c.UserContext(), tipo, datos)
	if err != nil {
		return errorInterno(c, "guardar imagen", err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id, "url": "/api/imagenes/" + strconv.Itoa(id)})
}

// Ver (GET /api/imagenes/:id, público): sirve la foto. Una foto nunca cambia (cada subida tiene su id),
// así que el navegador puede guardarla un año.
func (i *ImagenController) Ver(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return imagenInvalida(c, "El id de la imagen debe ser un número entero")
	}
	tipo, datos, err := i.repo.ObtenerImagen(c.UserContext(), id)
	if errors.Is(err, repository.ErrNoEncontrado) {
		return c.Status(fiber.StatusNotFound).JSON(models.NewAPIError(fiber.StatusNotFound, "Imagen no encontrada", nil))
	}
	if err != nil {
		return errorInterno(c, "ver imagen", err)
	}
	c.Set(fiber.HeaderContentType, tipo)
	c.Set(fiber.HeaderCacheControl, "public, max-age=31536000, immutable")
	c.Set("X-Content-Type-Options", "nosniff")
	return c.Send(datos)
}
