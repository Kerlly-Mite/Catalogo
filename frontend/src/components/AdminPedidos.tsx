import { Fragment, useEffect, useState } from 'react';
import {
  cambiarEstadoPedido,
  getPedidosAdmin,
  type Pagina,
  type PedidoListado,
} from '../services/pedidosService';
import { mensajeDeError } from '../services/http';
import {
  ESTADOS,
  ETIQUETA_ESTADO,
  SIGUIENTE_ESTADO,
  puedeCancelar,
  type EstadoPedido,
} from '../data/pedidos';
import EstadoBadge from './EstadoBadge';
import DetallePedido from './DetallePedido';
import Paginacion from './Paginacion';

const LIMITE = 10;

// /admin/pedidos: todos los pedidos con filtro por estado y cambio de estado.
const AdminPedidos = () => {
  const [page, setPage] = useState<number>(1);
  const [estado, setEstado] = useState<EstadoPedido | ''>('');
  const [recarga, setRecarga] = useState<number>(0);
  const [pagina, setPagina] = useState<Pagina<PedidoListado> | null>(null);
  const [error, setError] = useState<string>('');
  const [aviso, setAviso] = useState<string>('');
  const [cargando, setCargando] = useState<boolean>(true);
  const [abierto, setAbierto] = useState<number | null>(null);
  const [trabajando, setTrabajando] = useState<number | null>(null);

  useEffect(() => {
    let cancelado = false;
    getPedidosAdmin(page, LIMITE, estado)
      .then((p) => {
        if (cancelado) return;
        setPagina(p);
        setError('');
      })
      .catch((e) => !cancelado && setError(mensajeDeError(e)))
      .finally(() => !cancelado && setCargando(false));
    return () => {
      cancelado = true;
    };
  }, [page, estado, recarga]);

  const cambiar = async (p: PedidoListado, nuevo: EstadoPedido) => {
    if (nuevo === 'cancelado' && !window.confirm(`¿Cancelar el pedido ${p.numero}? Se devolverá el stock y se revertirán las comisiones.`)) {
      return;
    }
    setTrabajando(p.id);
    setError('');
    setAviso('');
    try {
      await cambiarEstadoPedido(p.id, nuevo);
      setAviso(`Pedido ${p.numero}: ahora está "${ETIQUETA_ESTADO[nuevo]}".`);
      setRecarga((n) => n + 1);
    } catch (e) {
      setError(mensajeDeError(e));
    } finally {
      setTrabajando(null);
    }
  };

  const pedidos = pagina?.data ?? [];
  const boton = 'min-h-[44px] px-3 rounded-lg text-sm font-semibold transition disabled:opacity-50';

  return (
    <div className="max-w-6xl mx-auto">
      <div className="flex flex-wrap items-end justify-between gap-4 mb-6">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold text-slate-900">Pedidos</h1>
          <p className="text-slate-500 mt-1">Gestiona el ciclo de vida de cada pedido.</p>
        </div>
        <label className="text-sm text-slate-700">
          <span className="block mb-1 font-medium">Filtrar por estado</span>
          <select
            value={estado}
            onChange={(e) => {
              setEstado(e.target.value as EstadoPedido | '');
              setPage(1);
            }}
            className="min-h-[44px] px-3 rounded-lg border border-slate-300 bg-white"
          >
            <option value="">Todos</option>
            {ESTADOS.map((s) => (
              <option key={s} value={s}>
                {ETIQUETA_ESTADO[s]}
              </option>
            ))}
          </select>
        </label>
      </div>

      {error && <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-4 border border-red-200">{error}</div>}
      {aviso && <div className="bg-green-50 text-green-700 p-3 rounded-lg text-sm mb-4 border border-green-200">{aviso}</div>}
      {cargando && <p className="text-slate-500">Cargando pedidos...</p>}
      {!cargando && !error && pedidos.length === 0 && <p className="text-slate-500">No hay pedidos con ese filtro.</p>}

      {pedidos.length > 0 && (
        <div className="bg-white border border-slate-200 rounded-xl shadow-sm overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-slate-50 text-slate-600 text-left">
              <tr>
                <th className="p-3">Pedido</th>
                <th className="p-3">Cliente</th>
                <th className="p-3">Total</th>
                <th className="p-3">Estado</th>
                <th className="p-3">Acciones</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-200">
              {pedidos.map((p) => {
                const siguiente = SIGUIENTE_ESTADO[p.estado];
                return (
                  <Fragment key={p.id}>
                    <tr>
                      <td className="p-3 whitespace-nowrap">
                        <button
                          onClick={() => setAbierto(abierto === p.id ? null : p.id)}
                          className="min-h-[44px] font-bold text-indigo-600 hover:underline text-left"
                        >
                          {p.numero}
                        </button>
                        <p className="text-xs text-slate-500">{new Date(p.fecha).toLocaleString('es-EC')}</p>
                      </td>
                      <td className="p-3">
                        <p className="font-medium text-slate-800">{p.cliente_nombre}</p>
                        <p className="text-xs text-slate-500">{p.cliente_email}</p>
                      </td>
                      <td className="p-3 font-bold text-slate-900 whitespace-nowrap">${p.total.toFixed(2)}</td>
                      <td className="p-3">
                        <EstadoBadge estado={p.estado} />
                      </td>
                      <td className="p-3">
                        <div className="flex flex-wrap gap-2">
                          {siguiente && (
                            <button
                              disabled={trabajando === p.id}
                              onClick={() => cambiar(p, siguiente)}
                              className={`${boton} bg-indigo-600 text-white hover:bg-indigo-700`}
                            >
                              Pasar a {ETIQUETA_ESTADO[siguiente]}
                            </button>
                          )}
                          {puedeCancelar(p.estado) && (
                            <button
                              disabled={trabajando === p.id}
                              onClick={() => cambiar(p, 'cancelado')}
                              className={`${boton} border border-red-300 text-red-600 hover:bg-red-50`}
                            >
                              Cancelar
                            </button>
                          )}
                          {!siguiente && !puedeCancelar(p.estado) && <span className="text-slate-400">—</span>}
                        </div>
                      </td>
                    </tr>
                    {abierto === p.id && (
                      <tr>
                        <td colSpan={5} className="p-4 bg-slate-50">
                          <DetallePedido id={p.id} />
                        </td>
                      </tr>
                    )}
                  </Fragment>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {pagina && pagina.total > 0 && (
        <Paginacion page={pagina.page} limit={pagina.limit} total={pagina.total} onChange={setPage} />
      )}
    </div>
  );
};

export default AdminPedidos;
