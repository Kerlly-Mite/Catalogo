import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { getMisPedidos, type Pagina, type PedidoListado } from '../services/pedidosService';
import { mensajeDeError } from '../services/http';
import EstadoBadge from './EstadoBadge';
import DetallePedido from './DetallePedido';
import Paginacion from './Paginacion';

const LIMITE = 10;

// /mis-pedidos: historial de compras del usuario con sesión iniciada.
const MisPedidos = () => {
  const [page, setPage] = useState<number>(1);
  const [pagina, setPagina] = useState<Pagina<PedidoListado> | null>(null);
  const [error, setError] = useState<string>('');
  const [cargando, setCargando] = useState<boolean>(true);
  const [abierto, setAbierto] = useState<number | null>(null);

  useEffect(() => {
    let cancelado = false;
    getMisPedidos(page, LIMITE)
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
  }, [page]);

  const pedidos = pagina?.data ?? [];

  return (
    <div className="max-w-4xl mx-auto">
      <h1 className="text-2xl sm:text-3xl font-bold text-slate-900">Mis pedidos</h1>
      <p className="text-slate-500 mt-1 mb-6">Consulta el estado de tus compras.</p>

      {error && <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-4 border border-red-200">{error}</div>}
      {cargando && <p className="text-slate-500">Cargando pedidos...</p>}

      {!cargando && !error && pedidos.length === 0 && (
        <div className="bg-white border border-slate-200 rounded-xl p-8 text-center">
          <p className="text-slate-600 mb-4">Todavía no tienes pedidos.</p>
          <Link to="/catalogo" className="inline-block bg-indigo-600 text-white font-semibold px-5 py-3 rounded-lg hover:bg-indigo-700 transition">
            Ir al catálogo
          </Link>
        </div>
      )}

      <div className="space-y-3">
        {pedidos.map((p) => (
          <div key={p.id} className="bg-white border border-slate-200 rounded-xl shadow-sm">
            <button
              onClick={() => setAbierto(abierto === p.id ? null : p.id)}
              className="w-full min-h-[44px] text-left p-4 flex flex-wrap items-center justify-between gap-3"
            >
              <div>
                <p className="font-bold text-slate-900">{p.numero}</p>
                <p className="text-xs text-slate-500">
                  {new Date(p.fecha).toLocaleString('es-EC')} · {p.num_items} producto(s)
                </p>
              </div>
              <div className="flex items-center gap-3">
                <EstadoBadge estado={p.estado} />
                <span className="font-bold text-slate-900">${p.total.toFixed(2)}</span>
              </div>
            </button>
            {abierto === p.id && (
              <div className="border-t border-slate-200 p-4 bg-slate-50 rounded-b-xl">
                <DetallePedido id={p.id} />
              </div>
            )}
          </div>
        ))}
      </div>

      {pagina && pagina.total > 0 && (
        <Paginacion page={pagina.page} limit={pagina.limit} total={pagina.total} onChange={setPage} />
      )}
    </div>
  );
};

export default MisPedidos;
