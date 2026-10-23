// src/pwa/useInstallPrompt.ts
// Captura el evento `beforeinstallprompt` (Chrome/Edge/Android) para ofrecer un botón "Instalar aplicación".
import { useCallback, useEffect, useState } from 'react';

interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
}

export const useInstallPrompt = () => {
  const [evento, setEvento] = useState<BeforeInstallPromptEvent | null>(null);

  useEffect(() => {
    const alOfrecer = (e: Event) => {
      e.preventDefault(); // evitamos el mini-aviso del navegador: lo mostramos nosotros
      setEvento(e as BeforeInstallPromptEvent);
    };
    const alInstalar = () => setEvento(null);
    window.addEventListener('beforeinstallprompt', alOfrecer);
    window.addEventListener('appinstalled', alInstalar);
    return () => {
      window.removeEventListener('beforeinstallprompt', alOfrecer);
      window.removeEventListener('appinstalled', alInstalar);
    };
  }, []);

  const instalar = useCallback(async () => {
    if (!evento) return;
    await evento.prompt();
    await evento.userChoice;
    setEvento(null); // el evento solo se puede usar una vez
  }, [evento]);

  return { puedeInstalar: evento !== null, instalar };
};
