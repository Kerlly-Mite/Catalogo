import { useEffect, useState } from 'react';
import { getPedido, type PedidoDetalle } from '../services/pedidosService';
import { mensajeDeError } from '../services/http';
import { abrirComprobante } from '../services/comprobantesService';
import { useAuth } from '../context/AuthContext';

// Detalle de un pedido (items con su nombre y datos de envío). Se carga al desplegarlo.
const DetallePedido = ({ id }: { id: number }) => {
  const [pedido, setPedido] = useState<PedidoDetalle | null>(null);
  const [error, setError] = useState<string>('');
  const [errorVoucher, setErrorVoucher] = useState<string>('');
  const { user } = useAuth();

  useEffect(() => {
    let cancelado = false;
    getPedido(id)
      .then((p) => !cancelado && setPedido(p))
      .catch((e) => !cancelado && setError(mensajeDeError(e)));
    return () => {
      cancelado = true;
    };
  }, [id]);

  if (error) return <p className="text-sm text-red-600">{error}</p>;
  if (!pedido) return <p className="text-sm text-slate-500">Cargando detalle...</p>;

  const pago = pedido.pago_detalle ?? {};
  const hayDetalle = Object.keys(pago).length > 0;

  return (
    <div className="space-y-3 text-sm">
      <ul className="divide-y divide-slate-200">
        {pedido.items.map((it) => (
          <li key={it.id} className="py-2 flex justify-between gap-3">
            <span className="text-slate-800">
              {it.nombre} <span className="text-slate-500">× {it.cantidad}</span>
            </span>
            <span className="font-semibold text-slate-800">${(it.cantidad * it.precio).toFixed(2)}</span>
          </li>
        ))}
      </ul>
      <div className="text-slate-600 space-y-1">
        <p>
          <span className="font-semibold">Envío a:</span> {pedido.direccion}, {pedido.ciudad}
          {pedido.telefono ? ` · Tel. ${pedido.telefono}` : ''}
        </p>
        <p>
          <span className="font-semibold">Pago:</span> {pedido.metodo_pago}
        </p>
        {hayDetalle && (
          <ul className="ml-4 list-disc space-y-0.5 break-words">
            {pago.ultimos4 && (
              <li>
                Tarjeta terminada en <span className="font-mono">{pago.ultimos4}</span> · {pago.titular} · vence {pago.vencimiento}
              </li>
            )}
            {pago.banco_origen && (
              <li>
                Transferencia desde {pago.banco_origen} · ref. <span className="font-mono">{pago.referencia}</span>
              </li>
            )}
            {pago.comprobante_id && user?.rol === 'admin' && (
              <li>
                <button
                  type="button"
                  onClick={() => {
                    setErrorVoucher('');
                    abrirComprobante(Number(pago.comprobante_id)).catch((e) => setErrorVoucher(mensajeDeError(e)));
                  }}
                  className="text-indigo-600 font-semibold hover:underline"
                >
                  Ver comprobante
                </button>
                {errorVoucher && <span className="text-red-600"> · {errorVoucher}</span>}
              </li>
            )}
            {pago.recibe && (
              <li>
                Recibe: {pago.recibe} · {pago.franja}
              </li>
            )}
            {pago.paga_con && <li>Paga con ${pago.paga_con}</li>}
            {pago.indicaciones && <li>Indicaciones: {pago.indicaciones}</li>}
          </ul>
        )}
      </div>
    </div>
  );
};

export default DetallePedido;
