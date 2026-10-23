// src/components/MiRed.tsx
// MI red multinivel (la del usuario con sesión), su link de invitación y su estado de cuenta de comisiones.
import { useEffect, useState } from 'react';
import { contarRed, sumarVentasRed, nivelAlcanzado, TASA_COMISION, type Referido } from '../data/red';
import { getMiRed, getMisComisiones, type ComisionesMias, type RedMia } from '../services/redService';
import { getMisLiquidaciones, type Liquidacion } from '../services/liquidacionesService';
import { mensajeDeError } from '../services/http';
import Paginacion from './Paginacion';

// Colores y etiquetas por nivel para el árbol
const NIVEL_ESTILO: Record<number, { badge: string; border: string; tasa: string }> = {
  1: { badge: 'bg-indigo-100 text-indigo-700', border: 'border-indigo-200', tasa: '10 %' },
  2: { badge: 'bg-purple-100 text-purple-700', border: 'border-purple-200', tasa: '5 %' },
  3: { badge: 'bg-pink-100 text-pink-700', border: 'border-pink-200', tasa: '2 %' },
};

const ESTILO_COMISION: Record<string, string> = {
  pendiente: 'bg-amber-100 text-amber-800',
  aprobada: 'bg-blue-100 text-blue-800',
  pagada: 'bg-green-100 text-green-800',
  revertida: 'bg-red-100 text-red-700',
};

const LIMITE_COMISIONES = 5;

// Componente recursivo: renderiza un referido y, debajo, a sus hijos.
// Nota: la prop se llama "nodo" porque "ref" es una prop reservada en React 19
const NodoReferido = ({ nodo }: { nodo: Referido }) => {
  const estilo = NIVEL_ESTILO[nodo.nivel] ?? { badge: 'bg-slate-100 text-slate-600', border: 'border-slate-200', tasa: 'sin comisión' };

  return (
    <li>
      <div className={`bg-white rounded-lg border ${estilo.border} shadow-sm p-4 flex flex-wrap items-center justify-between gap-3`}>
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-full bg-slate-200 flex items-center justify-center font-bold text-slate-600">
            {nodo.nombre.charAt(0)}
          </div>
          <div>
            <p className="font-semibold text-slate-800">{nodo.nombre}</p>
            <span className={`inline-block mt-1 text-xs font-semibold px-2 py-0.5 rounded-full ${estilo.badge}`}>
              Nivel {nodo.nivel} · {estilo.tasa}
            </span>
          </div>
        </div>
        <div className="text-right">
          <p className="text-sm text-slate-500">Ventas</p>
          <p className="font-bold text-slate-800">${nodo.ventas.toLocaleString()}</p>
        </div>
      </div>

      {nodo.hijos && nodo.hijos.length > 0 && (
        <ul className="mt-3 ml-6 space-y-3 border-l-2 border-slate-200 pl-4">
          {nodo.hijos.map((hijo) => (
            <NodoReferido key={hijo.id} nodo={hijo} />
          ))}
        </ul>
      )}
    </li>
  );
};

