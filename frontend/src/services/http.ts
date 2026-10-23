// src/services/http.ts
// Cliente HTTP único: agrega solo el header Authorization (JWT) y convierte los errores de la API en ApiError.

import { invalidarCatalogoEnCache, limpiarCachesDeDatos } from '../pwa/cache';

export const API_URL: string = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

/** Error de la API: conserva el status HTTP y el `details` (solo viene con APP_ENV=dev en el backend). */
export class ApiError extends Error {
  status: number;
  details?: unknown;

  constructor(status: number, message: string, details?: unknown) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.details = details;
  }
}

/** Texto amigable para mostrar cualquier error al usuario. */
export const mensajeDeError = (err: unknown): string =>
  err instanceof Error ? err.message : 'Error de conexión con el servidor';

export const apiFetch = async <T>(path: string, options: RequestInit = {}): Promise<T> => {
  const token = localStorage.getItem('token');
  const headers = new Headers(options.headers);
  if (options.body && !(options.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json'); // con FormData el navegador pone el suyo (con el boundary)
  }
  if (token) headers.set('Authorization', `Bearer ${token}`);

  const metodo = (options.method ?? 'GET').toUpperCase();
  const esEscritura = metodo !== 'GET';

  // PWA: las escrituras (checkout, altas, ediciones) NO funcionan sin conexión ni se encolan: se avisa.
  if (esEscritura && !navigator.onLine) {
    throw new ApiError(0, 'Sin conexión: esta acción requiere internet. Inténtalo de nuevo cuando vuelva la conexión.');
  }

  const res = await fetch(`${API_URL}${path}`, { ...options, headers });

  // Una escritura exitosa puede cambiar catálogo/stock: descartamos el catálogo guardado por el service worker.
  if (esEscritura && res.ok) void invalidarCatalogoEnCache();

  if (!res.ok) {
    // Sesión vencida o token inválido: cerramos la sesión local y volvemos al login.
    if (res.status === 401 && token) {
      localStorage.removeItem('token');
      localStorage.removeItem('usuario');
      void limpiarCachesDeDatos();
      window.location.assign('/login');
    }
    const data = await res.json().catch(() => ({}));
    throw new ApiError(res.status, data.message ?? data.error ?? 'Error en la petición', data.details);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
};
