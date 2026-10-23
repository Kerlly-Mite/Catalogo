package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// GuardarImagen guarda una foto (bytes) y devuelve su id. El tipo ya viene validado por el controlador.
func (r *PostgresRepository) GuardarImagen(ctx context.Context, tipo string, datos []byte) (int, error) {
	ctx, cancel := conTimeout(ctx, 10*time.Second)
	defer cancel()

	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO imagenes (tipo, datos, tamano) VALUES ($1, $2, $3) RETURNING id`,
		tipo, datos, len(datos)).Scan(&id)
	return id, err
}

// ObtenerImagen devuelve el tipo y los bytes de una foto, o ErrNoEncontrado.
func (r *PostgresRepository) ObtenerImagen(ctx context.Context, id int) (string, []byte, error) {
	ctx, cancel := conTimeout(ctx, 10*time.Second)
	defer cancel()

	var tipo string
	var datos []byte
	err := r.db.QueryRow(ctx, `SELECT tipo, datos FROM imagenes WHERE id = $1`, id).Scan(&tipo, &datos)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil, ErrNoEncontrado
	}
	return tipo, datos, err
}
