// src/components/Dashboard.tsx
// Panel del administrador con datos REALES del negocio (Fase 6): cada cifra viene de /api/reportes/*,
// calculada con SQL de agregación sobre pedidos y el ledger de comisiones. Nada se calcula con datos fijos.
import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  getComisionesPendientes,
  getResumen,
  getTopProductos,
  getTopReferidos,
  type ComisionesPendientesReporte,
  type ReporteResumen,
  type TopProducto,
  type TopReferido,
} from '../services/reportesService';
import { mensajeDeError } from '../services/http';

const dinero = (n: number): string =>
  `$${n.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

// Todo lo que muestra el panel para un periodo. `clave` identifica el periodo con el que se pidió.
interface Datos {
  clave: string;
  resumen: ReporteResumen;
  productos: TopProducto[];
  referidos: TopReferido[];
  deuda: ComisionesPendientesReporte;
}

const campoFecha =
  'min-h-[44px] w-full px-3 rounded-lg border border-slate-300 bg-white text-slate-800 focus:outline-none focus:ring-2 focus:ring-indigo-500';

const Dashboard = () => {
  const [desde, setDesde] = useState<string>('');
  const [hasta, setHasta] = useState<string>('');
  const [datos, setDatos] = useState<Datos | null>(null);
  const [fallo, setFallo] = useState<{ clave: string; mensaje: string } | null>(null);

  const clave = `${desde}|${hasta}`;
  const rangoValido = !desde || !hasta || desde <= hasta;
  // "Cargando" = todavía no llegó la respuesta del periodo que se está mirando.
  const cargando = rangoValido && datos?.clave !== clave && fallo?.clave !== clave;

  useEffect(() => {
    if (!rangoValido) return;
    let cancelado = false;
    Promise.all([
      getResumen(desde, hasta),
      getTopProductos(desde, hasta),
      getTopReferidos(desde, hasta),
      getComisionesPendientes(), // la deuda es de hoy: no depende del periodo
    ])
      .then(([resumen, productos, referidos, deuda]) => {
        if (cancelado) return;
        setDatos({ clave, resumen, productos, referidos, deuda });
        setFallo(null);
      })
      .catch((e) => !cancelado && setFallo({ clave, mensaje: mensajeDeError(e) }));
    return () => {
      cancelado = true;
    };
  }, [desde, hasta, clave, rangoValido]);

  const limpiar = () => {
    setDesde('');
    setHasta('');
  };

  const r = datos?.resumen;
  const kpis = r
    ? [
        { etiqueta: 'Total vendido', valor: dinero(r.total_vendido), nota: 'Sin pedidos cancelados', color: 'text-indigo-600' },
        { etiqueta: 'Pedidos', valor: r.pedidos.toString(), nota: `${r.pedidos_cancelados} cancelado(s)`, color: 'text-indigo-600' },
        { etiqueta: 'Ticket promedio', valor: dinero(r.ticket_promedio), nota: 'Total ÷ pedidos', color: 'text-indigo-600' },
        { etiqueta: 'Comisiones devengadas', valor: dinero(r.comisiones_devengadas), nota: `${dinero(r.comisiones_pagadas)} ya pagadas`, color: 'text-green-600' },
        { etiqueta: 'Comisiones por pagar', valor: dinero(r.comisiones_pendientes), nota: 'Pendientes de liquidar', color: 'text-amber-600' },
        { etiqueta: 'Distribuidores en la red', valor: r.distribuidores.toString(), nota: 'Sin contar la raíz', color: 'text-slate-700' },
      ]
    : [];

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-800 mb-1">Resumen General</h1>
      <p className="text-slate-500 text-sm mb-6">Ventas y comisiones reales del negocio, calculadas desde la base de datos.</p>

      {/* ===== FILTRO POR PERIODO (opcional: vacío = todo el historial) ===== */}
      <div className="bg-white p-4 rounded-lg shadow-sm border border-slate-100 mb-6">
        <div className="grid grid-cols-1 sm:grid-cols-[1fr_1fr_auto] gap-3 items-end">
          <label className="block text-sm font-semibold text-slate-600">
            Desde
            <input type="date" value={desde} max={hasta || undefined} onChange={(e) => setDesde(e.target.value)} className={`${campoFecha} mt-1`} />
          </label>
          <label className="block text-sm font-semibold text-slate-600">
            Hasta
            <input type="date" value={hasta} min={desde || undefined} onChange={(e) => setHasta(e.target.value)} className={`${campoFecha} mt-1`} />
          </label>
          <button
            type="button"
            onClick={limpiar}
            disabled={!desde && !hasta}
            className="min-h-[44px] px-4 rounded-lg border border-slate-300 text-slate-700 font-semibold hover:bg-slate-50 transition disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Ver todo
          </button>
        </div>
        <p className="text-xs text-slate-500 mt-2">
          {desde || hasta ? `Periodo: ${desde || 'inicio'} → ${hasta || 'hoy'} (ambos días incluidos).` : 'Mostrando todo el historial.'}
        </p>
        {!rangoValido && <p className="text-sm text-red-600 mt-2">La fecha "Desde" no puede ser posterior a "Hasta".</p>}
      </div>

      {fallo?.clave === clave && (
        <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-4 border border-red-200">{fallo.mensaje}</div>
      )}
      {cargando && <p className="text-slate-500 mb-4">Cargando reportes...</p>}

      {/* ===== KPIs ===== */}
      {r && (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6 mb-8">
          {kpis.map((kpi) => (
            <div key={kpi.etiqueta} className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
              <p className="text-sm text-slate-500 uppercase font-semibold">{kpi.etiqueta}</p>
              <p className={`text-3xl font-bold mt-2 ${kpi.color}`}>{kpi.valor}</p>
              <p className="text-xs text-slate-400 mt-1">{kpi.nota}</p>
            </div>
          ))}
        </div>
      )}

      {datos && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* ===== TOP PRODUCTOS ===== */}
          <section className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
            <h2 className="text-lg font-bold text-slate-800 mb-4">Productos más vendidos</h2>
            {datos.productos.length === 0 ? (
              <p className="text-slate-500 text-sm">No hay ventas en este periodo.</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead className="bg-slate-50 text-slate-600 text-left">
                    <tr>
                      <th className="p-3">#</th>
                      <th className="p-3">Producto</th>
                      <th className="p-3 text-right">Unidades</th>
                      <th className="p-3 text-right">Ingresos</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-200">
                    {datos.productos.map((p, i) => (
                      <tr key={p.id}>
                        <td className="p-3 text-slate-500">{i + 1}</td>
                        <td className="p-3 font-medium">{p.nombre}</td>
                        <td className="p-3 text-right">{p.unidades}</td>
                        <td className="p-3 text-right font-bold">{dinero(p.ingresos)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>

          {/* ===== TOP DISTRIBUIDORES ===== */}
          <section className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
            <h2 className="text-lg font-bold text-slate-800 mb-4">Distribuidores con más comisión</h2>
            {datos.referidos.length === 0 ? (
              <p className="text-slate-500 text-sm">Ninguna venta generó comisiones en este periodo.</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead className="bg-slate-50 text-slate-600 text-left">
                    <tr>
                      <th className="p-3">#</th>
                      <th className="p-3">Distribuidor</th>
                      <th className="p-3 text-right">Pedidos</th>
                      <th className="p-3 text-right">Ventas de su red</th>
                      <th className="p-3 text-right">Comisión</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-200">
                    {datos.referidos.map((d, i) => (
                      <tr key={d.referido_id}>
                        <td className="p-3 text-slate-500">{i + 1}</td>
                        <td className="p-3 font-medium">{d.nombre}</td>
                        <td className="p-3 text-right">{d.pedidos}</td>
                        <td className="p-3 text-right">{dinero(d.ventas_red)}</td>
                        <td className="p-3 text-right font-bold text-green-700">{dinero(d.comisiones)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>

          {/* ===== DEUDA DE COMISIONES ===== */}
          <section className="bg-white p-6 rounded-lg shadow-sm border border-slate-100 lg:col-span-2">
            <div className="flex flex-wrap items-center justify-between gap-2 mb-4">
              <h2 className="text-lg font-bold text-slate-800">Comisiones pendientes de pago</h2>
              <span className="text-sm text-slate-600">
                Total adeudado: <span className="font-bold text-amber-700">{dinero(datos.deuda.total_pendiente)}</span>
              </span>
            </div>
            {datos.deuda.data.length === 0 ? (
              <p className="text-slate-500 text-sm">No se le debe nada a ningún distribuidor.</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead className="bg-slate-50 text-slate-600 text-left">
                    <tr>
                      <th className="p-3">Distribuidor</th>
                      <th className="p-3 text-right">Comisiones</th>
                      <th className="p-3 text-right">Pendiente</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-200">
                    {datos.deuda.data.map((d) => (
                      <tr key={d.referido_id}>
                        <td className="p-3 font-medium">{d.nombre}</td>
                        <td className="p-3 text-right">{d.comisiones}</td>
                        <td className="p-3 text-right font-bold">{dinero(d.pendiente)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
            <Link to="/admin/liquidaciones" className="inline-flex items-center min-h-[44px] mt-4 text-indigo-600 font-semibold hover:underline">
              Ir a liquidaciones →
            </Link>
          </section>
        </div>
      )}
    </div>
  );
};

export default Dashboard;
