import { useEffect, useState, useTransition } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import SearchBox from "../components/SearchBox";
import ResultList from "../components/ResultList";
import { searchFoods, type FoodSearchResult } from "../lib/api";
import { localSearch } from "../lib/db/search";
import { upsertFoods } from "../lib/db/sync";
import { useAuth } from "../lib/auth/AuthContext";

export default function SearchPage() {
  const { t } = useTranslation();
  const [params, setParams] = useSearchParams();
  const { user } = useAuth();
  const initialQ = params.get("q") ?? "";
  const [query, setQuery] = useState(initialQ);
  const [, startTransition] = useTransition();

  const q = query.trim();
  const canSearch = q.length >= 2;

  const { data, isLoading, error, isError } = useQuery({
    queryKey: ["foods", "search", q],
    queryFn: async ({ signal }) => {
      const resp = await searchFoods(q, { limit: 20, page: 1, signal });
      // Warm local cache on the side (fire-and-forget).
      void upsertFoods(resp.results);
      return resp;
    },
    enabled: canSearch,
    staleTime: 60_000,
  });

  // Fallback to offline cache when network fails.
  const [fallback, setFallback] = useState<FoodSearchResult[] | null>(null);
  useEffect(() => {
    if (!canSearch || !isError) {
      setFallback(null);
      return;
    }
    let alive = true;
    void (async () => {
      const rows = await localSearch(q);
      if (!alive) return;
      setFallback(
        rows.map((r) => ({
          id: r.id,
          name_it: r.name_it,
          name_en: r.name_en,
          category: r.category,
          state: r.state,
          kcal: r.kcal,
          carb_g: r.carb_g,
          protein_g: r.protein_g,
          fat_g: r.fat_g,
          sugar_g: r.sugar_g,
          fiber_g: r.fiber_g,
          source: {
            code: r.source_code,
            license: r.source_license,
            license_url: r.source_license_url,
          },
        })),
      );
    })();
    return () => {
      alive = false;
    };
  }, [q, canSearch, isError]);

  const results = data?.results ?? fallback ?? [];

  function handleSubmit(value: string) {
    startTransition(() => {
      setQuery(value);
      setParams(value ? { q: value } : {}, { replace: true });
    });
  }
  function handleDebouncedChange(value: string) {
    startTransition(() => {
      setQuery(value);
      setParams(value ? { q: value } : {}, { replace: true });
    });
  }

  const showAddCTA = user && canSearch && !isLoading && results.length === 0 && !isError;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-3">
        <h1 className="text-2xl font-bold">{t("nav.search")}</h1>
        {user ? (
          <Link
            to="/alimento/nuovo"
            className="text-sm px-3 py-1.5 border border-[color:var(--color-border)] rounded-md hover:bg-[color:var(--color-brand-50)] no-underline"
          >
            + Aggiungi alimento
          </Link>
        ) : null}
      </div>
      <SearchBox
        defaultValue={initialQ}
        onSubmit={handleSubmit}
        onQueryChange={handleDebouncedChange}
        autoFocus
      />

      {!canSearch ? (
        <p className="text-sm opacity-70">{t("search.minChars", { n: 2 })}</p>
      ) : (
        <>
          <ResultList
            results={results}
            loading={isLoading}
            error={isError && !fallback?.length ? (error as Error) : null}
            offline={isError && (fallback?.length ?? 0) > 0}
          />
          {showAddCTA ? (
            <p className="text-sm text-center mt-4">
              Non trovi quello che cerchi?{" "}
              <Link to="/alimento/nuovo" className="underline">
                Aggiungi un tuo alimento
              </Link>
            </p>
          ) : null}
        </>
      )}
    </div>
  );
}
