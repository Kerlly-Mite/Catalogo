package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// Claves y duración de la caché del catálogo.
const (
	claveProductos = "productos:all"
	ttlProductos   = 5 * time.Minute
)

// ProductoController agrupa los endpoints del catálogo.
type ProductoController struct {
	repo  repository.DBRepository
	cache repository.CacheRepository
}

func NewProductoController(repo repository.DBRepository, cache repository.CacheRepository) *ProductoController {
	return &ProductoController{repo: repo, cache: cache}
}

// GetProductos devuelve el catálogo con el patrón Cache-Aside (GET /api/productos):
//  1. busca "productos:all" en Redis  -> si existe (HIT) responde sin tocar PostgreSQL
//  2. si no existe (MISS) consulta PostgreSQL, guarda el JSON en Redis por 5 minutos y responde
//
// La cabecera X-Cache indica HIT o MISS (se ve en la pestaña Network del navegador).
func (p *ProductoController) GetProductos(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// 1) Intentamos leer de la caché. Si Redis falla, seguimos con la BD (la caché nunca debe tumbar la API).
	cached, hit, err := p.cache.Get(ctx, claveProductos)
	if err != nil {
		log.Printf("caché no disponible (GET %s): %v", claveProductos, err)
	}
	if hit {
		c.Set("X-Cache", "HIT")
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		return c.Send(cached)
	}

	// 2) Cache Miss: consultamos PostgreSQL.
	productos, err := p.repo.ObtenerProductos(ctx)
	if err != nil {
		log.Printf("error al obtener productos: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewAPIError(fiber.StatusInternalServerError, "No se pudieron obtener los productos", nil))
	}

	// Serializamos, guardamos en Redis con TTL y respondemos.
	if data, err := json.Marshal(productos); err == nil {
		if err := p.cache.Set(ctx, claveProductos, data, ttlProductos); err != nil {
			log.Printf("caché no disponible (SET %s): %v", claveProductos, err)
		}
	}
	c.Set("X-Cache", "MISS")
	return c.JSON(productos)
}

// GetProductoPorID devuelve un producto (GET /api/productos/:id).
// Si el id no es un número, responde 400 con APIError; si no existe, 404.
func (p *ProductoController) GetProductoPorID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(
			fiber.StatusBadRequest,
			"El id del producto debe ser un número entero",
			fiber.Map{"id_recibido": c.Params("id")},
		))
	}

	producto, err := p.repo.ObtenerProductoPorID(c.UserContext(), id)
	if errors.Is(err, repository.ErrNoEncontrado) {
		return c.Status(fiber.StatusNotFound).JSON(models.NewAPIError(fiber.StatusNotFound, "Producto no encontrado", nil))
	}
	if err != nil {
		log.Printf("error al obtener producto %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewAPIError(fiber.StatusInternalServerError, "Error interno del servidor", nil))
	}
	return c.JSON(producto)
}

// ---------- ADMINISTRACIÓN DE PRODUCTOS (solo admin: RequireAuth + RequireAdmin en routes.go) ----------

// invalidarCatalogo borra la caché del catálogo: se llama después de CADA escritura (mismo patrón que el checkout).
func (p *ProductoController) invalidarCatalogo(c *fiber.Ctx) {
	if err := p.cache.Delete(c.UserContext(), claveProductos); err != nil {
		log.Printf("caché no disponible (DEL %s): %v", claveProductos, err)
	}
}

func respuestaValidacion(c *fiber.Ctx, problemas []string) error {
	return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(
		fiber.StatusBadRequest, "Datos del producto inválidos: "+strings.Join(problemas, "; "), fiber.Map{"errores": problemas}))
}

// Validaciones compartidas por crear y editar (cada una recibe el valor ya recortado).
func problemasNombre(v string) []string {
	if n := len([]rune(v)); n < 2 || n > 150 {
		return []string{"nombre debe tener entre 2 y 150 caracteres"}
	}
	return nil
}
func problemasPrecio(v float64) []string {
	if math.IsNaN(v) || v <= 0 || v > 99999999 {
		return []string{"precio debe ser mayor que cero"}
	}
	return nil
}
func problemasCategoria(v string) []string {
	if n := len([]rune(v)); n < 1 || n > 50 {
		return []string{"categoria es obligatoria (máximo 50 caracteres)"}
	}
	return nil
}
func problemasStock(v int) []string {
	if v < 0 || v > 1000000 {
		return []string{"stock debe estar entre 0 y 1.000.000"}
	}
	return nil
}
func problemasImagenes(img string, galeria []string) []string {
	var out []string
	if len(img) > 500 {
		out = append(out, "img no puede superar 500 caracteres")
	}
	if len(galeria) > 10 {
		out = append(out, "galeria admite como máximo 10 imágenes")
	}
	for _, g := range galeria {
		if len(g) > 500 {
			out = append(out, "cada imagen de galeria puede tener como máximo 500 caracteres")
			break
		}
	}
	return out
}

