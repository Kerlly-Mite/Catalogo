// src/data/red.ts
// Tipos de la red multinivel de referidos y funciones puras para calcular
// totales y el nivel alcanzado. Los DATOS vienen de la API (/api/red/mia), no de este archivo.

export interface Referido {
  id: number;
  nombre: string;
  nivel: number;     // 1 = directo, 2 = indirecto, 3 = tercer nivel
  ventas: number;    // ventas acumuladas en dólares
  saldo?: number;    // comisiones devengadas (la API lo envía en /api/red/mia)
  hijos?: Referido[]; // referidos de un nivel más abajo
}

// Porcentaje de comisión que recibe el usuario por las ventas de cada nivel
export const TASA_COMISION: Record<number, number> = {
  1: 0.10, // 10 % de las ventas de nivel 1
  2: 0.05, // 5 % de las ventas de nivel 2
  3: 0.02, // 2 % de las ventas de nivel 3
};

// Total de referidos de la red (sin contar al usuario raíz)
export const contarRed = (raiz: Referido): number => {
  const directos = raiz.hijos ?? [];
  return directos.reduce((total, hijo) => total + 1 + contarRed(hijo), 0);
};

// Suma de ventas de toda la red (excluye las ventas propias del usuario raíz)
export const sumarVentasRed = (raiz: Referido): number => {
  const directos = raiz.hijos ?? [];
  return directos.reduce((total, hijo) => total + hijo.ventas + sumarVentasRed(hijo), 0);
};

// Comisión que recibe el usuario por las ventas de cada referido (ventas * tasa del nivel)
export const comisionDeReferido = (ref: Referido): number =>
  ref.ventas * (TASA_COMISION[ref.nivel] ?? 0);

// Comisiones totales del mes acumuladas sobre toda la red
export const sumarComisiones = (raiz: Referido): number => {
  const directos = raiz.hijos ?? [];
  return directos.reduce(
    (total, hijo) => total + comisionDeReferido(hijo) + sumarComisiones(hijo),
    0
  );
};

// Nivel alcanzado por el usuario según la cantidad de referidos directos
export const nivelAlcanzado = (referidosDirectos: number): string => {
  if (referidosDirectos >= 6) return "Diamante";
  if (referidosDirectos >= 4) return "Oro";
  if (referidosDirectos >= 2) return "Plata";
  return "Bronce";
};
