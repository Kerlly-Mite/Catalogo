/// <reference lib="webworker" />
// src/sw.ts — Service worker de MultiCatálogo (Workbox, estrategia injectManifest).
//
// REGLAS DE SEGURIDAD (el backend usa JWT y rutas privadas):
//  1. Solo se cachean peticiones GET públicas: el catálogo (/api/productos[/:id]) e imágenes.
//  2. NUNCA se cachea nada más de /api/*: login, register, checkout, pedidos, red, comisiones,
//     liquidaciones, reportes, usuarios y admin pasan siempre directo a la red (sin handler = sin caché).
//  3. Las escrituras (POST/PUT/PATCH/DELETE) no se interceptan ni se encolan: sin conexión fallan con aviso.
import { cleanupOutdatedCaches, matchPrecache, precacheAndRoute, createHandlerBoundToURL } from 'workbox-precaching'
import { NavigationRoute, registerRoute, setCatchHandler } from 'workbox-routing'
import { CacheFirst, StaleWhileRevalidate } from 'workbox-strategies'
import { ExpirationPlugin } from 'workbox-expiration'
import { CacheableResponsePlugin } from 'workbox-cacheable-response'

declare const self: ServiceWorkerGlobalScope

// 1) App shell: HTML, JS, CSS e íconos generados en el build.
precacheAndRoute(self.__WB_MANIFEST)
cleanupOutdatedCaches()

// 2) Navegaciones de la SPA → index.html precacheado (así la app abre sin conexión).
//    Se excluye /api para no devolver HTML a una llamada de API.
registerRoute(new NavigationRoute(createHandlerBoundToURL('/index.html'), { denylist: [/^\/api\//] }))

// 3) Catálogo público: stale-while-revalidate (se ve al instante y se actualiza en segundo plano).
//    Se compara solo el pathname porque la API vive en otro origen (localhost:3000).
const esCatalogoPublico = ({ url, request }: { url: URL; request: Request }) =>
  request.method === 'GET' && /^\/api\/productos(\/\d+)?\/?$/.test(url.pathname) && !url.search

registerRoute(
  esCatalogoPublico,
  new StaleWhileRevalidate({
    cacheName: 'api-productos',
    plugins: [
      new CacheableResponsePlugin({ statuses: [200] }),
      new ExpirationPlugin({ maxEntries: 60, maxAgeSeconds: 60 * 60 * 24 * 7 }),
    ],
  }),
)

// 4) Imágenes (productos, avatar): caché primero con límite de tamaño y vigencia.
//    status 0 = respuesta opaca de imágenes de otros dominios (picsum.photos).
registerRoute(
  ({ request }) => request.destination === 'image',
  new CacheFirst({
    cacheName: 'imagenes',
    plugins: [
      new CacheableResponsePlugin({ statuses: [0, 200] }),
      new ExpirationPlugin({ maxEntries: 80, maxAgeSeconds: 60 * 60 * 24 * 30, purgeOnQuotaError: true }),
    ],
  }),
)

// 5) Respaldo: si una navegación falla del todo, mostramos offline.html (precacheada).
setCatchHandler(async ({ request }) => {
  if (request.destination === 'document') {
    const offline = await matchPrecache('/offline.html')
    if (offline) return offline
  }
  return Response.error()
})

// 6) Actualización controlada: el SW nuevo espera hasta que el usuario pulsa "Actualizar".
self.addEventListener('message', (event) => {
  if (event.data?.type === 'SKIP_WAITING') void self.skipWaiting()
})
self.addEventListener('activate', () => {
  void self.clients.claim()
})