func limpiarGaleria(g []string) []string {
	if g == nil {
		return nil
	}
	limpia := []string{}
	for _, u := range g {
		if u = strings.TrimSpace(u); u != "" {
			limpia = append(limpia, u)
		}
	}
	return limpia
}

func redondear2(v float64) float64 { return math.Round(v*100) / 100 }

// CrearProducto (POST /api/productos) -> 201 con el producto creado.
func (p *ProductoController) CrearProducto(c *fiber.Ctx) error {
	var in models.ProductoEntrada
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, "Cuerpo de petición inválido", nil))
	}
	in.Nombre = strings.TrimSpace(in.Nombre)
	in.Descripcion = strings.TrimSpace(in.Descripcion)
	in.Categoria = strings.TrimSpace(in.Categoria)
	in.Img = strings.TrimSpace(in.Img)
	in.Galeria = limpiarGaleria(in.Galeria)
	in.Precio = redondear2(in.Precio)

	var problemas []string
	problemas = append(problemas, problemasNombre(in.Nombre)...)
	problemas = append(problemas, problemasPrecio(in.Precio)...)
	problemas = append(problemas, problemasCategoria(in.Categoria)...)
	problemas = append(problemas, problemasStock(in.Stock)...)
	problemas = append(problemas, problemasImagenes(in.Img, in.Galeria)...)
	if len(problemas) > 0 {
		return respuestaValidacion(c, problemas)
	}

	producto, err := p.repo.CrearProducto(c.UserContext(), in)
	if err != nil {
		return errorInterno(c, "crear producto", err)
	}
	p.invalidarCatalogo(c)
	return c.Status(fiber.StatusCreated).JSON(producto)
}

// ActualizarProducto (PUT /api/productos/:id): edición PARCIAL. También sirve para reactivar ({"activo": true}).
func (p *ProductoController) ActualizarProducto(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, "El id del producto debe ser un número entero", fiber.Map{"id_recibido": c.Params("id")}))
	}
	var cambios models.ProductoCambios
	if err := c.BodyParser(&cambios); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, "Cuerpo de petición inválido", nil))
	}

	var problemas []string
	if cambios.Nombre != nil {
		v := strings.TrimSpace(*cambios.Nombre)
		cambios.Nombre = &v
		problemas = append(problemas, problemasNombre(v)...)
	}
	if cambios.Descripcion != nil {
		v := strings.TrimSpace(*cambios.Descripcion)
		cambios.Descripcion = &v
	}
	if cambios.Precio != nil {
		v := redondear2(*cambios.Precio)
		cambios.Precio = &v
		problemas = append(problemas, problemasPrecio(v)...)
	}
	if cambios.Categoria != nil {
		v := strings.TrimSpace(*cambios.Categoria)
		cambios.Categoria = &v
		problemas = append(problemas, problemasCategoria(v)...)
	}
	if cambios.Stock != nil {
		problemas = append(problemas, problemasStock(*cambios.Stock)...)
	}
	img := ""
	if cambios.Img != nil {
		v := strings.TrimSpace(*cambios.Img)
		cambios.Img = &v
		img = v
	}
	cambios.Galeria = limpiarGaleria(cambios.Galeria)
	problemas = append(problemas, problemasImagenes(img, cambios.Galeria)...)
	if len(problemas) > 0 {
		return respuestaValidacion(c, problemas)
	}

	producto, err := p.repo.ActualizarProducto(c.UserContext(), id, cambios)
	if errors.Is(err, repository.ErrNoEncontrado) {
		return c.Status(fiber.StatusNotFound).JSON(models.NewAPIError(fiber.StatusNotFound, "Producto no encontrado", nil))
	}
	if err != nil {
		return errorInterno(c, "actualizar producto", err)
	}
	p.invalidarCatalogo(c)
	return c.JSON(producto)
}

// EliminarProducto (DELETE /api/productos/:id): borrado LÓGICO (activo=false). La fila queda para los pedidos antiguos.
func (p *ProductoController) EliminarProducto(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(fiber.StatusBadRequest, "El id del producto debe ser un número entero", fiber.Map{"id_recibido": c.Params("id")}))
	}
	err = p.repo.DesactivarProducto(c.UserContext(), id)
	if errors.Is(err, repository.ErrNoEncontrado) {
		return c.Status(fiber.StatusNotFound).JSON(models.NewAPIError(fiber.StatusNotFound, "Producto no encontrado", nil))
	}
	if err != nil {
		return errorInterno(c, "desactivar producto", err)
	}
	p.invalidarCatalogo(c)
	return c.SendStatus(fiber.StatusNoContent)
}

// ListarAdmin (GET /api/admin/productos?q=&page=&limit=): activos E inactivos, con búsqueda y paginación.
func (p *ProductoController) ListarAdmin(c *fiber.Ctx) error {
	page, limit := paginacion(c)
	lista, total, err := p.repo.ListarProductosAdmin(c.UserContext(), c.Query("q"), page, limit)
	if err != nil {
		return errorInterno(c, "listar productos (admin)", err)
	}
	return c.JSON(models.Pagina[models.Producto]{Data: lista, Page: page, Limit: limit, Total: total})
}
