// src/services/adminProductosService.ts
// Administración de productos (solo admin). El borrado es lógico: el producto queda "inactivo".
import { API_URL, apiFetch } from './http';
import type { Producto } from '../data/productos';
import type { Pagina } from './pedidosService';

export interface ProductoAdmin extends Producto {
  stock: number;
  activo: boolean;
}

export interface ProductoForm {
  nombre: string;
  descripcion: string;
  precio: number;
  categoria: string;
  img: string;
  galeria: string[];
  stock: number;
}

export const getProductosAdmin = (q: string, page: number, limit: number) =>
  apiFetch<Pagina<ProductoAdmin>>(
    `/api/admin/productos?page=${page}&limit=${limit}&q=${encodeURIComponent(q)}`
  );

export const crearProducto = (datos: ProductoForm) =>
  apiFetch<ProductoAdmin>('/api/productos', { method: 'POST', body: JSON.stringify(datos) });

export const actualizarProducto = (id: number, cambios: Partial<ProductoForm> & { activo?: boolean }) =>
  apiFetch<ProductoAdmin>(`/api/productos/${id}`, { method: 'PUT', body: JSON.stringify(cambios) });

/** Desactiva el producto (no borra la fila). */
export const desactivarProducto = (id: number) =>
  apiFetch<void>(`/api/productos/${id}`, { method: 'DELETE' });

/** Sube una foto (se guarda en la base de datos) y devuelve su URL pública COMPLETA. */
export const subirImagen = async (foto: Blob): Promise<string> => {
  const formulario = new FormData();
  formulario.append('imagen', foto, 'foto.jpg');
  const respuesta = await apiFetch<{ id: number; url: string }>('/api/imagenes', { method: 'POST', body: formulario });
  return `${API_URL}${respuesta.url}`;
};
