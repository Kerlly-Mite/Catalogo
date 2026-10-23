// src/pwa/InstallButton.tsx
import { useInstallPrompt } from './useInstallPrompt';

/** Botón "Instalar aplicación": solo aparece cuando el navegador permite instalar la PWA. */
const InstallButton = () => {
  const { puedeInstalar, instalar } = useInstallPrompt();
  if (!puedeInstalar) return null;

  return (
    <button
      type="button"
      onClick={instalar}
      className="min-h-[44px] px-3 rounded-lg bg-indigo-600 text-white text-sm font-semibold hover:bg-indigo-700 transition focus:outline-none focus:ring-2 focus:ring-indigo-500"
    >
      <span className="hidden sm:inline">Instalar aplicación</span>
      <span className="sm:hidden" aria-label="Instalar aplicación">⬇ Instalar</span>
    </button>
  );
};

export default InstallButton;
