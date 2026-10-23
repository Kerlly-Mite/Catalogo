// src/services/liquidacionesService.ts
import { apiFetch } from './http';
import type { Pagina } from './pedidosService';

export interface Liquidacion {
  id: number;
  referido_id: number;
  referido: string;
  total: number;
  estado: string;
  creado_en: string;
  pagada_en: string;
}

export interface DeudaReferido {
  referido_id: number;
  nombre: string;
  pendiente: number;
  comisiones: number;
}

/** Cuánto se le debe a cada distribuidor (solo admin). */
export const getDeuda = () => apiFetch<DeudaReferido[]>('/api/liquidaciones/deuda');

/** Paga todas las comisiones pendientes del distribuidor (solo admin). */
export const liquidar = (referidoId: number) =>
  apiFetch<Liquidacion>('/api/liquidaciones', {
    method: 'POST',
    body: JSON.stringify({ referido_id: referidoId }),
  });

export const getLiquidaciones = (page: number, limit: number) =>
  apiFetch<Pagina<Liquidacion>>(`/api/liquidaciones?page=${page}&limit=${limit}`);

/** Mis cobros (distribuidor con sesión). */
export const getMisLiquidaciones = (page: number, limit: number) =>
  apiFetch<Pagina<Liquidacion>>(`/api/liquidaciones/mias?page=${page}&limit=${limit}`);
