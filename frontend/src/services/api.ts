// src/services/api.ts
// Capa de servicios: centraliza TODA la comunicación con la API RESTful (backend en Go).
// Así los componentes no saben de URLs ni de fetch, solo llaman funciones.

// Producto tal como lo devuelve hoy el backend (GET /api/productos): solo campos básicos.
// (El tipo enriquecido con descripción, categoría y galería vive en src/data/productos.ts)
export interface ProductoApi {
  id: number;
  nombre: string;
  precio: number;
  img: string;
}

// Leemos la URL base desde las variables de entorno de Vite (archivo .env).
// Si no existe, usamos localhost:3000 como valor por defecto.
const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

// Tipo de la respuesta que devuelve el endpoint POST /api/login
export interface LoginResponse {
  token: string;
  email: string;
  rol: string; // 'admin' | 'cliente' (Tema 5): lo normaliza el componente Login
}

/**
 * Envía las credenciales al backend (POST /api/login).
 * Devuelve el token, el email y el rol si son correctas; lanza un Error si no.
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
export const getProductos = async (): Promise<ProductoApi[]> => {
  const response = await fetch(`${API_URL}/api/productos`);

  if (!response.ok) {
    throw new Error('No se pudieron cargar los productos');
  }

  return response.json();
};

// Respuesta de POST /api/register: igual que el login + nombre y el código de referido propio.
export interface RegisterResponse extends LoginResponse {
  nombre: string;
  codigo_referido: string;
}

/**
 * Crea una cuenta nueva (POST /api/register). El código de referido es opcional.
 * Si todo sale bien, la API ya devuelve el token (inicio de sesión automático).
 */
export const registerRequest = async (
  nombre: string,
  email: string,
  password: string,
  ref?: string
): Promise<RegisterResponse> => {
  const response = await fetch(`${API_URL}/api/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ nombre, email, password, ref: ref?.trim() || undefined }),
  });

  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data.message ?? data.error ?? 'No se pudo crear la cuenta');
  }

  return response.json();
};
