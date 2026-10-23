// src/components/Checkout.tsx
// Checkout REAL: envía el pedido a POST /api/checkout (transacción ACID en el backend):
// - Formulario de envío con validación básica
// - Resumen del pedido en tiempo real
// - Si el servidor responde bien: vacía el carrito y muestra la confirmación con el pedido REAL
// - Si falla (stock, validación, caída de la BD): muestra el mensaje y CONSERVA el carrito
// - Cada método de pago pide sus propios datos (tarjeta / transferencia con voucher / efectivo contra entrega).
//   De la tarjeta SOLO viajan titular, últimos 4 dígitos y vencimiento: el número completo y el CVV no salen del navegador.
import { useEffect, useRef, useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useCart } from '../context/CartContext';
import { useAuth } from '../context/AuthContext';
import { crearPedido } from '../services/checkoutService';
import { subirComprobante, MAX_COMPROBANTE_BYTES } from '../services/comprobantesService';
import { mensajeDeError } from '../services/http';
import { useOnlineStatus } from '../pwa/useOnlineStatus';

interface DatosEnvio {
  nombre: string;
  direccion: string;
  ciudad: string;
  telefono: string;
  metodoPago: string;
}

const METODOS_PAGO = ["Tarjeta de crédito", "Transferencia bancaria", "Efectivo contra entrega"];

// ⚠️ Cambia estos datos por los de la cuenta REAL donde quieres recibir las transferencias.
const DATOS_BANCARIOS = {
  banco: "Banco Pichincha",
  tipo: "Cuenta de ahorros",
  cuenta: "0000000000",
  titular: "MultiCatálogo",
  documento: "0000000000001", // RUC / C.I.
};

const BANCOS = [
  "Banco Pichincha",
  "Banco del Pacífico",
  "Banco de Guayaquil",
  "Produbanco",
  "Banco Bolivariano",
  "Banco Internacional",
  "Banco del Austral",
  "Cooperativa JEP",
  "Otro banco",
];

// Deben ser EXACTAMENTE las mismas que acepta el backend (checkoutController.go).
const FRANJAS = ["Mañana (8:00 - 12:00)", "Tarde (12:00 - 18:00)", "Noche (18:00 - 21:00)"];

const soloDigitos = (v: string) => v.replace(/\D/g, "");
const formatearNumeroTarjeta = (v: string) => soloDigitos(v).slice(0, 19).replace(/(.{4})/g, "$1 ").trim();
const formatearVencimiento = (v: string) => {
  const d = soloDigitos(v).slice(0, 4);
  return d.length > 2 ? `${d.slice(0, 2)}/${d.slice(2)}` : d;
};

const CAMPO =
  "w-full min-h-[44px] px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none transition";

