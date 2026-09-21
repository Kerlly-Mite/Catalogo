// src/services/api.ts
// Capa de servicios: centraliza TODA la comunicación con la API RESTful (backend en Go).
// Así los componentes no saben de URLs ni de fetch, solo llaman funciones.

import type { Producto } from '../context/CartContext';

// Leemos la URL base desde las variables de entorno de Vite (archivo .env).
// Si no existe, usamos localhost:3000 como valor por defecto.
const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

// Tipo de la respuesta que devuelve el endpoint POST /api/login
export interface LoginResponse {
  token: string;
  email: string;
}

/**
 * Envía las credenciales al backend (POST /api/login).
 * Devuelve el token y el email si son correctas; lanza un Error si no.
 */
export const loginRequest = async (
  email: string,
  password: string
): Promise<LoginResponse> => {
  const response = await fetch(`${API_URL}/api/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    // El backend espera un JSON con los campos "email" y "password".
    body: JSON.stringify({ email, password }),
  });

  // Si el backend responde 401 o 400, leemos el mensaje de error que envió.
  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data.error ?? 'No se pudo iniciar sesión');
  }

  return response.json();
};

/**
 * Obtiene el catálogo de productos desde el backend (GET /api/productos).
 */
export const getProductos = async (): Promise<Producto[]> => {
  const response = await fetch(`${API_URL}/api/productos`);

  if (!response.ok) {
    throw new Error('No se pudieron cargar los productos');
  }

  return response.json();
};
