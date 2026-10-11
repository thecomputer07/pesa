import { forwardRef, useEffect, useId, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import Icon from "./ui/Icon";

export interface SearchBoxProps {
  defaultValue?: string;
  onSubmit: (query: string) => void;
  /** Debounced onChange for live search. */
  onQueryChange?: (query: string) => void;
  debounceMs?: number;
  autoFocus?: boolean;
  /** "lg" for the home hero. */
  size?: "md" | "lg";
}

const SearchBox = forwardRef<HTMLInputElement, SearchBoxProps>(function SearchBox(
  { defaultValue = "", onSubmit, onQueryChange, debounceMs = 250, autoFocus = false, size = "md" },
  ref,
) {
  const { t } = useTranslation();
  const id = useId();
  const [q, setQ] = useState(defaultValue);
  const timerRef = useRef<number | null>(null);
  // Latest callback in a ref: a parent re-render must not restart the timer
  // (it used to keep postponing it until after the user had tapped a result,
  // and the late URL update then undid that navigation).
  const onChangeRef = useRef(onQueryChange);
  onChangeRef.current = onQueryChange;
  // Only typing triggers a live search, not mounting with a default value.
  const lastEmitted = useRef(defaultValue.trim());

  useEffect(() => {
    const value = q.trim();
    if (!onChangeRef.current || value === lastEmitted.current) return;
    timerRef.current = window.setTimeout(() => {
      lastEmitted.current = value;
      onChangeRef.current?.(value);
    }, debounceMs);
    return () => {
      if (timerRef.current) window.clearTimeout(timerRef.current);
    };
  }, [q, debounceMs]);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (timerRef.current) window.clearTimeout(timerRef.current);
    lastEmitted.current = q.trim();
    onSubmit(q.trim());
  }

  const lg = size === "lg";

  return (
    <form
      role="search"
      onSubmit={handleSubmit}
      className={[
        "group relative flex items-center bg-white transition-shadow",
        "shadow-[0_0_0_1px_var(--color-border),var(--shadow-soft)]",
        "focus-within:shadow-[0_0_0_1px_var(--color-brand-500),0_0_0_5px_rgb(49_87_224/0.14),var(--shadow-lift)]",
        lg ? "rounded-[1.375rem] p-2" : "rounded-[1.125rem] p-1.5",
      ].join(" ")}
    >
      <label htmlFor={id} className="sr-only">
        {t("search.label")}
      </label>
      <Icon
        name="search"
        size={lg ? 22 : 20}
        className={`shrink-0 text-[color:var(--color-ink-faint)] group-focus-within:text-[color:var(--color-brand-600)] ${lg ? "ml-3" : "ml-2.5"}`}
      />
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
        className={[
          "flex-1 min-w-0 bg-transparent border-0 !shadow-none focus:!shadow-none outline-none placeholder:text-[color:var(--color-ink-faint)]",
          lg ? "px-3 py-3 text-lg" : "px-2.5 py-2 text-base",
        ].join(" ")}
      />
      <button
        type="submit"
        className={`btn-primary shrink-0 ${lg ? "!px-6 !py-3 !rounded-2xl text-base" : "!px-4 !py-2 !rounded-[0.875rem] text-sm"}`}
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
