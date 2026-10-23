package repository

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"multicatalogo-backend/models"
)

// columnasUsuarioListado: el orden coincide con escanearUsuarioListado. Nunca se selecciona el password.
const columnasUsuarioListado = `u.id, u.email, COALESCE(u.nombre, ''), u.rol, u.activo, u.creado_en, COALESCE(r.codigo, '')`

func escanearUsuarioListado(row pgx.Row) (*models.UsuarioListado, error) {
	var u models.UsuarioListado
	var creado time.Time
	if err := row.Scan(&u.ID, &u.Email, &u.Nombre, &u.Rol, &u.Activo, &creado, &u.Codigo); err != nil {
		return nil, err
	}
	u.CreadoEn = creado.Format(time.RFC3339)
	return &u, nil
}

// ListarUsuarios devuelve una página de usuarios (búsqueda ?q= por email o nombre, con ILIKE).
func (r *PostgresRepository) ListarUsuarios(ctx context.Context, q string, page, limit int) ([]models.UsuarioListado, int, error) {
	ctx, cancel := conTimeout(ctx, 5*time.Second)
	defer cancel()

	patron := "%" + escaparLike(strings.TrimSpace(q)) + "%"
	const filtro = `(u.email ILIKE $1 OR COALESCE(u.nombre, '') ILIKE $1)`

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM usuarios u WHERE `+filtro, patron).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx,
		`SELECT `+columnasUsuarioListado+`
		   FROM usuarios u LEFT JOIN referidos r ON r.id = u.referido_id
		  WHERE `+filtro+` ORDER BY u.id LIMIT $2 OFFSET $3`,
		patron, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	lista := []models.UsuarioListado{}
	for rows.Next() {
		u, err := escanearUsuarioListado(rows)
		if err != nil {
			return nil, 0, err
		}
		lista = append(lista, *u)
	}
	return lista, total, rows.Err()
}

// ActualizarUsuario cambia SOLO el rol y/o el estado enviados (COALESCE conserva el valor actual si llega NULL).
// Devuelve ErrNoEncontrado si el id no existe.
func (r *PostgresRepository) ActualizarUsuario(ctx context.Context, id int, c models.UsuarioCambios) (*models.UsuarioListado, error) {
	ctx, cancel := conTimeout(ctx, 5*time.Second)
	defer cancel()

	tag, err := r.db.Exec(ctx,
		`UPDATE usuarios SET rol = COALESCE($2, rol), activo = COALESCE($3, activo) WHERE id = $1`,
		id, c.Rol, c.Activo)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNoEncontrado
	}
	return escanearUsuarioListado(r.db.QueryRow(ctx,
		`SELECT `+columnasUsuarioListado+` FROM usuarios u LEFT JOIN referidos r ON r.id = u.referido_id WHERE u.id = $1`, id))
}
