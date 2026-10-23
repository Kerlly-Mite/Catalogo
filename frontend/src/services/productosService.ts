// src/services/productosService.ts
// Capa de servicios del catálogo.
// Actualmente devuelve datos de ejemplo (mock) simulando una llamada asíncrona.
// En la Unidad 2, estas funciones se reemplazarán por fetch a GET /api/productos
// sin necesidad de modificar los componentes que las consumen.

import type { Producto } from "../data/productos";

const API_URL = import.meta.env.VITE_API_URL ?? "http://localhost:3000";

export const getProductos = async (): Promise<Producto[]> => {
  const res = await fetch(`${API_URL}/api/productos`);
  if (!res.ok) throw new Error("No se pudieron cargar los productos");
  return res.json();
};

export const getProductoById = async (id: number): Promise<Producto | undefined> => {
  const res = await fetch(`${API_URL}/api/productos/${id}`);
  if (res.status === 404) return undefined;
  if (!res.ok) throw new Error("No se pudo cargar el producto");
  return res.json();
};