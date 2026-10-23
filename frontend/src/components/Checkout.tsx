// src/components/Checkout.tsx
// Checkout REAL: envía el pedido a POST /api/checkout (transacción ACID en el backend):
// - Formulario de envío con validación básica
// - Resumen del pedido en tiempo real
// - Si el servidor responde bien: vacía el carrito y muestra la confirmación con el pedido REAL
// - Si falla (stock, validación, caída de la BD): muestra el mensaje y CONSERVA el carrito
import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useCart } from '../context/CartContext';
import { useAuth } from '../context/AuthContext';
import { crearPedido } from '../services/checkoutService';
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
  const online = useOnlineStatus(); // PWA: sin conexión no se puede pagar
  const [procesando, setProcesando] = useState<boolean>(false);
  const [error, setError] = useState<string>("");

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

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (procesando) return; // evita enviar el pedido dos veces con doble clic
    setError("");
    setProcesando(true);

    try {
      const pedidoCreado = await crearPedido({
        cliente: {
          nombre: datos.nombre.trim(),
          email: emailCuenta,
          direccion: datos.direccion.trim(),
          ciudad: datos.ciudad.trim(),
          telefono: datos.telefono.trim(),
        },
        metodoPago: datos.metodoPago,
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
          {error && (
            <div role="alert" className="bg-red-50 text-red-600 p-3 rounded-lg text-sm border border-red-200">
              {error}
            </div>
          )}

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
                  className="flex items-center gap-3 min-h-[44px] p-3 rounded-lg border border-slate-200 cursor-pointer hover:border-indigo-300 transition"
                >
                  <input
                    type="radio"
                    name="metodoPago"
                    value={metodo}
                    checked={datos.metodoPago === metodo}
                    onChange={(e) => handleChange("metodoPago", e.target.value)}
                    className="accent-indigo-600"
                  />
                  <span className="text-sm text-slate-700">{metodo}</span>
                </label>
              ))}
            </div>
          </div>

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
