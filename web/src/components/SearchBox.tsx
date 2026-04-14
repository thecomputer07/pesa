import { forwardRef, useEffect, useId, useRef, useState } from "react";
import { useTranslation } from "react-i18next";

export interface SearchBoxProps {
  defaultValue?: string;
  onSubmit: (query: string) => void;
  /** Debounced onChange for live search. */
  onQueryChange?: (query: string) => void;
  debounceMs?: number;
  autoFocus?: boolean;
}

const SearchBox = forwardRef<HTMLInputElement, SearchBoxProps>(function SearchBox(
  { defaultValue = "", onSubmit, onQueryChange, debounceMs = 250, autoFocus = false },
  ref,
) {
  const { t } = useTranslation();
  const id = useId();
  const [q, setQ] = useState(defaultValue);
  const timerRef = useRef<number | null>(null);

  useEffect(() => {
    if (!onQueryChange) return;
    if (timerRef.current) window.clearTimeout(timerRef.current);
    timerRef.current = window.setTimeout(() => onQueryChange(q.trim()), debounceMs);
    return () => {
      if (timerRef.current) window.clearTimeout(timerRef.current);
    };
  }, [q, debounceMs, onQueryChange]);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    onSubmit(q.trim());
  }

  return (
    <form role="search" onSubmit={handleSubmit} className="flex gap-2">
      <label htmlFor={id} className="sr-only">
        {t("search.label")}
      </label>
      <input
        id={id}
        ref={ref}
        type="search"
        value={q}
        onChange={(e) => setQ(e.target.value)}
        placeholder={t("search.placeholder")}
        autoComplete="off"
        autoFocus={autoFocus}
        aria-describedby={`${id}-hint`}
        className="flex-1 px-3 py-2 rounded-md border border-[color:var(--color-border)] bg-white"
      />
      <button
        type="submit"
        className="px-4 py-2 rounded-md bg-[color:var(--color-brand-600)] hover:bg-[color:var(--color-brand-700)] text-white font-medium"
      >
        {t("search.submit")}
      </button>
      <span id={`${id}-hint`} className="sr-only">
        {t("search.minChars", { n: 2 })}
      </span>
    </form>
  );
});

export default SearchBox;
