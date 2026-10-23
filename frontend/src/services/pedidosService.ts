// src/services/pedidosService.ts
import { apiFetch } from './http';
import type { EstadoPedido } from '../data/pedidos';

export interface Pagina<T> {
  data: T[];
  page: number;
  limit: number;
  total: number;
}

export interface PedidoListado {
  id: number;
  numero: string;
  fecha: string;
  total: number;
  estado: EstadoPedido;
  cliente_nombre: string;
  cliente_email: string;
  num_items: number;
}

export interface PedidoItem {
  id: number;
  nombre: string;
  cantidad: number;
  precio: number;
}

export interface PedidoDetalle extends PedidoListado {
  direccion: string;
  ciudad: string;
  telefono: string;
  metodo_pago: string;
  pago_detalle?: Record<string, string>;
  items: PedidoItem[];
}

/** Pedidos del usuario con sesión iniciada (GET /api/pedidos/mios). */
export const getMisPedidos = (page: number, limit: number) =>
  apiFetch<Pagina<PedidoListado>>(`/api/pedidos/mios?page=${page}&limit=${limit}`);

/** Todos los pedidos, con filtro opcional por estado (solo admin). */
export const getPedidosAdmin = (page: number, limit: number, estado: EstadoPedido | '') =>
  apiFetch<Pagina<PedidoListado>>(
    `/api/pedidos?page=${page}&limit=${limit}${estado ? `&estado=${estado}` : ''}`
  );

export const getPedido = (id: number) => apiFetch<PedidoDetalle>(`/api/pedidos/${id}`);

/** Cambia el estado (solo admin). Cancelar devuelve el stock y revierte comisiones. */
export const cambiarEstadoPedido = (id: number, estado: EstadoPedido) =>
  apiFetch<PedidoDetalle>(`/api/pedidos/${id}/estado`, {
    method: 'PATCH',
    body: JSON.stringify({ estado }),
  });
