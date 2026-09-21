// src/context/AuthContext.tsx
import { createContext, useContext, useState, type ReactNode } from 'react';

interface AuthContextType {
  isAuthenticated: boolean;
  userEmail: string | null;
  token: string | null; // <-- token que nos devuelve la API
  login: (email: string, token: string) => void; // <-- ahora recibe también el token
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

export const AuthProvider = ({ children }: AuthProviderProps) => {
  // Inicializamos leyendo localStorage para que la sesión no se pierda al refrescar (F5).
  const [token, setToken] = useState<string | null>(() =>
    localStorage.getItem('token')
  );
  const [userEmail, setUserEmail] = useState<string | null>(() =>
    localStorage.getItem('userEmail')
  );

  // Estamos autenticados solo si existe un token.
  const isAuthenticated = token !== null;

  const login = (email: string, newToken: string) => {
    setToken(newToken);
    setUserEmail(email);
    // Persistimos la sesión en el navegador.
    localStorage.setItem('token', newToken);
    localStorage.setItem('userEmail', email);
  };

  const logout = () => {
    setToken(null);
    setUserEmail(null);
    localStorage.removeItem('token');
    localStorage.removeItem('userEmail');
  };

  return (
    <AuthContext.Provider
      value={{ isAuthenticated, userEmail, token, login, logout }}
    >
      {children}
    </AuthContext.Provider>
  );
};
