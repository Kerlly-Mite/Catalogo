import { Link } from 'react-router-dom';
import { LayoutDashboard, ShoppingBag, Users, Sparkles, X } from 'lucide-react';

interface Props {
  isCollapsed: boolean;
  isMobileOpen: boolean;
  onCloseMobile: () => void;
}

const enlaces = [
  { to: '/', label: 'Panel', icono: LayoutDashboard },
  { to: '/catalogo', label: 'Catálogo', icono: ShoppingBag },
  { to: '/mi-red', label: 'Mi Red', icono: Users },
];

const Sidebar = ({ isCollapsed, isMobileOpen, onCloseMobile }: Props) => {
  return (
    <>
      {isMobileOpen && (
        <div
          className="fixed inset-0 bg-slate-900/50 z-40 md:hidden"
          onClick={onCloseMobile}
        />
      )}

      <aside
        className={`fixed md:static inset-y-0 left-0 z-50 bg-slate-900 text-white flex flex-col
          transition-all duration-200 w-64
          ${isMobileOpen ? 'translate-x-0' : '-translate-x-full'} md:translate-x-0
          ${isCollapsed ? 'md:w-20' : 'md:w-64'}
        `}
      >
        <div
          className={`p-6 text-2xl font-bold border-b border-slate-700 whitespace-nowrap overflow-hidden flex items-center justify-between ${
            isCollapsed ? 'md:justify-center md:px-0' : ''
          }`}
        >
          <span className={isCollapsed ? 'md:hidden' : ''}>MultiCatálogo</span>
          {isCollapsed && (
            <span className="hidden md:inline text-emerald-400">
              <Sparkles size={24} />
            </span>
          )}
          <button
            onClick={onCloseMobile}
            className="md:hidden text-slate-400 hover:text-white transition"
            aria-label="Cerrar menú"
          >
            <X size={22} />
          </button>
        </div>
        <nav className="flex-1 p-4 space-y-2">
          {enlaces.map((enlace) => {
            const Icono = enlace.icono;
            return (
              <Link
                key={enlace.to}
                to={enlace.to}
                onClick={onCloseMobile}
                className={`flex items-center gap-3 p-3 rounded hover:bg-slate-800 hover:text-emerald-400 transition ${
                  isCollapsed ? 'md:justify-center' : ''
                }`}
                title={enlace.label}
              >
                <Icono size={20} className="shrink-0" />
                <span className={isCollapsed ? 'md:hidden' : ''}>{enlace.label}</span>
              </Link>
            );
          })}
        </nav>
      </aside>
    </>
  );
};

export default Sidebar;
