// src/pwa/PwaStatus.tsx
// Registra el service worker (solo en producción) y muestra los avisos "Sin conexión" y "Nueva versión".
import { useRegisterSW } from 'virtual:pwa-register/react';
import { useOnlineStatus } from './useOnlineStatus';

const PwaStatus = () => {
  const online = useOnlineStatus();
  const {
    needRefresh: [hayActualizacion, setHayActualizacion],
    updateServiceWorker,
  } = useRegisterSW();

  if (online && !hayActualizacion) return null;

  return (
    <div
      className="fixed inset-x-0 bottom-0 z-[100] flex flex-col items-center gap-2 p-3 pointer-events-none"
      style={{ paddingBottom: 'max(0.75rem, env(safe-area-inset-bottom))' }}
    >
      {!online && (
        <div
          role="status"
          className="pointer-events-auto w-full max-w-md rounded-lg bg-slate-800 text-white text-sm px-4 py-3 shadow-lg"
        >
          📡 Sin conexión. Puedes ver el catálogo; las compras y los pagos se habilitan al volver internet.
        </div>
      )}
      {hayActualizacion && (
        <div
          role="alert"
          className="pointer-events-auto w-full max-w-md rounded-lg bg-white border border-indigo-200 text-slate-800 text-sm px-4 py-3 shadow-lg flex items-center justify-between gap-3"
        >
          <span>Hay una nueva versión disponible.</span>
          <span className="flex gap-2 shrink-0">
            <button
              type="button"
              onClick={() => setHayActualizacion(false)}
              className="min-h-[44px] px-3 rounded-lg text-slate-600 hover:bg-slate-100"
            >
              Luego
            </button>
            <button
              type="button"
              onClick={() => updateServiceWorker(true)}
              className="min-h-[44px] px-3 rounded-lg bg-indigo-600 text-white font-semibold hover:bg-indigo-700"
            >
              Actualizar
            </button>
          </span>
        </div>
      )}
    </div>
  );
};

export default PwaStatus;
