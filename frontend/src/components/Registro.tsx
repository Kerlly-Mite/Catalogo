import { useState, type FormEvent } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { registerRequest } from '../services/api';

// Registro público: crea la cuenta (rol cliente) y, si viene un código de referido
// (?ref=CODIGO en la URL), el nuevo nodo cuelga de quien invitó en la red multinivel.
const Registro = () => {
  const [searchParams] = useSearchParams();
  const [nombre, setNombre] = useState<string>('');
  const [email, setEmail] = useState<string>('');
  const [password, setPassword] = useState<string>('');
  const [ref, setRef] = useState<string>(searchParams.get('ref') ?? '');
  const [error, setError] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(false);

  const navigate = useNavigate();
  const { login } = useAuth();

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const data = await registerRequest(nombre, email, password, ref);
      // El backend ya devuelve el token: iniciamos sesión automáticamente.
      login({ email: data.email, rol: 'cliente' }, data.token);
      navigate('/tienda');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Error de conexión con el servidor');
    } finally {
      setLoading(false);
    }
  };

  const campo =
    'w-full px-4 py-3 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition';

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-100 px-4 py-8">
      <div className="max-w-md w-full bg-white rounded-xl shadow-md p-6 sm:p-8 border border-slate-200">
        <div className="text-center mb-6">
          <h2 className="text-3xl font-bold text-slate-900">Crear cuenta</h2>
          <p className="text-slate-500 mt-2">Únete a MultiCatálogo</p>
        </div>

        {error && (
          <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-6 text-center border border-red-200">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-2">Nombre completo</label>
            <input
              type="text"
              value={nombre}
              onChange={(e) => setNombre(e.target.value)}
              className={campo}
              placeholder="Ana Pérez"
              minLength={2}
              maxLength={100}
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-2">Correo electrónico</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className={campo}
              placeholder="tucorreo@ejemplo.com"
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-2">Contraseña</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className={campo}
              placeholder="Mínimo 8 caracteres"
              minLength={8}
              maxLength={72}
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-2">
              Código de referido <span className="text-slate-400 font-normal">(opcional)</span>
            </label>
            <input
              type="text"
              value={ref}
              onChange={(e) => setRef(e.target.value.toUpperCase())}
              className={`${campo} font-mono`}
              placeholder="Déjalo vacío si no tienes uno"
            />
            <p className="text-xs text-slate-400 mt-1">
              Solo escríbelo si alguien te invitó. Si lo dejas vacío, igual puedes crear tu cuenta.
            </p>
          </div>
          <button
            type="submit"
            disabled={loading}
            className="w-full min-h-[44px] bg-indigo-600 text-white font-bold py-3 rounded-lg hover:bg-indigo-700 transition disabled:opacity-60 disabled:cursor-not-allowed"
          >
            {loading ? 'Creando cuenta...' : 'Crear cuenta'}
          </button>
        </form>

        <p className="text-center text-sm text-slate-600 mt-6">
          ¿Ya tienes cuenta?{' '}
          <Link to="/login" className="text-indigo-600 font-semibold hover:underline">
            Inicia sesión
          </Link>
        </p>
      </div>
    </div>
  );
};

export default Registro;
