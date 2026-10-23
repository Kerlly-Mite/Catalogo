import { COLOR_ESTADO, ETIQUETA_ESTADO, type EstadoPedido } from '../data/pedidos';

// Etiqueta de color con el estado del pedido.
const EstadoBadge = ({ estado }: { estado: EstadoPedido }) => (
  <span className={`inline-block px-2.5 py-1 rounded-full text-xs font-semibold ${COLOR_ESTADO[estado] ?? 'bg-slate-100 text-slate-700'}`}>
    {ETIQUETA_ESTADO[estado] ?? estado}
  </span>
);

export default EstadoBadge;
