// src/pwa/cache.ts
// Cachés de datos creadas por el service worker. Solo guardan información pública (catálogo, imágenes),
// pero se vacían al cerrar sesión y tras escrituras que cambian el catálogo/stock para no mostrar datos viejos.
const CACHES_DE_DATOS = ['api-productos', 'imagenes'];

export const limpiarCachesDeDatos = async (): Promise<void> => {
  if (typeof caches === 'undefined') return;
  try {
    await Promise.all(CACHES_DE_DATOS.map((nombre) => caches.delete(nombre)));
  } catch {
    // Si el navegador bloquea Cache Storage no es un error para la app.
  }
};

/** Solo el catálogo (no las imágenes): se invalida cuando cambia el stock o los productos. */
export const invalidarCatalogoEnCache = async (): Promise<void> => {
  if (typeof caches === 'undefined') return;
  try {
    await caches.delete('api-productos');
  } catch {
    // ignorado
  }
};
