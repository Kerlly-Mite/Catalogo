// src/services/redService.ts
// Red propia y estado de cuenta de comisiones del usuario con sesión.
import { apiFetch } from './http';
import type { Pagina } from './pedidosService';
import type { Referido } from '../data/red';

/** GET /api/red/mia: el subárbol desde MI nodo (nivel 1 = mis directos) y mi código de invitación. */
export interface RedMia {
  codigo_referido: string;
  red: Referido;
}

export type EstadoComision = 'pendiente' | 'aprobada' | 'pagada' | 'revertida';

export interface ComisionLedger {
  id: number;
  pedido_id: number;
  numero: string;
  nivel: number;
  porcentaje: number;
  monto: number;
  estado: EstadoComision;
  fecha: string;
}

export interface ComisionesMias extends Pagina<ComisionLedger> {
  totales: Record<EstadoComision, number>;
}

export const getMiRed = () => apiFetch<RedMia>('/api/red/mia');

export const getMisComisiones = (page: number, limit: number) =>
  apiFetch<ComisionesMias>(`/api/comisiones/mias?page=${page}&limit=${limit}`);
