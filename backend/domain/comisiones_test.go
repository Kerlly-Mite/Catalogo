package domain

import (
	"math"
	"testing"
)

func casi(a, b float64) bool { return math.Abs(a-b) < 0.0001 }

func TestRepartirComisiones(t *testing.T) {
	casos := []struct {
		nombre string
		total  float64
		montos []float64
	}{
		{"pedido de 90", 90, []float64{9.00, 4.50, 1.80}},
		{"redondeo a 2 decimales", 199.98, []float64{20.00, 10.00, 4.00}},
		{"pedido pequeño", 10, []float64{1.00, 0.50, 0.20}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			reparto := RepartirComisiones(c.total, TasasComision)
			if len(reparto) != len(c.montos) {
				t.Fatalf("se esperaban %d comisiones y llegaron %d", len(c.montos), len(reparto))
			}
			for i, esperado := range c.montos {
				if reparto[i].Nivel != i+1 {
					t.Errorf("nivel %d: llegó %d", i+1, reparto[i].Nivel)
				}
				if !casi(reparto[i].Monto, esperado) {
					t.Errorf("nivel %d: se esperaba %.2f y llegó %.2f", i+1, esperado, reparto[i].Monto)
				}
			}
		})
	}
}

func TestRepartirComisionesSinTotal(t *testing.T) {
	for _, total := range []float64{0, -5, math.NaN()} {
		if r := RepartirComisiones(total, TasasComision); len(r) != 0 {
			t.Errorf("total %v: no debía repartir nada y repartió %d", total, len(r))
		}
	}
}

func TestPorcentajes(t *testing.T) {
	r := RepartirComisiones(100, TasasComision)
	for i, esperado := range []float64{10, 5, 2} {
		if !casi(r[i].Porcentaje, esperado) {
			t.Errorf("nivel %d: porcentaje %.2f en vez de %.2f", i+1, r[i].Porcentaje, esperado)
		}
	}
}
