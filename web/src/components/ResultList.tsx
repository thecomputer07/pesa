import { useTranslation } from "react-i18next";
import type { FoodSearchResult } from "../lib/api";
import ResultCard from "./ResultCard";

export interface ResultListProps {
  results: FoodSearchResult[];
  loading?: boolean;
  error?: Error | null;
  offline?: boolean;
}

export default function ResultList({ results, loading, error, offline }: ResultListProps) {
  const { t } = useTranslation();

  if (loading) {
    return (
      <p role="status" aria-live="polite" className="py-8 text-center text-sm opacity-70">
        {t("search.loading")}
      </p>
    );
  }
  if (error) {
    return (
      <p role="alert" className="py-8 text-center text-sm text-red-700">
        {t("search.error")}
      </p>
    );
  }
  if (results.length === 0) {
    return (
      <p role="status" aria-live="polite" className="py-8 text-center text-sm opacity-70">
        {t("search.empty")}
      </p>
    );
  }

  return (
    <>
      {offline ? (
        <p role="status" className="text-xs mb-3 opacity-80">
          {t("search.offlineNotice")}
        </p>
      ) : null}
      <p className="text-xs mb-2 text-[color:var(--color-ink-muted)]" aria-live="polite">
        {t("search.resultsCount", { count: results.length })}
      </p>
      <ul className="space-y-2">
        {results.map((f) => (
          <ResultCard key={`${f.source.code}-${f.id}`} food={f} />
        ))}
      </ul>
    </>
  );
}
