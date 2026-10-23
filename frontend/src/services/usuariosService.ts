// src/services/usuariosService.ts
// Administración de cuentas (solo admin).
import { apiFetch } from './http';
import type { Pagina } from './pedidosService';
import type { Rol } from '../context/AuthContext';

export interface UsuarioListado {
  id: number;
  email: string;
  nombre: string;
  rol: Rol;
  activo: boolean;
  creado_en: string;
  codigo_referido: string;
}

export const getUsuarios = (page: number, limit: number, q: string) => {
  const p = new URLSearchParams({ page: String(page), limit: String(limit) });
  if (q) p.set('q', q);
  return apiFetch<Pagina<UsuarioListado>>(`/api/usuarios?${p.toString()}`);
};

/** Cambia el rol y/o activa-desactiva la cuenta. Devuelve el usuario ya actualizado. */
export const actualizarUsuario = (id: number, cambios: { rol?: Rol; activo?: boolean }) =>
  apiFetch<UsuarioListado>(`/api/usuarios/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(cambios),
  });
