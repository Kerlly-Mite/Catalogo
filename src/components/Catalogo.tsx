import { useEffect, useState } from 'react';
import { useCart, type Producto } from '../context/CartContext';
import { getProductos } from '../services/api';

const Catalogo = () => {
  const { addToCart } = useCart();

  // Ya NO hay arreglo quemado: los productos vienen de la API.
  const [productos, setProductos] = useState<Producto[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');

  // useEffect con arreglo de dependencias vacío: se ejecuta una sola vez al montar el componente.
  useEffect(() => {
    const cargarProductos = async () => {
      try {
        const data = await getProductos(); // GET /api/productos
        setProductos(data);
      } catch {
        setError('No se pudo conectar con el servidor. Verifica que el backend esté corriendo en el puerto 3000.');
      } finally {
        setLoading(false);
      }
    };

    cargarProductos();
  }, []);

  // Estado de carga
  if (loading) {
    return (
      <div>
        <h1 className="text-2xl font-bold text-slate-800 mb-6">Catálogo de Productos</h1>
        <p className="text-slate-500">Cargando productos...</p>
      </div>
    );
  }

  // Estado de error
  if (error) {
    return (
      <div>
        <h1 className="text-2xl font-bold text-slate-800 mb-6">Catálogo de Productos</h1>
        <div className="bg-red-50 text-red-600 p-4 rounded-lg border border-red-200">{error}</div>
      </div>
    );
  }

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-800 mb-6">Catálogo de Productos</h1>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        {productos.map((prod) => (
          <div
            key={prod.id}
            className="bg-white rounded-lg overflow-hidden border border-slate-200 shadow-sm flex flex-col"
          >
            <img src={prod.img} alt={prod.nombre} className="w-full h-40 object-cover" />
            <div className="p-4 flex flex-col flex-1">
              <h3 className="font-semibold text-slate-700">{prod.nombre}</h3>
              <p className="text-indigo-600 font-bold mt-2 mb-4">${prod.precio.toFixed(2)}</p>
              <button
                onClick={() => addToCart(prod)}
                className="mt-auto w-full bg-slate-900 text-white py-2 rounded text-sm hover:bg-indigo-600 transition"
              >
                Añadir al Carrito
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default Catalogo;
