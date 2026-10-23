import { useEffect, useState, type ChangeEvent, type FormEvent } from 'react';
import {
  actualizarProducto,
  crearProducto,
  desactivarProducto,
  getProductosAdmin,
  subirImagen,
  type ProductoAdmin,
  type ProductoForm,
} from '../services/adminProductosService';
import type { Pagina } from '../services/pedidosService';
import { mensajeDeError } from '../services/http';
import { prepararImagen } from '../utils/imagen';
import { categorias } from '../data/productos';
import Paginacion from './Paginacion';

const LIMITE = 10;

// Campos del formulario como texto (los números se convierten al guardar).
interface FormTexto {
  nombre: string;
  descripcion: string;
  precio: string;
  categoria: string;
  img: string;
  galeria: string[]; // URLs de las fotos de la galería
  stock: string;
}

const FORM_VACIO: FormTexto = {
  nombre: '',
  descripcion: '',
  precio: '',
  categoria: categorias[0] ?? '',
  img: '',
  galeria: [],
  stock: '0',
};

const aFormTexto = (p: ProductoAdmin): FormTexto => ({
  nombre: p.nombre,
  descripcion: p.descripcion,
  precio: String(p.precio),
  categoria: p.categoria,
  img: p.img,
  galeria: [...p.galeria],
  stock: String(p.stock),
});