const Checkout = () => {
  const { cart, totalPrice, clearCart } = useCart();
  const { user } = useAuth();
  const navigate = useNavigate();

  const [datos, setDatos] = useState<DatosEnvio>({
    nombre: "",
    direccion: "",
    ciudad: "",
    telefono: "",
    metodoPago: METODOS_PAGO[0],
  });
  const [tarjeta, setTarjeta] = useState({ numero: "", titular: "", vencimiento: "", cvv: "" });
  const [transferencia, setTransferencia] = useState({ banco: "", referencia: "" });
  const [voucher, setVoucher] = useState<File | null>(null);
  const [efectivo, setEfectivo] = useState({ recibe: "", franja: "", pagaCon: "", indicaciones: "" });
  const voucherSubido = useRef<{ archivo: File; id: number } | null>(null); // evita subirlo dos veces si se reintenta
  const errorRef = useRef<HTMLDivElement>(null);

  const online = useOnlineStatus(); // PWA: sin conexión no se puede pagar
  const [procesando, setProcesando] = useState<boolean>(false);
  const [error, setError] = useState<string>("");

  // El mensaje de error aparece junto al botón: lo traemos a la vista (en el celular el formulario es largo).
  useEffect(() => {
    if (error) errorRef.current?.scrollIntoView({ behavior: "smooth", block: "center" });
  }, [error]);

  // Si el carrito está vacío, mostramos un aviso en lugar del formulario
  if (cart.length === 0) {
    return (
      <div className="bg-white p-10 rounded-lg border border-slate-200 text-center">
        <p className="text-slate-600 text-lg mb-4">No tienes productos en el carrito.</p>
        <Link to="/catalogo" className="text-indigo-600 font-semibold hover:underline">
          ← Ir al catálogo
        </Link>
      </div>
    );
  }

  // El pedido se registra a nombre de la cuenta con sesión (el backend usa el email del token).
  const emailCuenta = user?.email ?? "";

  const handleChange = (campo: keyof DatosEnvio, valor: string) => {
    setDatos((prev) => ({ ...prev, [campo]: valor }));
  };

  // Devuelve el mensaje del primer problema, o "" si los datos del método elegido están bien.
  const validarPago = (): string => {
    if (datos.metodoPago === METODOS_PAGO[0]) {
      const numero = soloDigitos(tarjeta.numero);
      if (numero.length < 13 || numero.length > 19) return "El número de tarjeta debe tener entre 13 y 19 dígitos.";
      if (tarjeta.titular.trim().length < 3) return "Escribe el nombre del titular como aparece en la tarjeta.";
      const m = /^(0[1-9]|1[0-2])\/(\d{2})$/.exec(tarjeta.vencimiento);
      if (!m) return "El vencimiento debe tener el formato MM/AA (por ejemplo 08/28).";
      const hoy = new Date();
      const anio = 2000 + Number(m[2]);
      const mes = Number(m[1]);
      if (anio < hoy.getFullYear() || (anio === hoy.getFullYear() && mes < hoy.getMonth() + 1)) return "La tarjeta está vencida.";
      if (tarjeta.cvv.length < 3) return "El CVV debe tener 3 o 4 dígitos.";
    } else if (datos.metodoPago === METODOS_PAGO[1]) {
      if (!transferencia.banco) return "Selecciona el banco desde el que transferiste.";
      if (transferencia.referencia.trim().length < 4) return "Escribe el número de referencia o comprobante de la transferencia.";
      if (!voucher) return "Sube el voucher (foto o PDF) de la transferencia.";
      if (voucher.type !== "application/pdf" && !voucher.type.startsWith("image/")) return "El voucher debe ser una imagen o un PDF.";
      if (voucher.type === "application/pdf" && voucher.size > MAX_COMPROBANTE_BYTES) return "El PDF no puede superar 2 MB.";
    } else {
      if (efectivo.recibe.trim().length < 3) return "Escribe el nombre de quien recibe el pedido.";
      if (!efectivo.franja) return "Selecciona la franja horaria de entrega.";
      const paga = Number(efectivo.pagaCon);
      if (!efectivo.pagaCon || Number.isNaN(paga)) return "Indica con cuánto vas a pagar.";
      if (paga < totalPrice) return `Debes pagar con al menos $${totalPrice.toFixed(2)} (el total del pedido).`;
    }
    return "";
  };

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (procesando) return; // evita enviar el pedido dos veces con doble clic
    setError("");

    const problema = validarPago();
    if (problema) {
      setError(problema);
      return;
    }
    setProcesando(true);

    try {
      // Datos del método elegido. De la tarjeta NO se envía el número completo ni el CVV.
      let pagoDetalle: Record<string, string> = {};
      if (datos.metodoPago === METODOS_PAGO[0]) {
        pagoDetalle = {
          titular: tarjeta.titular.trim(),
          ultimos4: soloDigitos(tarjeta.numero).slice(-4),
          vencimiento: tarjeta.vencimiento,
        };
      } else if (datos.metodoPago === METODOS_PAGO[1] && voucher) {
        if (voucherSubido.current?.archivo !== voucher) {
          voucherSubido.current = { archivo: voucher, id: await subirComprobante(voucher) };
        }
        pagoDetalle = {
          banco_origen: transferencia.banco,
          referencia: transferencia.referencia.trim(),
          comprobante_id: String(voucherSubido.current.id),
        };
      } else {
        pagoDetalle = {
          recibe: efectivo.recibe.trim(),
          franja: efectivo.franja,
          paga_con: efectivo.pagaCon,
          indicaciones: efectivo.indicaciones.trim(),
        };
      }

      const pedidoCreado = await crearPedido({
        cliente: {
          nombre: datos.nombre.trim(),
          email: emailCuenta,
          direccion: datos.direccion.trim(),
          ciudad: datos.ciudad.trim(),
          telefono: datos.telefono.trim(),
        },
        metodoPago: datos.metodoPago,
        pagoDetalle,
        // Solo viajan id y cantidad: el precio y el total los calcula el SERVIDOR.
        items: cart.map((item) => ({ id: item.id, cantidad: item.cantidad })),
      });

      // Pedido guardado: vaciamos el carrito y pasamos el pedido REAL a la confirmación.
      const pedido = {
        numero: pedidoCreado.numero,
        items: pedidoCreado.items.map((it) => ({ nombre: it.nombre, cantidad: it.cantidad, precio: it.precio })),
        total: pedidoCreado.total,
        cliente: { nombre: pedidoCreado.cliente, email: emailCuenta, ciudad: datos.ciudad.trim() },
        fecha: new Date(pedidoCreado.fecha).toLocaleString("es-EC"),
      };
      clearCart();
      navigate("/confirmacion", { state: { pedido } });
    } catch (err) {
      // El carrito NO se vacía: la persona puede corregir o reintentar.
      setError(mensajeDeError(err));
      setProcesando(false);
    }
  };

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-800 mb-6">Finalizar Compra</h1>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* ===== FORMULARIO DE ENVÍO ===== */}
        <form onSubmit={handleSubmit} className="lg:col-span-2 space-y-5">
          <div className="bg-white p-6 rounded-lg border border-slate-200 shadow-sm">
            <h2 className="text-lg font-bold text-slate-800 mb-4">Datos de Envío</h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Nombre completo *</label>
                <input
                  type="text"
                  required
                  minLength={2}
                  value={datos.nombre}
                  onChange={(e) => handleChange("nombre", e.target.value)}
                  className={CAMPO}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Cuenta (correo)</label>
                <input
                  type="email"
                  readOnly
                  value={emailCuenta}
                  className={`${CAMPO} bg-slate-50 text-slate-500`}
                  aria-describedby="ayuda-cuenta"
                />
                <p id="ayuda-cuenta" className="text-xs text-slate-500 mt-1">
                  El pedido se registra a nombre de tu cuenta.
                </p>
              </div>
              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-slate-700 mb-1">Dirección *</label>
                <input
                  type="text"
                  required
                  value={datos.direccion}
                  onChange={(e) => handleChange("direccion", e.target.value)}
                  className={CAMPO}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Ciudad *</label>
                <input
                  type="text"
                  required
                  value={datos.ciudad}
                  onChange={(e) => handleChange("ciudad", e.target.value)}
                  className={CAMPO}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Teléfono *</label>
                <input
                  type="tel"
                  required
                  value={datos.telefono}
                  onChange={(e) => handleChange("telefono", e.target.value)}
                  className={CAMPO}
                />
              </div>
            </div>
          </div>

          <div className="bg-white p-6 rounded-lg border border-slate-200 shadow-sm">
            <h2 className="text-lg font-bold text-slate-800 mb-4">Método de Pago</h2>
            <div className="space-y-3">
              {METODOS_PAGO.map((metodo) => (
                <label
                  key={metodo}
                  className="flex items-center gap-3 min-h-[44px] p-3 rounded-lg border border-slate-200 cursor-pointer hover:border-indigo-300 transition has-[:checked]:border-indigo-600 has-[:checked]:bg-indigo-50"
                >
                  <input
                    type="radio"
                    name="metodoPago"
                    value={metodo}
                    checked={datos.metodoPago === metodo}
                    onChange={(e) => {
                      handleChange("metodoPago", e.target.value);
                      setError("");
                    }}
                    className="accent-indigo-600"
                  />
                  <span className="text-sm text-slate-700">{metodo}</span>
                </label>
              ))}
            </div>

            <div className="border-t border-slate-200 mt-6 pt-6">
              {/* ===== TARJETA ===== */}
              {datos.metodoPago === METODOS_PAGO[0] && (
                <div>
                  <h3 className="font-semibold text-slate-800 mb-4">Datos de la tarjeta</h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="col-span-2">
                      <label htmlFor="tc-numero" className="block text-sm font-medium text-slate-700 mb-1">
                        Número de tarjeta *
                      </label>
                      <input
                        id="tc-numero"
                        type="text"
                        inputMode="numeric"
                        autoComplete="cc-number"
                        placeholder="1234 5678 9012 3456"
                        value={tarjeta.numero}
                        onChange={(e) => setTarjeta((p) => ({ ...p, numero: formatearNumeroTarjeta(e.target.value) }))}
                        className={CAMPO}
                      />
                    </div>
                    <div className="col-span-2">
                      <label htmlFor="tc-titular" className="block text-sm font-medium text-slate-700 mb-1">
                        Nombre del titular *
                      </label>
                      <input
                        id="tc-titular"
                        type="text"
                        autoComplete="cc-name"
                        placeholder="Como aparece en la tarjeta"
                        value={tarjeta.titular}
                        onChange={(e) => setTarjeta((p) => ({ ...p, titular: e.target.value }))}
                        className={CAMPO}
                      />
                    </div>
                    <div className="col-span-1">
                      <label htmlFor="tc-venc" className="block text-sm font-medium text-slate-700 mb-1">
                        Vencimiento *
                      </label>
                      <input
                        id="tc-venc"
                        type="text"
                        inputMode="numeric"
                        autoComplete="cc-exp"
                        placeholder="MM/AA"
                        value={tarjeta.vencimiento}
                        onChange={(e) => setTarjeta((p) => ({ ...p, vencimiento: formatearVencimiento(e.target.value) }))}
                        className={CAMPO}
                      />
                    </div>
                    <div className="col-span-1">
                      <label htmlFor="tc-cvv" className="block text-sm font-medium text-slate-700 mb-1">
                        CVV *
                      </label>
                      <input
                        id="tc-cvv"
                        type="password"
                        inputMode="numeric"
                        autoComplete="cc-csc"
                        placeholder="•••"
                        maxLength={4}
                        value={tarjeta.cvv}
                        onChange={(e) => setTarjeta((p) => ({ ...p, cvv: soloDigitos(e.target.value).slice(0, 4) }))}
                        className={CAMPO}
                      />
                    </div>
                  </div>
                  <p className="text-xs text-slate-500 mt-4">
                    🔒 Por seguridad solo guardamos los últimos 4 dígitos; el número completo y el CVV no se almacenan.
                  </p>
                </div>
              )}

              {/* ===== TRANSFERENCIA ===== */}
              {datos.metodoPago === METODOS_PAGO[1] && (
                <div className="space-y-4">
                  <div className="rounded-lg bg-slate-50 border border-slate-200 p-4 text-sm text-slate-600 space-y-1">
                    <p className="font-bold text-slate-800">Transfiere ${totalPrice.toFixed(2)} a:</p>
                    <p>
                      Banco: <span className="text-slate-800">{DATOS_BANCARIOS.banco}</span>
                    </p>
                    <p>
                      Tipo: <span className="text-slate-800">{DATOS_BANCARIOS.tipo}</span>
                    </p>
                    <p>
                      N.º de cuenta: <span className="font-mono text-slate-800">{DATOS_BANCARIOS.cuenta}</span>
                    </p>
                    <p>
                      Titular: <span className="text-slate-800">{DATOS_BANCARIOS.titular}</span>
                    </p>
                    <p>
                      RUC / C.I.: <span className="font-mono text-slate-800">{DATOS_BANCARIOS.documento}</span>
                    </p>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label htmlFor="tr-banco" className="block text-sm font-medium text-slate-700 mb-1">
                        Banco desde el que transferiste *
                      </label>
                      <select
                        id="tr-banco"
                        value={transferencia.banco}
                        onChange={(e) => setTransferencia((p) => ({ ...p, banco: e.target.value }))}
                        className={`${CAMPO} bg-white`}
                      >
                        <option value="">Selecciona un banco</option>
                        {BANCOS.map((b) => (
                          <option key={b} value={b}>
                            {b}
                          </option>
                        ))}
                      </select>
                    </div>
                    <div>
                      <label htmlFor="tr-ref" className="block text-sm font-medium text-slate-700 mb-1">
                        N.º de referencia / comprobante *
                      </label>
                      <input
                        id="tr-ref"
                        type="text"
                        maxLength={40}
                        value={transferencia.referencia}
                        onChange={(e) => setTransferencia((p) => ({ ...p, referencia: e.target.value }))}
                        className={CAMPO}
                      />
                    </div>
                  </div>

                  <div>
                    <label htmlFor="tr-voucher" className="block text-sm font-medium text-slate-700 mb-1">
                      Voucher de la transferencia *
                    </label>
                    <input
                      id="tr-voucher"
                      type="file"
                      accept="image/png,image/jpeg,image/webp,application/pdf"
                      onChange={(e) => setVoucher(e.target.files?.[0] ?? null)}
                      className="block w-full text-sm text-slate-600 min-h-[44px] file:mr-3 file:min-h-[44px] file:rounded-lg file:border-0 file:bg-indigo-50 file:px-4 file:py-2 file:font-semibold file:text-indigo-700 hover:file:bg-indigo-100"
                    />
                    <p className="text-xs text-slate-500 mt-1">Imagen (PNG, JPG, WEBP) o PDF, máximo 2 MB.</p>
                  </div>
                </div>
              )}

              {/* ===== EFECTIVO CONTRA ENTREGA ===== */}
              {datos.metodoPago === METODOS_PAGO[2] && (
                <div>
                  <h3 className="font-semibold text-slate-800 mb-4">Datos para la entrega</h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label htmlFor="ef-recibe" className="block text-sm font-medium text-slate-700 mb-1">
                        ¿Quién recibe el pedido? *
                      </label>
                      <input
                        id="ef-recibe"
                        type="text"
                        placeholder="Nombre de quien recibe"
                        value={efectivo.recibe}
                        onChange={(e) => setEfectivo((p) => ({ ...p, recibe: e.target.value }))}
                        className={CAMPO}
                      />
                    </div>
                    <div>
                      <label htmlFor="ef-franja" className="block text-sm font-medium text-slate-700 mb-1">
                        Franja horaria de entrega *
                      </label>
                      <select
                        id="ef-franja"
                        value={efectivo.franja}
                        onChange={(e) => setEfectivo((p) => ({ ...p, franja: e.target.value }))}
                        className={`${CAMPO} bg-white`}
                      >
                        <option value="">Selecciona una franja</option>
                        {FRANJAS.map((f) => (
                          <option key={f} value={f}>
                            {f}
                          </option>
                        ))}
                      </select>
                    </div>
                    <div>
                      <label htmlFor="ef-paga" className="block text-sm font-medium text-slate-700 mb-1">
                        ¿Con cuánto vas a pagar? ($) *
                      </label>
                      <input
                        id="ef-paga"
                        type="number"
                        inputMode="decimal"
                        min={totalPrice}
                        step="0.01"
                        placeholder={totalPrice.toFixed(2)}
                        value={efectivo.pagaCon}
                        onChange={(e) => setEfectivo((p) => ({ ...p, pagaCon: e.target.value }))}
                        className={CAMPO}
                      />
                      {Number(efectivo.pagaCon) > totalPrice && (
                        <p className="text-xs text-slate-500 mt-1">
                          Te daremos de cambio ${(Number(efectivo.pagaCon) - totalPrice).toFixed(2)}.
                        </p>
                      )}
                    </div>
                    <div className="md:col-span-2">
                      <label htmlFor="ef-ind" className="block text-sm font-medium text-slate-700 mb-1">
                        Referencia / indicaciones para el repartidor
                      </label>
                      <textarea
                        id="ef-ind"
                        rows={3}
                        maxLength={200}
                        placeholder="Ej.: casa azul frente al parque, tocar el timbre"
                        value={efectivo.indicaciones}
                        onChange={(e) => setEfectivo((p) => ({ ...p, indicaciones: e.target.value }))}
                        className={`${CAMPO} resize-y`}
                      />
                    </div>
                  </div>
                  <p className="text-xs text-slate-500 mt-4">💵 Pagas en efectivo al recibir tu pedido.</p>
                </div>
              )}
            </div>
          </div>

          {error && (
            <div
              ref={errorRef}
              role="alert"
              className="bg-red-50 text-red-600 p-3 rounded-lg text-sm border border-red-200"
            >
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={procesando || !online}
            className="w-full min-h-[44px] bg-indigo-600 text-white py-3 rounded-lg font-semibold hover:bg-indigo-700 transition disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {procesando ? "Procesando pedido..." : "Confirmar Pedido"}
          </button>
          {!online && (
            <p role="alert" className="text-sm text-amber-700 bg-amber-50 border border-amber-200 rounded-lg p-3">
              Sin conexión: el pago requiere internet. Podrás confirmar tu pedido cuando vuelva la conexión.
            </p>
          )}
        </form>

        {/* ===== RESUMEN DEL PEDIDO ===== */}
        <div className="bg-white p-6 rounded-lg border border-slate-200 shadow-sm h-fit">
          <h2 className="text-lg font-bold text-slate-800 mb-4">Resumen del Pedido</h2>
          <div className="space-y-3 border-b border-slate-100 pb-4 mb-4">
            {cart.map((item) => (
              <div key={item.id} className="flex justify-between text-sm">
                <span className="text-slate-600">
                  {item.nombre} × {item.cantidad}
                </span>
                <span className="font-semibold text-slate-800">${(item.precio * item.cantidad).toFixed(2)}</span>
              </div>
            ))}
          </div>
          <div className="flex justify-between items-center">
            <span className="text-slate-800 font-bold">Total</span>
            <span className="text-2xl font-bold text-indigo-600">${totalPrice.toFixed(2)}</span>
          </div>
          <p className="text-xs text-slate-500 mt-3">El total final lo confirma el servidor con los precios vigentes.</p>
        </div>
      </div>
    </div>
  );
};

export default Checkout;
