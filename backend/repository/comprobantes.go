package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// GuardarComprobante guarda el voucher de una transferencia (imagen o PDF) y devuelve su id.
// El tipo ya viene validado por el controlador.
func (r *PostgresRepository) GuardarComprobante(ctx context.Context, usuarioID int, tipo string, datos []byte) (int, error) {
	ctx, cancel := conTimeout(ctx, 10*time.Second)
	defer cancel()

	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO comprobantes (usuario_id, tipo, datos, tamano) VALUES ($1, $2, $3, $4) RETURNING id`,
		usuarioID, tipo, datos, len(datos)).Scan(&id)
	return id, err
}

// ObtenerComprobante devuelve el tipo y los bytes de un voucher, o ErrNoEncontrado.
func (r *PostgresRepository) ObtenerComprobante(ctx context.Context, id int) (string, []byte, error) {
	ctx, cancel := conTimeout(ctx, 10*time.Second)
	defer cancel()

	var tipo string
	var datos []byte
	err := r.db.QueryRow(ctx, `SELECT tipo, datos FROM comprobantes WHERE id = $1`, id).Scan(&tipo, &datos)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil, ErrNoEncontrado
	}
	return tipo, datos, err
}
