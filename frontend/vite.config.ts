import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'
import { VitePWA } from 'vite-plugin-pwa'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    VitePWA({
      // injectManifest: el service worker lo escribimos nosotros (src/sw.ts) y Workbox
      // solo le inyecta la lista de archivos del build para el precaché del app shell.
      strategies: 'injectManifest',
      srcDir: 'src',
      filename: 'sw.ts',
      // El registro lo hace UpdatePrompt (virtual:pwa-register/react) y solo en producción.
      injectRegister: false,
      registerType: 'prompt',
      // El SW NO corre con `npm run dev`; se prueba con `npm run build` + `npm run preview`.
      devOptions: { enabled: false },
      includeAssets: ['favicon.svg', 'icons/apple-touch-icon.png', 'offline.html'],
      injectManifest: {
        globPatterns: ['**/*.{js,css,html,svg,png,webmanifest}'],
      },
      manifest: {
        name: 'MultiCatálogo',
        short_name: 'MultiCatálogo',
        description: 'Tienda de cosméticos con red de ventas multinivel.',
        lang: 'es',
        start_url: '/',
        scope: '/',
        display: 'standalone',
        orientation: 'portrait',
        theme_color: '#2563eb',
        background_color: '#f8fafc',
        icons: [
          { src: '/icons/icon-192.png', sizes: '192x192', type: 'image/png', purpose: 'any' },
          { src: '/icons/icon-512.png', sizes: '512x512', type: 'image/png', purpose: 'any' },
          { src: '/icons/icon-maskable-512.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
        ],
      },
    }),
  ],
  // El backend solo permite CORS desde http://localhost:5173, así que `npm run preview`
  // (el build de producción) usa ese mismo puerto en lugar del 4173 por defecto.
  preview: { port: 5173, strictPort: true },
})
