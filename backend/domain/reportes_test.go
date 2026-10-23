package domain

import (
	"errors"
	"testing"
)

func TestRangoFechas(t *testing.T) {
	ini, fin, err := RangoFechas("2026-10-01", "2026-10-31")
	if err != nil {
		t.Fatalf("no debía fallar: %v", err)
	}
	if got := ini.Format("2006-01-02 15:04 -0700"); got != "2026-10-01 00:00 -0500" {
		t.Errorf("inicio: %s", got)
	}
	// "hasta" es inclusivo: el límite superior (excluyente) es el día siguiente a las 00:00.
	if got := fin.Format("2006-01-02 15:04 -0700"); got != "2026-11-01 00:00 -0500" {
		t.Errorf("fin: %s", got)
	}
}

func TestRangoFechasAbierto(t *testing.T) {
	ini, fin, err := RangoFechas("", "")
	if err != nil || ini != nil || fin != nil {
		t.Fatalf("sin fechas no hay límites: %v %v %v", ini, fin, err)
	}
	ini, fin, err = RangoFechas("2026-10-05", "")
	if err != nil || ini == nil || fin != nil {
		t.Fatalf("solo desde: %v %v %v", ini, fin, err)
	}
}

func TestRangoFechasInvalido(t *testing.T) {
	casos := [][2]string{
		{"2026/10/01", ""},           // formato incorrecto
		{"", "31-10-2026"},           // formato incorrecto
		{"2026-13-01", ""},           // mes inexistente
		{"2026-10-31", "2026-10-01"}, // desde posterior a hasta
	}
	for _, c := range casos {
		if _, _, err := RangoFechas(c[0], c[1]); !errors.Is(err, ErrRangoInvalido) {
			t.Errorf("%v: se esperaba ErrRangoInvalido y llegó %v", c, err)
		}
	}
	// El mismo día en ambos extremos es válido (un solo día).
	if _, _, err := RangoFechas("2026-10-22", "2026-10-22"); err != nil {
		t.Errorf("un solo día debe ser válido: %v", err)
	}
}

func TestTicketPromedio(t *testing.T) {
	if got := TicketPromedio(100, 3); !casi(got, 33.33) {
		t.Errorf("100/3 = %v", got)
	}
	if got := TicketPromedio(50, 0); got != 0 {
		t.Errorf("sin pedidos el ticket es 0, llegó %v", got)
	}
}