// /admin/productos: tabla de productos (activos e inactivos) con alta, edición y activar/desactivar.
const AdminProductos = () => {
  const [page, setPage] = useState<number>(1);
  const [q, setQ] = useState<string>('');
  const [busqueda, setBusqueda] = useState<string>('');
  const [recarga, setRecarga] = useState<number>(0);
  const [pagina, setPagina] = useState<Pagina<ProductoAdmin> | null>(null);
  const [error, setError] = useState<string>('');
  const [aviso, setAviso] = useState<string>('');
  const [cargando, setCargando] = useState<boolean>(true);

  // Modal: null = cerrado; { id: null } = nuevo; { id: n } = editando el producto n.
  const [modal, setModal] = useState<{ id: number | null } | null>(null);
  const [form, setForm] = useState<FormTexto>(FORM_VACIO);
  const [errorForm, setErrorForm] = useState<string>('');
  const [guardando, setGuardando] = useState<boolean>(false);
  const [subiendo, setSubiendo] = useState<boolean>(false);

  useEffect(() => {
    let cancelado = false;
    getProductosAdmin(busqueda, page, LIMITE)
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
  }, [busqueda, page, recarga]);

  const productos = pagina?.data ?? [];

  const abrirNuevo = () => {
    setForm(FORM_VACIO);
    setErrorForm('');
    setModal({ id: null });
  };

  const abrirEdicion = (p: ProductoAdmin) => {
    setForm(aFormTexto(p));
    setErrorForm('');
    setModal({ id: p.id });
  };

  const buscar = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setBusqueda(q.trim());
    setPage(1);
  };

  const guardar = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!modal) return;
    setErrorForm('');
    setGuardando(true);

    const datos: ProductoForm = {
      nombre: form.nombre.trim(),
      descripcion: form.descripcion.trim(),
      precio: Number(form.precio),
      categoria: form.categoria.trim(),
      img: form.img.trim(),
      galeria: form.galeria,
      stock: Number(form.stock),
    };

    if (!datos.img) {
      setErrorForm('Sube una foto principal del producto.');
      setGuardando(false);
      return;
    }

    try {
      if (modal.id === null) {
        const creado = await crearProducto(datos);
        setAviso(`Producto "${creado.nombre}" creado.`);
      } else {
        const editado = await actualizarProducto(modal.id, datos);
        setAviso(`Producto "${editado.nombre}" actualizado.`);
      }
      setModal(null);
      setRecarga((n) => n + 1);
    } catch (err) {
      setErrorForm(mensajeDeError(err));
    } finally {
      setGuardando(false);
    }
  };

  // Reduce la foto, la sube a la API y devuelve su URL pública.
  const subirArchivo = async (archivo: File): Promise<string> => subirImagen(await prepararImagen(archivo));

  const elegirPrincipal = async (e: ChangeEvent<HTMLInputElement>) => {
    const archivo = e.target.files?.[0];
    e.target.value = ''; // permite volver a elegir la misma foto
    if (!archivo) return;
    setErrorForm('');
    setSubiendo(true);
    try {
      const url = await subirArchivo(archivo);
      setForm((prev) => ({ ...prev, img: url }));
    } catch (err) {
      setErrorForm(mensajeDeError(err));
    } finally {
      setSubiendo(false);
    }
  };

  const elegirGaleria = async (e: ChangeEvent<HTMLInputElement>) => {
    const archivos = Array.from(e.target.files ?? []);
    e.target.value = '';
    if (archivos.length === 0) return;
    if (form.galeria.length + archivos.length > 10) {
      setErrorForm('La galería admite como máximo 10 fotos.');
      return;
    }
    setErrorForm('');
    setSubiendo(true);
    try {
      const urls: string[] = [];
      for (const archivo of archivos) urls.push(await subirArchivo(archivo));
      setForm((prev) => ({ ...prev, galeria: [...prev.galeria, ...urls] }));
    } catch (err) {
      setErrorForm(mensajeDeError(err));
    } finally {
      setSubiendo(false);
    }
  };

  const quitarDeGaleria = (indice: number) =>
    setForm((prev) => ({ ...prev, galeria: prev.galeria.filter((_, i) => i !== indice) }));

  const alternarActivo = async (p: ProductoAdmin) => {
    const accion = p.activo ? 'desactivar' : 'activar';
    if (!window.confirm(`¿Quieres ${accion} "${p.nombre}"?${p.activo ? ' Dejará de verse en el catálogo.' : ''}`)) return;
    setError('');
    setAviso('');
    try {
      if (p.activo) await desactivarProducto(p.id);
      else await actualizarProducto(p.id, { activo: true });
      setAviso(`Producto "${p.nombre}" ${p.activo ? 'desactivado' : 'activado'}.`);
      setRecarga((n) => n + 1);
    } catch (err) {
      setError(mensajeDeError(err));
    }
  };

  const campo = 'w-full min-h-[44px] px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none';
  const opcionesCategoria = categorias.includes(form.categoria) || !form.categoria ? categorias : [...categorias, form.categoria];

  return (
    <div className="max-w-6xl mx-auto">
      <div className="flex flex-wrap items-center justify-between gap-4 mb-6">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold text-slate-900">Productos</h1>
          <p className="text-slate-500 mt-1">Administra el catálogo. Los inactivos no se ven en la tienda.</p>
        </div>
        <button
          onClick={abrirNuevo}
          className="min-h-[44px] px-5 bg-indigo-600 text-white font-semibold rounded-lg hover:bg-indigo-700 transition"
        >
          + Nuevo producto
        </button>
      </div>

      <form onSubmit={buscar} className="flex gap-2 mb-4">
        <input
          value={q}
          onChange={(e) => setQ(e.target.value)}
          placeholder="Buscar por nombre, descripción o categoría..."
          className={campo}
        />
        <button type="submit" className="min-h-[44px] px-5 border border-slate-300 rounded-lg font-semibold text-slate-700 hover:bg-slate-50">
          Buscar
        </button>
      </form>

      {error && <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-4 border border-red-200">{error}</div>}
      {aviso && <div className="bg-green-50 text-green-700 p-3 rounded-lg text-sm mb-4 border border-green-200">{aviso}</div>}
      {cargando && <p className="text-slate-500">Cargando productos...</p>}
      {!cargando && !error && productos.length === 0 && <p className="text-slate-500">No se encontraron productos.</p>}

      {productos.length > 0 && (
        <div className="bg-white border border-slate-200 rounded-xl shadow-sm overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-slate-50 text-slate-600 text-left">
              <tr>
                <th className="p-3">Imagen</th>
                <th className="p-3">Nombre</th>
                <th className="p-3">Categoría</th>
                <th className="p-3">Precio</th>
                <th className="p-3">Stock</th>
                <th className="p-3">Estado</th>
                <th className="p-3">Acciones</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-200">
              {productos.map((p) => (
                <tr key={p.id} className={p.activo ? '' : 'bg-slate-50 text-slate-500'}>
                  <td className="p-3">
                    <img src={p.img} alt={p.nombre} className="w-12 h-12 object-cover rounded-lg bg-slate-100" />
                  </td>
                  <td className="p-3 font-medium">{p.nombre}</td>
                  <td className="p-3">{p.categoria}</td>
                  <td className="p-3 whitespace-nowrap">${p.precio.toFixed(2)}</td>
                  <td className="p-3">{p.stock}</td>
                  <td className="p-3">
                    <span className={`px-2.5 py-1 rounded-full text-xs font-semibold ${p.activo ? 'bg-green-100 text-green-800' : 'bg-slate-200 text-slate-600'}`}>
                      {p.activo ? 'Activo' : 'Inactivo'}
                    </span>
                  </td>
                  <td className="p-3">
                    <div className="flex flex-wrap gap-2">
                      <button
                        onClick={() => abrirEdicion(p)}
                        className="min-h-[44px] px-3 rounded-lg border border-slate-300 text-slate-700 font-semibold hover:bg-slate-50"
                      >
                        Editar
                      </button>
                      <button
                        onClick={() => alternarActivo(p)}
                        className={`min-h-[44px] px-3 rounded-lg border font-semibold ${
                          p.activo ? 'border-red-300 text-red-600 hover:bg-red-50' : 'border-green-300 text-green-700 hover:bg-green-50'
                        }`}
                      >
                        {p.activo ? 'Desactivar' : 'Activar'}
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {pagina && pagina.total > 0 && (
        <Paginacion page={pagina.page} limit={pagina.limit} total={pagina.total} onChange={setPage} />
      )}

      {modal && (
        <div className="fixed inset-0 z-[60] bg-black/50 flex items-center justify-center p-4">
          <form onSubmit={guardar} className="bg-white w-full max-w-lg rounded-xl shadow-xl p-5 sm:p-6 max-h-[90vh] overflow-y-auto space-y-4">
            <h2 className="text-xl font-bold text-slate-900">{modal.id === null ? 'Nuevo producto' : 'Editar producto'}</h2>

            {errorForm && <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm border border-red-200">{errorForm}</div>}

            <label className="block text-sm font-medium text-slate-700">
              Nombre
              <input className={`${campo} mt-1`} value={form.nombre} onChange={(e) => setForm({ ...form, nombre: e.target.value })} required maxLength={150} />
            </label>
            <label className="block text-sm font-medium text-slate-700">
              Descripción
              <textarea className={`${campo} mt-1`} rows={3} value={form.descripcion} onChange={(e) => setForm({ ...form, descripcion: e.target.value })} />
            </label>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <label className="block text-sm font-medium text-slate-700">
                Precio (USD)
                <input className={`${campo} mt-1`} type="number" step="0.01" min="0.01" value={form.precio} onChange={(e) => setForm({ ...form, precio: e.target.value })} required />
              </label>
              <label className="block text-sm font-medium text-slate-700">
                Stock
                <input className={`${campo} mt-1`} type="number" step="1" min="0" value={form.stock} onChange={(e) => setForm({ ...form, stock: e.target.value })} required />
              </label>
            </div>
            <label className="block text-sm font-medium text-slate-700">
              Categoría
              <select className={`${campo} mt-1 bg-white`} value={form.categoria} onChange={(e) => setForm({ ...form, categoria: e.target.value })}>
                {opcionesCategoria.map((c) => (
                  <option key={c} value={c}>
                    {c}
                  </option>
                ))}
              </select>
            </label>
            <div>
              <p className="text-sm font-medium text-slate-700 mb-1">Imagen principal</p>
              <div className="flex items-center gap-3">
                {form.img ? (
                  <img src={form.img} alt="Vista previa" className="w-20 h-20 object-cover rounded-lg border border-slate-200 bg-slate-100" />
                ) : (
                  <div className="w-20 h-20 rounded-lg border border-dashed border-slate-300 bg-slate-50 flex items-center justify-center text-xs text-slate-400 text-center px-1">
                    Sin imagen
                  </div>
                )}
                <label className="min-h-[44px] px-4 inline-flex items-center rounded-lg border border-slate-300 font-semibold text-slate-700 hover:bg-slate-50 cursor-pointer">
                  {subiendo ? 'Subiendo...' : form.img ? 'Cambiar foto' : 'Subir foto'}
                  <input type="file" accept="image/jpeg,image/png,image/webp" className="hidden" disabled={subiendo} onChange={elegirPrincipal} />
                </label>
              </div>
              <p className="text-xs text-slate-500 mt-1">Elige una foto de tu dispositivo (JPG, PNG o WebP). Se ajusta sola.</p>
              <details className="mt-2">
                <summary className="text-xs text-slate-500 cursor-pointer">¿Prefieres usar un enlace?</summary>
                <input className={`${campo} mt-2`} value={form.img} onChange={(e) => setForm({ ...form, img: e.target.value })} placeholder="https://..." />
              </details>
            </div>

            <div>
              <p className="text-sm font-medium text-slate-700 mb-1">Galería (opcional)</p>
              {form.galeria.length > 0 && (
                <div className="flex flex-wrap gap-3 mb-2">
                  {form.galeria.map((url, i) => (
                    <div key={`${url}-${i}`} className="relative">
                      <img src={url} alt={`Galería ${i + 1}`} className="w-16 h-16 object-cover rounded-lg border border-slate-200 bg-slate-100" />
                      <button
                        type="button"
                        aria-label="Quitar foto"
                        onClick={() => quitarDeGaleria(i)}
                        className="absolute -top-2 -right-2 w-6 h-6 rounded-full bg-red-600 text-white text-sm font-bold leading-none"
                      >
                        ×
                      </button>
                    </div>
                  ))}
                </div>
              )}
              <label className="min-h-[44px] px-4 inline-flex items-center rounded-lg border border-slate-300 font-semibold text-slate-700 hover:bg-slate-50 cursor-pointer">
                {subiendo ? 'Subiendo...' : '+ Agregar fotos'}
                <input type="file" multiple accept="image/jpeg,image/png,image/webp" className="hidden" disabled={subiendo} onChange={elegirGaleria} />
              </label>
            </div>

            <div className="flex gap-3 pt-2">
              <button type="button" onClick={() => setModal(null)} className="flex-1 min-h-[44px] border border-slate-300 rounded-lg font-semibold text-slate-700 hover:bg-slate-50">
                Cancelar
              </button>
              <button type="submit" disabled={guardando || subiendo} className="flex-1 min-h-[44px] bg-indigo-600 text-white rounded-lg font-semibold hover:bg-indigo-700 disabled:opacity-60">
                {guardando ? 'Guardando...' : 'Guardar'}
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
};

export default AdminProductos;
