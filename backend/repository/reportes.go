package repository

import (
	"context"
	"time"

	"multicatalogo-backend/domain"
	"multicatalogo-backend/models"
)

// Filtro de fechas común a los reportes: $1 = desde (incluido) y $2 = hasta (excluido), NULL = sin límite.
// Se aplica sobre la fecha del PEDIDO, así una comisión pertenece al periodo de la venta que la generó.
const filtroFechaPedido = `($1::timestamptz IS NULL OR p.fecha >= $1) AND ($2::timestamptz IS NULL OR p.fecha < $2)`

// ResumenVentas calcula los KPIs del periodo con agregaciones SQL (COUNT/SUM ... FILTER).
func (r *PostgresRepository) ResumenVentas(ctx context.Context, desde, hasta *time.Time) (*models.ReporteResumen, error) {
	ctx, cancel := conTimeout(ctx, 5*time.Second)
	defer cancel()

	var res models.ReporteResumen

	// Ventas: los pedidos cancelados se cuentan aparte y NO suman al total vendido.
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FILTER (WHERE p.estado <> 'cancelado'),
		        COALESCE(SUM(p.total) FILTER (WHERE p.estado <> 'cancelado'), 0),
		        COUNT(*) FILTER (WHERE p.estado = 'cancelado')
		   FROM pedidos p
		  WHERE `+filtroFechaPedido, desde, hasta,
	).Scan(&res.Pedidos, &res.TotalVendido, &res.PedidosCancelados); err != nil {
		return nil, err
	}

	// Comisiones del ledger de esos mismos pedidos: devengada = todo lo que no se revirtió.
	if err := r.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(c.monto) FILTER (WHERE c.estado <> 'revertida'), 0),
		        COALESCE(SUM(c.monto) FILTER (WHERE c.estado = 'pendiente'), 0),
		        COALESCE(SUM(c.monto) FILTER (WHERE c.estado = 'pagada'), 0)
		   FROM comisiones c JOIN pedidos p ON p.id = c.pedido_id
		  WHERE `+filtroFechaPedido, desde, hasta,
	).Scan(&res.ComisionesDevengadas, &res.ComisionesPendientes, &res.ComisionesPagadas); err != nil {
		return nil, err
	}

	// Tamaño de la red (no depende del periodo): todos los nodos menos la raíz.
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM referidos WHERE parent_id IS NOT NULL`).Scan(&res.Distribuidores); err != nil {
		return nil, err
	}

	res.TicketPromedio = domain.TicketPromedio(res.TotalVendido, res.Pedidos)
	return &res, nil
}

// TopProductos: ranking por ingresos (SUM cantidad×precio_unitario) agrupado por producto, sin pedidos cancelados.
func (r *PostgresRepository) TopProductos(ctx context.Context, desde, hasta *time.Time, limite int) ([]models.TopProducto, error) {
	ctx, cancel := conTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.Query(ctx,
		`SELECT pr.id, pr.nombre, SUM(pi.cantidad), SUM(pi.cantidad * pi.precio_unitario)
		   FROM pedido_items pi
		   JOIN pedidos p    ON p.id  = pi.pedido_id
		   JOIN productos pr ON pr.id = pi.producto_id
		  WHERE p.estado <> 'cancelado' AND `+filtroFechaPedido+`
		  GROUP BY pr.id, pr.nombre
		  ORDER BY SUM(pi.cantidad * pi.precio_unitario) DESC, pr.id
		  LIMIT $3`, desde, hasta, limite)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lista := []models.TopProducto{} // vacío, no nil: el JSON es [] y el front muestra el estado vacío
	for rows.Next() {
		var t models.TopProducto
		if err := rows.Scan(&t.ID, &t.Nombre, &t.Unidades, &t.Ingresos); err != nil {
			return nil, err
		}
		lista = append(lista, t)
	}
	return lista, rows.Err()
}

// TopReferidos: ranking de distribuidores por comisión devengada. Sale del LEDGER (fuente de verdad), no del
// contador referidos.ventas. Cada fila del ledger es "este distribuidor cobró por este pedido", así que
// SUM(p.total) son las ventas de su red y COUNT(*) los pedidos que le generaron comisión.
func (r *PostgresRepository) TopReferidos(ctx context.Context, desde, hasta *time.Time, limite int) ([]models.TopReferido, error) {
	ctx, cancel := conTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.Query(ctx,
		`SELECT r.id, r.nombre, COUNT(*), SUM(p.total), SUM(c.monto)
		   FROM comisiones c
		   JOIN referidos r ON r.id = c.referido_id
		   JOIN pedidos p   ON p.id = c.pedido_id
		  WHERE c.estado <> 'revertida' AND `+filtroFechaPedido+`
		  GROUP BY r.id, r.nombre
		  ORDER BY SUM(c.monto) DESC, r.id
		  LIMIT $3`, desde, hasta, limite)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lista := []models.TopReferido{}
	for rows.Next() {
		var t models.TopReferido
		if err := rows.Scan(&t.ReferidoID, &t.Nombre, &t.Pedidos, &t.VentasRed, &t.Comisiones); err != nil {
			return nil, err
		}
		lista = append(lista, t)
	}
	return lista, rows.Err()
}