const MiRed = () => {
  const [miRed, setMiRed] = useState<RedMia | null>(null);
  const [errorRed, setErrorRed] = useState<string>('');
  const [cargandoRed, setCargandoRed] = useState<boolean>(true);

  const [pageC, setPageC] = useState<number>(1);
  const [comisiones, setComisiones] = useState<ComisionesMias | null>(null);
  const [errorCom, setErrorCom] = useState<string>('');

  const [liquidaciones, setLiquidaciones] = useState<Liquidacion[]>([]);
  const [copiado, setCopiado] = useState<boolean>(false);

  useEffect(() => {
    let cancelado = false;
    getMiRed()
      .then((r) => !cancelado && setMiRed(r))
      .catch((e) => !cancelado && setErrorRed(mensajeDeError(e)))
      .finally(() => !cancelado && setCargandoRed(false));
    getMisLiquidaciones(1, 5)
      .then((p) => !cancelado && setLiquidaciones(p.data))
      .catch(() => undefined); // los cobros son informativos: si fallan no bloquean la pantalla
    return () => {
      cancelado = true;
    };
  }, []);

  useEffect(() => {
    let cancelado = false;
    getMisComisiones(pageC, LIMITE_COMISIONES)
      .then((c) => {
        if (cancelado) return;
        setComisiones(c);
        setErrorCom('');
      })
      .catch((e) => !cancelado && setErrorCom(mensajeDeError(e)));
    return () => {
      cancelado = true;
    };
  }, [pageC]);

  const link = miRed ? `${window.location.origin}/registro?ref=${miRed.codigo_referido}` : '';

  const copiar = async () => {
    try {
      await navigator.clipboard.writeText(link);
      setCopiado(true);
      setTimeout(() => setCopiado(false), 2000);
    } catch {
      window.prompt('Copia tu enlace de referido:', link);
    }
  };

  if (cargandoRed) return <p className="text-slate-500">Cargando tu red...</p>;
  if (errorRed || !miRed) {
    return <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm border border-red-200">{errorRed || 'No se pudo cargar tu red'}</div>;
  }

  const red = miRed.red;
  const directos = red.hijos?.length ?? 0;
  const nivel = nivelAlcanzado(directos);
  const totales = comisiones?.totales;
  const ganado = (totales?.pendiente ?? 0) + (totales?.aprobada ?? 0) + (totales?.pagada ?? 0);

  const resumen = [
    { etiqueta: 'Referidos Activos', valor: contarRed(red).toString(), color: 'text-indigo-600' },
    { etiqueta: 'Ventas de la Red', valor: `$${sumarVentasRed(red).toLocaleString()}`, color: 'text-indigo-600' },
    { etiqueta: 'Comisiones Ganadas', valor: `$${ganado.toFixed(2)}`, color: 'text-green-600' },
    { etiqueta: 'Nivel Alcanzado', valor: nivel, color: 'text-amber-600' },
  ];

  return (
    <div className="max-w-5xl mx-auto">
      <h1 className="text-2xl font-bold text-slate-800 mb-2">Mi Red de Referidos</h1>
      <p className="text-slate-500 mb-6">
        Estructura multinivel: gana el {TASA_COMISION[1] * 100} % de nivel 1, {TASA_COMISION[2] * 100} % de nivel 2 y {TASA_COMISION[3] * 100} % de nivel 3.
      </p>

      {/* ===== MI LINK DE REFERIDO ===== */}
      <div className="bg-white rounded-lg border border-slate-200 shadow-sm p-4 sm:p-6 mb-8">
        <h2 className="text-lg font-bold text-slate-800 mb-1">Tu enlace de invitación</h2>
        <p className="text-sm text-slate-500 mb-3">
          Quien se registre con este enlace entra a tu red. Tu código: <span className="font-mono font-bold text-slate-800">{miRed.codigo_referido}</span>
        </p>
        <div className="flex flex-col sm:flex-row gap-2">
          <input readOnly value={link} className="flex-1 min-h-[44px] px-3 rounded-lg border border-slate-300 bg-slate-50 text-sm font-mono" />
          <button onClick={copiar} className="min-h-[44px] px-5 bg-indigo-600 text-white font-semibold rounded-lg hover:bg-indigo-700 transition">
            {copiado ? '¡Copiado!' : 'Copiar'}
          </button>
        </div>
      </div>

      {/* ===== TARJETAS DE RESUMEN ===== */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {resumen.map((item) => (
          <div key={item.etiqueta} className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
            <p className="text-sm text-slate-500 uppercase font-semibold">{item.etiqueta}</p>
            <p className={`text-3xl font-bold mt-2 ${item.color}`}>{item.valor}</p>
          </div>
        ))}
      </div>

      {/* ===== ÁRBOL DE REFERIDOS ===== */}
      <div className="bg-white rounded-lg border border-slate-200 shadow-sm p-4 sm:p-6 mb-8">
        <h2 className="text-lg font-bold text-slate-800 mb-4">Jerarquía de tu Red</h2>
        <div className="bg-slate-900 text-white rounded-lg p-4 mb-4 flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-indigo-500 flex items-center justify-center font-bold">{red.nombre.charAt(0)}</div>
            <div>
              <p className="font-semibold">{red.nombre}</p>
              <p className="text-xs text-slate-300">Nivel alcanzado: {nivel}</p>
            </div>
          </div>
          <p className="font-bold text-green-400">Saldo acumulado: ${red.saldo?.toFixed?.(2) ?? '0.00'}</p>
        </div>

        {directos === 0 ? (
          <p className="text-slate-500 text-sm">Todavía no tienes referidos. ¡Comparte tu enlace para empezar a crecer!</p>
        ) : (
          <ul className="space-y-3">
            {red.hijos?.map((hijo) => (
              <NodoReferido key={hijo.id} nodo={hijo} />
            ))}
          </ul>
        )}
      </div>

      {/* ===== MIS COMISIONES (ledger) ===== */}
      <div className="bg-white rounded-lg border border-slate-200 shadow-sm p-4 sm:p-6 mb-8">
        <h2 className="text-lg font-bold text-slate-800 mb-4">Mis comisiones</h2>

        {errorCom && <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-4 border border-red-200">{errorCom}</div>}

        {totales && (
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-4">
            {(['pendiente', 'pagada', 'revertida'] as const).map((estado) => (
              <div key={estado} className="rounded-lg border border-slate-200 p-3">
                <span className={`text-xs font-semibold px-2 py-0.5 rounded-full ${ESTILO_COMISION[estado]}`}>{estado}</span>
                <p className="text-xl font-bold text-slate-800 mt-2">${(totales[estado] ?? 0).toFixed(2)}</p>
              </div>
            ))}
          </div>
        )}

        {comisiones && comisiones.data.length === 0 ? (
          <p className="text-slate-500 text-sm">Aún no tienes comisiones. Se generan cuando alguien de tu red compra.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="text-left text-slate-600 bg-slate-50">
                <tr>
                  <th className="p-2">Pedido</th>
                  <th className="p-2">Nivel</th>
                  <th className="p-2">%</th>
                  <th className="p-2">Monto</th>
                  <th className="p-2">Estado</th>
                  <th className="p-2">Fecha</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-200">
                {comisiones?.data.map((c) => (
                  <tr key={c.id}>
                    <td className="p-2 font-medium">{c.numero}</td>
                    <td className="p-2">{c.nivel}</td>
                    <td className="p-2">{c.porcentaje} %</td>
                    <td className="p-2 font-bold">${c.monto.toFixed(2)}</td>
                    <td className="p-2">
                      <span className={`text-xs font-semibold px-2 py-0.5 rounded-full ${ESTILO_COMISION[c.estado] ?? ''}`}>{c.estado}</span>
                    </td>
                    <td className="p-2 whitespace-nowrap text-slate-500">{new Date(c.fecha).toLocaleDateString('es-EC')}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {comisiones && comisiones.total > 0 && (
          <Paginacion page={comisiones.page} limit={comisiones.limit} total={comisiones.total} onChange={setPageC} />
        )}
      </div>

      {/* ===== MIS COBROS ===== */}
      <div className="bg-white rounded-lg border border-slate-200 shadow-sm p-4 sm:p-6">
        <h2 className="text-lg font-bold text-slate-800 mb-4">Mis cobros</h2>
        {liquidaciones.length === 0 ? (
          <p className="text-slate-500 text-sm">Todavía no se te ha liquidado ninguna comisión.</p>
        ) : (
          <ul className="divide-y divide-slate-200">
            {liquidaciones.map((l) => (
              <li key={l.id} className="py-3 flex justify-between gap-3 text-sm">
                <span className="text-slate-600">{new Date(l.pagada_en || l.creado_en).toLocaleDateString('es-EC')} · Liquidación #{l.id}</span>
                <span className="font-bold text-green-700">${l.total.toFixed(2)}</span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
};

export default MiRed;
