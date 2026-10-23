// src/data/pedidos.ts
// Estados del pedido y sus transiciones (deben coincidir con el backend).

export type EstadoPedido = 'pendiente' | 'preparando' | 'enviado' | 'entregado' | 'cancelado';

export const ESTADOS: EstadoPedido[] = ['pendiente', 'preparando', 'enviado', 'entregado', 'cancelado'];

export const ETIQUETA_ESTADO: Record<EstadoPedido, string> = {
  pendiente: 'Pendiente',
  preparando: 'Preparando',
  enviado: 'Enviado',
  entregado: 'Entregado',
  cancelado: 'Cancelado',
};

export const COLOR_ESTADO: Record<EstadoPedido, string> = {
  pendiente: 'bg-amber-100 text-amber-800',
  preparando: 'bg-blue-100 text-blue-800',
  enviado: 'bg-indigo-100 text-indigo-800',
  entregado: 'bg-green-100 text-green-800',
  cancelado: 'bg-red-100 text-red-700',
};

// Siguiente paso "normal" del pedido (el admin lo avanza con un botón).
export const SIGUIENTE_ESTADO: Partial<Record<EstadoPedido, EstadoPedido>> = {
  pendiente: 'preparando',
  preparando: 'enviado',
  enviado: 'entregado',
};

// Solo se puede cancelar antes de enviar el pedido.
export const puedeCancelar = (estado: EstadoPedido): boolean =>
  estado === 'pendiente' || estado === 'preparando';
