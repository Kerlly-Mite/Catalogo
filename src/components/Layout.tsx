import { useState } from 'react';
import { Outlet } from 'react-router-dom';
import Sidebar from './Sidebar';
import Navbar from './Navbar';

const Layout = () => {
  // Estado global (a nivel de Layout) que controla el colapso del Sidebar en escritorio
  const [isCollapsed, setIsCollapsed] = useState<boolean>(false);
  // Estado independiente que controla el drawer del Sidebar en mobile
  const [isMobileOpen, setIsMobileOpen] = useState<boolean>(false);

  const handleToggle = () => {
    setIsCollapsed((prev) => !prev);
    setIsMobileOpen((prev) => !prev);
  };

  return (
    <div className="flex h-screen bg-slate-50">
      <Sidebar
        isCollapsed={isCollapsed}
        isMobileOpen={isMobileOpen}
        onCloseMobile={() => setIsMobileOpen(false)}
      />
      {/* Área de Contenido Principal */}
      <div className="flex-1 flex flex-col overflow-hidden min-w-0">
        <Navbar isCollapsed={isCollapsed} onToggle={handleToggle} />
        <main className="flex-1 overflow-x-hidden overflow-y-auto p-4 sm:p-8">
          <Outlet />
        </main>
      </div>
    </div>
  );
};

export default Layout;