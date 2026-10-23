import { useEffect, useState } from 'react';
import { getPedido, type PedidoDetalle } from '../services/pedidosService';
import { mensajeDeError } from '../services/http';

// Detalle de un pedido (items con su nombre y datos de envío). Se carga al desplegarlo.
const DetallePedido = ({ id }: { id: number }) => {
  const [pedido, setPedido] = useState<PedidoDetalle | null>(null);
  const [error, setError] = useState<string>('');

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
      </div>
    </div>
  );
};

export default DetallePedido;
