// src/components/Navbar.tsx
import { useEffect, useRef, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useCart } from "../context/CartContext";
import { useAuth } from "../context/AuthContext";
import InstallButton from "../pwa/InstallButton";

// Definimos la interfaz para las props
interface NavbarProps {
  toggleSidebar: () => void;
}

const Navbar = ({ toggleSidebar }: NavbarProps) => {
  const { totalItems } = useCart();
  const { logout, user } = useAuth();
  const navigate = useNavigate();

  // Menú del perfil: se abre al hacer clic/tocar el avatar y se cierra al tocar fuera o con Escape.
  const [menuAbierto, setMenuAbierto] = useState<boolean>(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!menuAbierto) return;
    const cerrarSiTocaFuera = (e: PointerEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) setMenuAbierto(false);
    };
    const cerrarConEscape = (e: KeyboardEvent) => {
      if (e.key === "Escape") setMenuAbierto(false);
    };
    document.addEventListener("pointerdown", cerrarSiTocaFuera);
    document.addEventListener("keydown", cerrarConEscape);
    return () => {
      document.removeEventListener("pointerdown", cerrarSiTocaFuera);
      document.removeEventListener("keydown", cerrarConEscape);
    };
  }, [menuAbierto]);

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  return (
    <header className="h-16 bg-white border-b border-slate-200 flex items-center justify-between px-4 md:px-8">
      <div className="flex items-center gap-2 md:gap-4">
        {/* Botón Hamburguesa siempre visible */}
        <button
          onClick={toggleSidebar}
          className="p-2 rounded-md hover:bg-slate-100 transition text-slate-600 focus:outline-none focus:ring-2 focus:ring-indigo-500"
          aria-label="Alternar menú"
        >
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
        <h2 className="text-slate-600 font-medium text-base md:text-lg hidden sm:block">
          {user?.rol === "admin" ? "Panel de Administración" : "Tienda MultiCatálogo"}
        </h2>
      </div>

      <div className="flex items-center gap-4 md:gap-6">
        <InstallButton />
        <Link
          to="/carrito"
          className="relative p-2 hover:bg-slate-100 rounded-full transition"
        >
          <span className="text-xl">🛒</span>
          {totalItems > 0 && (
            <span className="absolute top-0 right-0 bg-indigo-600 text-white text-xs font-bold w-5 h-5 flex items-center justify-center rounded-full transform translate-x-1 -translate-y-1">
              {totalItems}
            </span>
          )}
        </Link>

        <div className="flex items-center gap-2 md:gap-4">
          <span className="text-xs md:text-sm text-slate-500 hidden sm:block">{user?.email}</span>
          <span
            className={`hidden md:inline-block text-xs font-semibold px-2 py-1 rounded-full uppercase ${
              user?.rol === "admin"
                ? "bg-amber-100 text-amber-700"
                : "bg-indigo-100 text-indigo-700"
            }`}
          >
            {user?.rol}
          </span>

          <div ref={menuRef} className="relative">
            <button
              type="button"
              onClick={() => setMenuAbierto((abierto) => !abierto)}
              aria-haspopup="menu"
              aria-expanded={menuAbierto}
              aria-label="Menú de usuario"
              className="w-10 h-10 rounded-full bg-slate-200 overflow-hidden border border-slate-300 flex items-center justify-center focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <img
                src="https://fastly.picsum.photos/id/64/4326/2884.jpg?hmac=9_SzX666YRpR_fOyYStXpfSiJ_edO3ghlSRnH2w09Kg"
                alt="Avatar del usuario"
                className="w-full h-full object-cover"
              />
            </button>

            {menuAbierto && (
              <div
                role="menu"
                className="absolute right-0 top-full mt-2 w-60 bg-white border border-slate-200 rounded-lg shadow-lg z-50 overflow-hidden"
              >
                <div className="px-4 py-3 border-b border-slate-100">
                  <p className="text-sm font-semibold text-slate-800 truncate">{user?.email}</p>
                  <p className="text-xs text-slate-500 uppercase">{user?.rol}</p>
                </div>
                <button
                  type="button"
                  role="menuitem"
                  onClick={handleLogout}
                  className="w-full min-h-[44px] text-left px-4 py-2 text-sm text-red-600 font-semibold hover:bg-red-50 transition-colors"
                >
                  Cerrar Sesión
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  );
};

export default Navbar;
