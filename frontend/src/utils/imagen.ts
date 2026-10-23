// src/utils/imagen.ts
// Prepara una foto antes de subirla: la reduce (máx. 1000 px de lado) y la convierte a JPEG.
// Así una foto de celular de 5 MB sube en segundos y pesa unos 100-300 KB.

export const prepararImagen = async (archivo: File, maxLado = 1000, calidad = 0.85): Promise<Blob> => {
  let bitmap: ImageBitmap;
  try {
    bitmap = await createImageBitmap(archivo);
  } catch {
    throw new Error('No se pudo leer la imagen. Usa una foto JPG, PNG o WebP.');
  }

  const escala = Math.min(1, maxLado / Math.max(bitmap.width, bitmap.height));
  const ancho = Math.round(bitmap.width * escala);
  const alto = Math.round(bitmap.height * escala);

  const canvas = document.createElement('canvas');
  canvas.width = ancho;
  canvas.height = alto;
  const ctx = canvas.getContext('2d');
  if (!ctx) {
    bitmap.close();
    throw new Error('Tu navegador no pudo procesar la imagen.');
  }
  ctx.fillStyle = '#ffffff'; // fondo blanco: los PNG con transparencia no quedan negros en JPEG
  ctx.fillRect(0, 0, ancho, alto);
  ctx.drawImage(bitmap, 0, 0, ancho, alto);
  bitmap.close();

  const blob = await new Promise<Blob | null>((resolver) => canvas.toBlob(resolver, 'image/jpeg', calidad));
  if (!blob) throw new Error('No se pudo procesar la imagen.');
  return blob;
};
