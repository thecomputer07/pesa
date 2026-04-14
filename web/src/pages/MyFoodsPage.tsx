import { useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { apiFetch, type FoodSearchResult } from "../lib/api";
import ResultCard from "../components/ResultCard";

export default function MyFoodsPage() {
  const [q, setQ] = useState("");
  const { data, isLoading, error } = useQuery({
    queryKey: ["foods-mine"],
    queryFn: async () => {
      const res = await apiFetch("/foods/mine");
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return (await res.json()) as { results: FoodSearchResult[]; total: number };
    },
    staleTime: 10_000,
  });

  const filtered = useMemo(() => {
    if (!data) return [];
    const needle = q.trim().toLowerCase();
    if (!needle) return data.results;
    return data.results.filter(
      (f) =>
        f.name_it.toLowerCase().includes(needle) ||
        (f.category ?? "").toLowerCase().includes(needle),
    );
  }, [data, q]);

  return (
    <main className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">I miei alimenti</h1>
        <Link
          to="/alimento/nuovo"
          className="text-sm px-3 py-1.5 border border-[color:var(--color-border)] rounded-md hover:bg-[color:var(--color-brand-50)] no-underline"
        >
          + Nuovo
        </Link>
      </div>

      {isLoading ? (
        <p role="status" className="text-sm opacity-70">Caricamento…</p>
      ) : error ? (
        <p role="alert" className="text-sm text-red-700">Errore nel caricamento.</p>
      ) : !data || data.results.length === 0 ? (
        <div className="rounded-md border border-dashed border-[color:var(--color-border)] p-6 text-center text-sm text-[color:var(--color-ink-muted)]">
          <p className="mb-2">Non hai ancora aggiunto alimenti personali.</p>
          <Link to="/alimento/nuovo" className="underline">
            Aggiungine uno
          </Link>
        </div>
      ) : (
        <>
          <input
            type="search"
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="Cerca tra i tuoi alimenti…"
            aria-label="Cerca alimenti personali"
            className="w-full border border-[color:var(--color-border)] rounded-md px-3 py-2 text-sm"
          />
          {filtered.length === 0 ? (
            <p className="text-sm text-[color:var(--color-ink-muted)] py-4 text-center">
              Nessun alimento corrisponde a "{q}".
            </p>
          ) : null}
          <ul className="space-y-2">
            {filtered.map((food) => (
            <li
              key={food.id}
              className="border border-[color:var(--color-border)] rounded-[color:var(--radius-card)] bg-white overflow-hidden"
            >
              <ul className="list-none m-0 p-0">
                <ResultCard food={food} />
              </ul>
              <div className="border-t border-[color:var(--color-border)] px-3 py-2 flex justify-end">
                <Link
                  to={`/alimento/${food.id}/modifica`}
                  className="text-xs px-3 py-1.5 border border-[color:var(--color-border)] rounded-md hover:bg-[color:var(--color-brand-50)] no-underline"
                  aria-label={`Modifica ${food.name_it}`}
                >
                  Modifica
                </Link>
              </div>
            </li>
          ))}
          </ul>
        </>
      )}
    </main>
  );
}
