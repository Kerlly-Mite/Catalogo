package domain

import (
	"errors"
	"math"
	"time"
)

// ZonaEcuador es UTC-5 fijo (Ecuador continental no usa horario de verano). Así los reportes por fecha
// no dependen de la zona horaria del servidor ni de que exista la base de datos de zonas (Windows).
var ZonaEcuador = time.FixedZone("ECT", -5*60*60)

// ErrRangoInvalido: una fecha no tiene el formato AAAA-MM-DD o "desde" es posterior a "hasta".
var ErrRangoInvalido = errors.New("rango de fechas inválido")

const formatoFecha = "2006-01-02"

// RangoFechas convierte "desde" y "hasta" (AAAA-MM-DD, AMBAS inclusivas, ambas opcionales) en un
// intervalo semiabierto [inicio, fin) en hora de Ecuador, listo para `fecha >= inicio AND fecha < fin`.
// Un extremo vacío devuelve nil (sin límite por ese lado).
func RangoFechas(desde, hasta string) (inicio, fin *time.Time, err error) {
	if desde != "" {
		t, e := time.ParseInLocation(formatoFecha, desde, ZonaEcuador)
		if e != nil {
			return nil, nil, ErrRangoInvalido
		}
		inicio = &t
	}
	if hasta != "" {
		t, e := time.ParseInLocation(formatoFecha, hasta, ZonaEcuador)
		if e != nil {
			return nil, nil, ErrRangoInvalido
		}
		t = t.AddDate(0, 0, 1) // "hasta" incluye todo ese día
		fin = &t
	}
	if inicio != nil && fin != nil && !inicio.Before(*fin) {
		return nil, nil, ErrRangoInvalido
	}
	return inicio, fin, nil
}

// TicketPromedio es el total vendido dividido entre el número de pedidos (0 si no hubo pedidos).
func TicketPromedio(total float64, pedidos int) float64 {
	if pedidos <= 0 || math.IsNaN(total) {
		return 0
	}
	return math.Round(total/float64(pedidos)*100) / 100
}
