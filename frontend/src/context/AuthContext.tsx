// src/context/AuthContext.tsx
import { createContext, useContext, useState, type ReactNode } from 'react';

// 1. Tipos de rol que maneja la aplicación
export type Rol = 'admin' | 'cliente';

// 2. Usuario autenticado: correo + rol (lo entrega la API en /api/login)
export interface Usuario {
  email: string;
  rol: Rol;
}

interface AuthContextType {
  isAuthenticated: boolean;
  user: Usuario | null;
  token: string | null; // token que devuelve la API (se conserva del Tema 4)
  login: (usuario: Usuario, token?: string) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth debe ser usado dentro de un AuthProvider');
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

// Lee el usuario guardado en localStorage (para que la sesión no se pierda al refrescar con F5).
// Si el dato no existe o está corrupto, devolvemos null en lugar de romper la app.
const leerUsuarioGuardado = (): Usuario | null => {
  try {
    const guardado = localStorage.getItem('usuario');
    if (!guardado) return null;
    const data = JSON.parse(guardado);
    if (typeof data?.email === 'string' && (data.rol === 'admin' || data.rol === 'cliente')) {
      return { email: data.email, rol: data.rol };
    }
  } catch {
    // JSON inválido: se ignora y se trata como sesión cerrada
  }
  return null;
};

export const AuthProvider = ({ children }: AuthProviderProps) => {
  const [user, setUser] = useState<Usuario | null>(leerUsuarioGuardado);
  const [token, setToken] = useState<string | null>(() =>
    localStorage.getItem('token')
  );

  // Estamos autenticados solo si hay un usuario (con su rol).
  const isAuthenticated = user !== null;

  const login = (usuario: Usuario, newToken?: string) => {
    setUser(usuario);
    localStorage.setItem('usuario', JSON.stringify(usuario));
    if (newToken) {
      setToken(newToken);
      localStorage.setItem('token', newToken);
    }
  };

  const logout = () => {
    setUser(null);
    setToken(null);
    localStorage.removeItem('usuario');
    localStorage.removeItem('token');
    // Limpiamos también las claves de la sesión anterior (Tema 4)
    localStorage.removeItem('userEmail');
  };

  return (
    <AuthContext.Provider value={{ isAuthenticated, user, token, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
};
