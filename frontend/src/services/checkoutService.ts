// src/services/checkoutService.ts
// Compra real: POST /api/checkout (transacción ACID en el backend). El servidor recalcula precios y total,
// descuenta el stock y toma la identidad del TOKEN (el email del formulario es solo un dato de contacto).
import { apiFetch } from './http';

export interface DatosCliente {
  nombre: string;
  email: string;
  direccion: string;
  ciudad: string;
  telefono: string;
}

export interface ItemCheckout {
  id: number;
  cantidad: number;
}

export interface SolicitudCheckout {
  cliente: DatosCliente;
  metodoPago: string;
  /** Datos del método elegido. Tarjeta: solo titular, últimos 4 y vencimiento (NUNCA el número completo ni el CVV). */
  pagoDetalle: Record<string, string>;
  items: ItemCheckout[];
}

/** Respuesta 201 del backend: el pedido tal como quedó guardado (precios y total del SERVIDOR). */
export interface PedidoCreado {
  pedido_id: number;
  numero: string;
  total: number;
  estado: string;
  fecha: string;
  cliente: string;
  items: { id: number; nombre: string; cantidad: number; precio: number }[];
}

export const crearPedido = (solicitud: SolicitudCheckout) =>
  apiFetch<PedidoCreado>('/api/checkout', {
    method: 'POST',
    body: JSON.stringify(solicitud),
  });
