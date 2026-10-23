// src/services/reportesService.ts
// Reportes del negocio (solo admin): todo sale de SQL de agregación en el backend.
import { apiFetch } from './http';
import type { DeudaReferido } from './liquidacionesService';

export interface ReporteResumen {
  desde?: string;
  hasta?: string;
  total_vendido: number;
  pedidos: number;
  pedidos_cancelados: number;
  ticket_promedio: number;
  comisiones_devengadas: number;
  comisiones_pendientes: number;
  comisiones_pagadas: number;
  distribuidores: number;
}

export interface TopProducto {
  id: number;
  nombre: string;
  unidades: number;
  ingresos: number;
}

export interface TopReferido {
  referido_id: number;
  nombre: string;
  pedidos: number;
  ventas_red: number;
  comisiones: number;
}

export interface ComisionesPendientesReporte {
  data: DeudaReferido[];
  total_pendiente: number;
}

/** Arma "?desde=...&hasta=...&limit=..." solo con lo que tenga valor (fechas AAAA-MM-DD, ambas inclusivas). */
const consulta = (desde: string, hasta: string, limit?: number): string => {
  const p = new URLSearchParams();
  if (desde) p.set('desde', desde);
  if (hasta) p.set('hasta', hasta);
  if (limit) p.set('limit', String(limit));
  const texto = p.toString();
  return texto ? `?${texto}` : '';
};

export const getResumen = (desde: string, hasta: string) =>
  apiFetch<ReporteResumen>(`/api/reportes/resumen${consulta(desde, hasta)}`);

export const getTopProductos = (desde: string, hasta: string, limit = 5) =>
  apiFetch<TopProducto[]>(`/api/reportes/top-productos${consulta(desde, hasta, limit)}`);

export const getTopReferidos = (desde: string, hasta: string, limit = 5) =>
  apiFetch<TopReferido[]>(`/api/reportes/top-referidos${consulta(desde, hasta, limit)}`);

export const getComisionesPendientes = () =>
  apiFetch<ComisionesPendientesReporte>('/api/reportes/comisiones-pendientes');
