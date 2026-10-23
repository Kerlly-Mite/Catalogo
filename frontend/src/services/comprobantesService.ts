// src/services/comprobantesService.ts
// Voucher de una transferencia bancaria: se sube ANTES de confirmar el pedido y su id viaja en pagoDetalle.
// Las fotos se reducen en el navegador; un PDF se sube tal cual (máximo 2 MB).
import { apiFetch, API_URL } from './http';
import { prepararImagen } from '../utils/imagen';

export const MAX_COMPROBANTE_BYTES = 2 * 1024 * 1024;

/** Valida y sube el voucher. Devuelve el id que guarda el servidor. */
export const subirComprobante = async (archivo: File): Promise<number> => {
  let contenido: Blob = archivo;
  let nombre = archivo.name;

  if (archivo.type === 'application/pdf') {
    if (archivo.size > MAX_COMPROBANTE_BYTES) throw new Error('El PDF no puede superar 2 MB.');
  } else if (archivo.type.startsWith('image/')) {
    contenido = await prepararImagen(archivo, 1400); // JPEG reducido: pesa mucho menos de 2 MB
    nombre = 'comprobante.jpg';
  } else {
    throw new Error('El comprobante debe ser una imagen (JPG, PNG, WebP) o un PDF.');
  }

  const datos = new FormData();
  datos.append('comprobante', contenido, nombre);
  const { id } = await apiFetch<{ id: number }>('/api/comprobantes', { method: 'POST', body: datos });
  return id;
};

/** Abre el voucher en otra pestaña (solo admin). Pide el archivo con el token y lo muestra como blob. */
export const abrirComprobante = async (id: number): Promise<void> => {
  // La pestaña se abre YA (dentro del clic) para que el celular no la bloquee; luego se le pone el archivo.
  const ventana = window.open('', '_blank');
  try {
    const res = await fetch(`${API_URL}/api/comprobantes/${id}`, {
      headers: { Authorization: `Bearer ${localStorage.getItem('token') ?? ''}` },
    });
    if (!res.ok) throw new Error('No se pudo abrir el comprobante.');
    const url = URL.createObjectURL(await res.blob());
    if (ventana) ventana.location.href = url;
    else window.location.assign(url);
  } catch (e) {
    ventana?.close();
    throw e;
  }
};
