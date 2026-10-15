// Declaramos el paquete models para agrupar las estructuras de datos de la aplicación.
package models

// LoginRequest define la estructura esperada para el cuerpo de la petición (JSON) al iniciar sesión.
type LoginRequest struct {
	// Email representa el correo del usuario; la etiqueta `json:"email"` indica cómo se mapea el JSON a esta variable.
	Email    string `json:"email"`
	// Password representa la contraseña del usuario; se extrae del campo "password" del JSON entrante.
	Password string `json:"password"`
}

// Producto define la estructura de los datos de un artículo en nuestro catálogo.
type Producto struct {
	// ID es el identificador único numérico del producto.
	ID     int     `json:"id"`
	// Nombre es la descripción en texto del producto.
	Nombre string  `json:"nombre"`
	// Precio es el costo del producto, almacenado como un número con decimales (float64).
	Precio float64 `json:"precio"`
	// Img es la URL o ruta que apunta a la fotografía o imagen del producto.
	Img    string  `json:"img"`
}

// APIError estandariza TODOS los mensajes de error HTTP de la API.
type APIError struct {
	// Status es el código HTTP del error (400, 401, 404...).
	Status int `json:"status"`
	// Message es el mensaje legible que describe el error.
	Message string `json:"message"`
	// Details es información adicional opcional (ej. campos inválidos). Se omite si es nil.
	Details interface{} `json:"details,omitempty"`
	// Error repite Message: el frontend actual lee `data.error` en el login,
	// así mantenemos compatibilidad sin tocar React (regla 3 del DRS).
	Error string `json:"error"`
}

// NewAPIError construye un APIError rellenando también el campo de compatibilidad.
func NewAPIError(status int, message string, details interface{}) APIError {
	return APIError{Status: status, Message: message, Details: details, Error: message}
}
