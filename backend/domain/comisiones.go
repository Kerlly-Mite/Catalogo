// Paquete domain: reglas de negocio PURAS (sin base de datos, sin HTTP). Se pueden probar sin infraestructura.
package domain

import "math"

// TasasComision son los porcentajes por nivel ascendente: la posición 0 es el nivel 1 (10 %), luego 5 % y 2 %.
// Deben coincidir con TASA_COMISION de frontend/src/data/red.ts.
var TasasComision = []float64{0.10, 0.05, 0.02}

// Comision es lo que cobra el ascendente de un nivel por una venta.
type Comision struct {
	Nivel      int     // 1 = referente directo, 2 = el de arriba, 3 = tercer nivel
	Porcentaje float64 // ej. 10 para el 10 %
	Monto      float64 // en dólares, redondeado a 2 decimales
}

// RepartirComisiones calcula la comisión de cada nivel sobre el total de una venta.
// Con un total <= 0 no hay nada que repartir.
func RepartirComisiones(total float64, tasas []float64) []Comision {
	if total <= 0 || math.IsNaN(total) {
		return []Comision{}
	}
	reparto := make([]Comision, 0, len(tasas))
	for i, tasa := range tasas {
		reparto = append(reparto, Comision{
			Nivel:      i + 1,
			Porcentaje: math.Round(tasa*10000) / 100,
			Monto:      math.Round(total*tasa*100) / 100,
		})
	}
	return reparto
}
