// Seguimos dentro del paquete controllers, ya que maneja la lógica de otra sección del negocio.
package controllers

import (
	"strconv"

	// Importamos Fiber para manejar la respuesta HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de modelos para usar la estructura Producto.
	"multicatalogo-backend/models"
)

// GetProductos es la función controladora encargada de devolver el catálogo de artículos.
func GetProductos(c *fiber.Ctx) error {
	// Declaramos e inicializamos un 'slice' (arreglo dinámico) usando nuestro modelo models.Producto.
	productos := listaProductos()

	// Fiber convierte automáticamente el slice de estructuras a formato JSON y lo envía como respuesta al cliente.
	return c.JSON(productos)
}
// listaProductos centraliza los datos estáticos (luego vendrán del repositorio/BD en la Fase 2).
func listaProductos() []models.Producto {
	return []models.Producto{
		{ID: 1, Nombre: "Serum Revitalizante", Precio: 45.00, Img: "https://picsum.photos/seed/serum/150"},
		{ID: 2, Nombre: "Crema Hidratante Pro", Precio: 32.50, Img: "https://picsum.photos/seed/crema/150"},
		{ID: 3, Nombre: "Tónico Purificante", Precio: 28.00, Img: "https://picsum.photos/seed/tonico/150"},
		{ID: 4, Nombre: "Mascarilla Nocturna", Precio: 50.00, Img: "https://picsum.photos/seed/mascarilla/150"},
	}
}

// GetProductoPorID devuelve un producto (GET /api/productos/:id).
// Paso 3 (Fase 1): si el id no es un número, responde 400 con APIError.
func GetProductoPorID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewAPIError(
			fiber.StatusBadRequest,
			"El id del producto debe ser un número entero",
			fiber.Map{"id_recibido": c.Params("id")},
		))
	}
	for _, p := range listaProductos() {
		if p.ID == id {
			return c.JSON(p)
		}
	}
	return c.Status(fiber.StatusNotFound).JSON(models.NewAPIError(fiber.StatusNotFound, "Producto no encontrado", nil))
}
