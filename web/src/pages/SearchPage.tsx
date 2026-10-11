import { useEffect, useState, useTransition } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import SearchBox from "../components/SearchBox";
import ResultList from "../components/ResultList";
import { searchFoods, type FoodSearchResult, type SearchSource } from "../lib/api";
import { localSearch } from "../lib/db/search";
import { upsertFoods } from "../lib/db/sync";
import { useAuth } from "../lib/auth/AuthContext";

const SOURCE_KEY = "pesa.search.source";

function readSource(v: string | null): SearchSource | "" {
  return v === "crea" || v === "ciqual" || v === "off" || v === "mine" ? v : "";
}

export default function SearchPage() {
  const { t } = useTranslation();
  const [params, setParams] = useSearchParams();
  const { user } = useAuth();
  const initialQ = params.get("q") ?? "";
  const [query, setQuery] = useState(initialQ);
  const [, startTransition] = useTransition();
  // Optional source filter ("" = all). Remembered per device; URL wins.
  const [source, setSource] = useState<SearchSource | "">(() => {
    const fromUrl = readSource(params.get("source"));
    if (fromUrl) return fromUrl;
    try {
      return readSource(localStorage.getItem(SOURCE_KEY));
    } catch {
      return "";
    }
  });

  const q = query.trim();
  const canSearch = q.length >= 2;

  const { data, isLoading, error, isError } = useQuery({
    queryKey: ["foods", "search", q, source],
    queryFn: async ({ signal }) => {
      const resp = await searchFoods(q, { limit: 20, page: 1, source: source || undefined, signal });
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

  const offlineRows = (fallback ?? []).filter((r) =>
    source === "" ? true : source === "mine" ? r.source.code === "manual" : r.source.code === source,
  );
  const results = data?.results ?? (fallback ? offlineRows : []);

  function urlParams(value: string, src: SearchSource | "") {
    const p: Record<string, string> = {};
    if (value) p.q = value;
    if (src) p.source = src;
    return p;
  }
  function chooseSource(src: SearchSource | "") {
    setSource(src);
    try {
      if (src) localStorage.setItem(SOURCE_KEY, src);
      else localStorage.removeItem(SOURCE_KEY);
    } catch {
      /* storage unavailable: filter still works for this session */
    }
    setParams(urlParams(query, src), { replace: true });
  }

  function handleSubmit(value: string) {
    startTransition(() => {
      setQuery(value);
      setParams(urlParams(value, source), { replace: true });
    });
  }
  function handleDebouncedChange(value: string) {
    startTransition(() => {
      setQuery(value);
      setParams(urlParams(value, source), { replace: true });
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

      <div role="radiogroup" aria-label={t("search.filterLabel")} className="flex flex-wrap items-center gap-2 text-sm">
        <span className="opacity-70">{t("search.filterLabel")}:</span>
        {(
          [
            ["", t("search.filterAll")],
            ["crea", t("search.filterCrea")],
            ["ciqual", t("search.filterCiqual")],
            ["off", t("search.filterOff")],
            ...(user ? [["mine", t("search.filterMine")]] : []),
          ] as [SearchSource | "", string][]
        ).map(([value, label]) => (
          <button
            key={value || "all"}
            type="button"
            role="radio"
            aria-checked={source === value}
            onClick={() => chooseSource(value)}
            className={[
              "px-3 py-1 rounded-full border",
              source === value
                ? "bg-[color:var(--color-brand-600)] text-white border-[color:var(--color-brand-600)]"
                : "bg-white border-[color:var(--color-border)]",
            ].join(" ")}
          >
            {label}
          </button>
        ))}
      </div>

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
