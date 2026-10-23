import { useEffect, useState, type FormEvent } from 'react';
import { actualizarUsuario, getUsuarios, type UsuarioListado } from '../services/usuariosService';
import type { Pagina } from '../services/pedidosService';
import { mensajeDeError } from '../services/http';
import { useAuth, type Rol } from '../context/AuthContext';
import Paginacion from './Paginacion';

const LIMITE = 10;

// /admin/usuarios: listado paginado de cuentas con búsqueda, cambio de rol y activar/desactivar.
const AdminUsuarios = () => {
  const { user } = useAuth();
  const [pagina, setPagina] = useState<Pagina<UsuarioListado> | null>(null);
  const [page, setPage] = useState<number>(1);
  const [busqueda, setBusqueda] = useState<string>(''); // lo que se escribe
  const [q, setQ] = useState<string>(''); // lo que se aplicó al pulsar Buscar
  const [error, setError] = useState<string>('');
  const [aviso, setAviso] = useState<string>('');
  const [cargando, setCargando] = useState<boolean>(true);
  const [guardando, setGuardando] = useState<number | null>(null);

  useEffect(() => {
    let cancelado = false;
    getUsuarios(page, LIMITE, q)
      .then((p) => {
        if (cancelado) return;
        setPagina(p);
        setError('');
      })
      .catch((e) => !cancelado && setError(mensajeDeError(e)))
      .finally(() => !cancelado && setCargando(false));
    return () => {
      cancelado = true;
    };
  }, [page, q]);

  const buscar = (e: FormEvent) => {
    e.preventDefault();
    setQ(busqueda.trim());
    setPage(1);
  };

  const cambiar = async (u: UsuarioListado, cambios: { rol?: Rol; activo?: boolean }, pregunta: string) => {
    if (!window.confirm(pregunta)) return;
    setGuardando(u.id);
    setError('');
    setAviso('');
    try {
      const nuevo = await actualizarUsuario(u.id, cambios);
      setPagina((p) => (p ? { ...p, data: p.data.map((x) => (x.id === nuevo.id ? nuevo : x)) } : p));
      setAviso(`Cuenta de ${nuevo.email} actualizada.`);
    } catch (e) {
      setError(mensajeDeError(e));
    } finally {
      setGuardando(null);
    }
  };

  const boton =
    'min-h-[44px] px-3 rounded-lg border border-slate-300 text-slate-700 text-xs font-semibold hover:bg-slate-50 transition disabled:opacity-40 disabled:cursor-not-allowed';

  return (
    <div className="max-w-6xl mx-auto">
      <h1 className="text-2xl sm:text-3xl font-bold text-slate-900">Usuarios</h1>
      <p className="text-slate-500 mt-1 mb-6">Cuentas registradas: busca, cambia el rol o desactiva una cuenta (no podrá volver a iniciar sesión).</p>

      <form onSubmit={buscar} className="flex flex-col sm:flex-row gap-3 mb-4">
        <input
          value={busqueda}
          onChange={(e) => setBusqueda(e.target.value)}
          placeholder="Buscar por email o nombre"
          className="min-h-[44px] flex-1 px-3 rounded-lg border border-slate-300 bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
        <button type="submit" className="min-h-[44px] px-5 bg-indigo-600 text-white font-semibold rounded-lg hover:bg-indigo-700 transition">
          Buscar
        </button>
      </form>

      {error && <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-4 border border-red-200">{error}</div>}
      {aviso && <div className="bg-green-50 text-green-700 p-3 rounded-lg text-sm mb-4 border border-green-200">{aviso}</div>}
      {cargando && <p className="text-slate-500">Cargando...</p>}

      <div className="bg-white border border-slate-200 rounded-xl shadow-sm p-4 sm:p-6">
        {pagina && pagina.data.length === 0 ? (
          <p className="text-slate-500 text-sm">{q ? 'Ningún usuario coincide con la búsqueda.' : 'Todavía no hay usuarios.'}</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-slate-50 text-slate-600 text-left">
                <tr>
                  <th className="p-3">Usuario</th>
                  <th className="p-3">Rol</th>
                  <th className="p-3">Estado</th>
                  <th className="p-3">Código de referido</th>
                  <th className="p-3">Registro</th>
                  <th className="p-3">Acciones</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-200">
                {pagina?.data.map((u) => {
                  const esYo = u.email.toLowerCase() === user?.email.toLowerCase();
                  const ocupado = guardando === u.id;
                  const rolDestino: Rol = u.rol === 'admin' ? 'cliente' : 'admin';
                  return (
                    <tr key={u.id}>
                      <td className="p-3">
                        <p className="font-medium">{u.nombre || '—'}{esYo && <span className="ml-2 text-xs text-slate-400">(tú)</span>}</p>
                        <p className="text-xs text-slate-500">{u.email}</p>
                      </td>
                      <td className="p-3">
                        <span className={`text-xs font-semibold px-2 py-0.5 rounded-full ${u.rol === 'admin' ? 'bg-amber-100 text-amber-800' : 'bg-slate-100 text-slate-700'}`}>{u.rol}</span>
                      </td>
                      <td className="p-3">
                        <span className={`text-xs font-semibold px-2 py-0.5 rounded-full ${u.activo ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-700'}`}>{u.activo ? 'activa' : 'desactivada'}</span>
                      </td>
                      <td className="p-3 font-mono text-xs">{u.codigo_referido || '—'}</td>
                      <td className="p-3 whitespace-nowrap text-slate-500">{new Date(u.creado_en).toLocaleDateString('es-EC')}</td>
                      <td className="p-3">
                        <div className="flex flex-wrap gap-2">
                          <button
                            className={boton}
                            disabled={ocupado || esYo}
                            title={esYo ? 'No puedes cambiar tu propio rol' : undefined}
                            onClick={() => cambiar(u, { rol: rolDestino }, `¿Cambiar el rol de ${u.email} a "${rolDestino}"?`)}
                          >
                            Hacer {rolDestino}
                          </button>
                          <button
                            className={boton}
                            disabled={ocupado || esYo}
                            title={esYo ? 'No puedes desactivar tu propia cuenta' : undefined}
                            onClick={() =>
                              cambiar(
                                u,
                                { activo: !u.activo },
                                u.activo ? `¿Desactivar la cuenta de ${u.email}? No podrá volver a iniciar sesión.` : `¿Reactivar la cuenta de ${u.email}?`,
                              )
                            }
                          >
                            {u.activo ? 'Desactivar' : 'Activar'}
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
        {pagina && pagina.total > 0 && <Paginacion page={pagina.page} limit={pagina.limit} total={pagina.total} onChange={setPage} />}
      </div>
    </div>
  );
};

export default AdminUsuarios;
