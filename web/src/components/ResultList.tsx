import { useTranslation } from "react-i18next";
import type { FoodSearchResult } from "../lib/api";
import ResultCard from "./ResultCard";
import Icon from "./ui/Icon";

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
      <div role="status" aria-live="polite">
        <span className="sr-only">{t("search.loading")}</span>
        <ul aria-hidden="true" className="surface divide-y divide-[color:var(--color-border)] overflow-hidden">
          {Array.from({ length: 5 }, (_, i) => (
            <li key={i} className="flex items-center gap-4 p-4 animate-pulse">
              <span className="w-14 h-14 rounded-2xl bg-[color:var(--color-surface-sunken)]" />
              <span className="flex-1 space-y-2">
                <span className="block h-3.5 w-2/5 rounded bg-[color:var(--color-surface-sunken)]" />
                <span className="block h-3 w-1/4 rounded bg-[color:var(--color-surface-sunken)]" />
              </span>
              <span className="h-5 w-14 rounded bg-[color:var(--color-surface-sunken)]" />
            </li>
          ))}
        </ul>
      </div>
    );
  }
  if (error) {
    return (
      <p role="alert" className="surface py-10 px-6 text-center text-sm text-red-700">
        {t("search.error")}
      </p>
    );
  }
  if (results.length === 0) {
    return (
      <div role="status" aria-live="polite" className="surface py-12 px-6 text-center">
        <span className="mx-auto grid place-items-center w-12 h-12 rounded-2xl bg-[color:var(--color-surface-sunken)] text-[color:var(--color-ink-faint)]">
          <Icon name="search" size={22} />
        </span>
        <p className="mt-4 text-sm text-[color:var(--color-ink-muted)]">{t("search.empty")}</p>
      </div>
    );
  }

  return (
    <>
      {offline ? (
        <p role="status" className="mb-3 flex items-center gap-2 text-xs text-[color:var(--color-ink-muted)]">
          <Icon name="wifiOff" size={14} /> {t("search.offlineNotice")}
        </p>
      ) : null}
      <p className="mb-3 text-xs font-medium text-[color:var(--color-ink-faint)] tabular" aria-live="polite">
        {t("search.resultsCount", { count: results.length })}
      </p>
      <ul className="surface divide-y divide-[color:var(--color-border)] overflow-hidden">
        {results.map((f) => (
          <ResultCard key={`${f.source.code}-${f.id}`} food={f} />
        ))}
      </ul>
    </>
  );
}
