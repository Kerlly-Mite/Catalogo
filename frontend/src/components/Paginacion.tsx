interface PaginacionProps {
  page: number;
  limit: number;
  total: number;
  onChange: (page: number) => void;
}

// Controles Anterior / Siguiente (botones táctiles de al menos 44px).
const Paginacion = ({ page, limit, total, onChange }: PaginacionProps) => {
  const totalPaginas = Math.max(1, Math.ceil(total / limit));
  const boton =
    'min-h-[44px] px-4 rounded-lg border border-slate-300 text-slate-700 font-semibold hover:bg-slate-50 transition disabled:opacity-40 disabled:cursor-not-allowed';

  return (
    <div className="flex items-center justify-between gap-3 mt-4">
      <button className={boton} disabled={page <= 1} onClick={() => onChange(page - 1)}>
        Anterior
      </button>
      <span className="text-sm text-slate-600 text-center">
        Página {page} de {totalPaginas} · {total} en total
      </span>
      <button className={boton} disabled={page >= totalPaginas} onClick={() => onChange(page + 1)}>
        Siguiente
      </button>
    </div>
  );
};

export default Paginacion;
