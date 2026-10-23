import { useEffect, useState } from 'react';
import {
  getDeuda,
  getLiquidaciones,
  liquidar,
  type DeudaReferido,
  type Liquidacion,
} from '../services/liquidacionesService';
import type { Pagina } from '../services/pedidosService';
import { mensajeDeError } from '../services/http';
import Paginacion from './Paginacion';

const LIMITE = 10;

// /admin/liquidaciones: lo que se le debe a cada distribuidor + historial de pagos.
const AdminLiquidaciones = () => {
  const [deuda, setDeuda] = useState<DeudaReferido[]>([]);
  const [historial, setHistorial] = useState<Pagina<Liquidacion> | null>(null);
  const [page, setPage] = useState<number>(1);
  const [recarga, setRecarga] = useState<number>(0);
  const [error, setError] = useState<string>('');
  const [aviso, setAviso] = useState<string>('');
  const [cargando, setCargando] = useState<boolean>(true);
  const [pagando, setPagando] = useState<number | null>(null);

  useEffect(() => {
    let cancelado = false;
    Promise.all([getDeuda(), getLiquidaciones(page, LIMITE)])
      .then(([d, h]) => {
        if (cancelado) return;
        setDeuda(d);
        setHistorial(h);
        setError('');
      })
      .catch((e) => !cancelado && setError(mensajeDeError(e)))
      .finally(() => !cancelado && setCargando(false));
    return () => {
      cancelado = true;
    };
  }, [page, recarga]);

  const pagar = async (d: DeudaReferido) => {
    if (!window.confirm(`¿Liquidar $${d.pendiente.toFixed(2)} a ${d.nombre}? Las ${d.comisiones} comisión(es) pasarán a "pagada".`)) return;
    setPagando(d.referido_id);
    setError('');
    setAviso('');
    try {
      const liq = await liquidar(d.referido_id);
      setAviso(`Liquidación #${liq.id}: se pagaron $${liq.total.toFixed(2)} a ${liq.referido}.`);
      setPage(1);
      setRecarga((n) => n + 1);
    } catch (e) {
      setError(mensajeDeError(e));
    } finally {
      setPagando(null);
    }
  };

  const totalDeuda = deuda.reduce((suma, d) => suma + d.pendiente, 0);

  return (
    <div className="max-w-5xl mx-auto">
      <h1 className="text-2xl sm:text-3xl font-bold text-slate-900">Liquidaciones</h1>
      <p className="text-slate-500 mt-1 mb-6">Paga a los distribuidores las comisiones pendientes de su estado de cuenta.</p>

      {error && <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-4 border border-red-200">{error}</div>}
      {aviso && <div className="bg-green-50 text-green-700 p-3 rounded-lg text-sm mb-4 border border-green-200">{aviso}</div>}
      {cargando && <p className="text-slate-500">Cargando...</p>}

      <div className="bg-white border border-slate-200 rounded-xl shadow-sm p-4 sm:p-6 mb-8">
        <div className="flex flex-wrap items-center justify-between gap-2 mb-4">
          <h2 className="text-lg font-bold text-slate-800">Deuda por distribuidor</h2>
          <span className="text-sm text-slate-600">
            Total pendiente: <span className="font-bold text-amber-700">${totalDeuda.toFixed(2)}</span>
          </span>
        </div>
        {!cargando && deuda.length === 0 ? (
          <p className="text-slate-500 text-sm">No hay comisiones pendientes de pago.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-slate-50 text-slate-600 text-left">
                <tr>
                  <th className="p-3">Distribuidor</th>
                  <th className="p-3">Comisiones</th>
                  <th className="p-3">Pendiente</th>
                  <th className="p-3">Acción</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-200">
                {deuda.map((d) => (
                  <tr key={d.referido_id}>
                    <td className="p-3 font-medium">{d.nombre}</td>
                    <td className="p-3">{d.comisiones}</td>
                    <td className="p-3 font-bold">${d.pendiente.toFixed(2)}</td>
                    <td className="p-3">
                      <button
                        disabled={pagando === d.referido_id}
                        onClick={() => pagar(d)}
                        className="min-h-[44px] px-4 bg-indigo-600 text-white font-semibold rounded-lg hover:bg-indigo-700 transition disabled:opacity-50"
                      >
                        {pagando === d.referido_id ? 'Pagando...' : 'Liquidar'}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <div className="bg-white border border-slate-200 rounded-xl shadow-sm p-4 sm:p-6">
        <h2 className="text-lg font-bold text-slate-800 mb-4">Historial de pagos</h2>
        {historial && historial.data.length === 0 ? (
          <p className="text-slate-500 text-sm">Todavía no se ha liquidado nada.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-slate-50 text-slate-600 text-left">
                <tr>
                  <th className="p-3">#</th>
                  <th className="p-3">Distribuidor</th>
                  <th className="p-3">Total</th>
                  <th className="p-3">Estado</th>
                  <th className="p-3">Fecha</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-200">
                {historial?.data.map((l) => (
                  <tr key={l.id}>
                    <td className="p-3">{l.id}</td>
                    <td className="p-3 font-medium">{l.referido}</td>
                    <td className="p-3 font-bold text-green-700">${l.total.toFixed(2)}</td>
                    <td className="p-3">
                      <span className="text-xs font-semibold px-2 py-0.5 rounded-full bg-green-100 text-green-800">{l.estado}</span>
                    </td>
                    <td className="p-3 whitespace-nowrap text-slate-500">{new Date(l.pagada_en || l.creado_en).toLocaleString('es-EC')}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        {historial && historial.total > 0 && (
          <Paginacion page={historial.page} limit={historial.limit} total={historial.total} onChange={setPage} />
        )}
      </div>
    </div>
  );
};

export default AdminLiquidaciones;
