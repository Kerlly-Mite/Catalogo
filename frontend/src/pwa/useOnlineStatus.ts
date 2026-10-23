// src/pwa/useOnlineStatus.ts
import { useSyncExternalStore } from 'react';

const suscribir = (callback: () => void) => {
  window.addEventListener('online', callback);
  window.addEventListener('offline', callback);
  return () => {
    window.removeEventListener('online', callback);
    window.removeEventListener('offline', callback);
  };
};

/** `true` si el navegador reporta conexión. Se actualiza solo al perderla o recuperarla. */
export const useOnlineStatus = (): boolean =>
  useSyncExternalStore(suscribir, () => navigator.onLine, () => true);
