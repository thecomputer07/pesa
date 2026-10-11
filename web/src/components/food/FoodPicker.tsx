import { useEffect, useId, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useQuery } from "@tanstack/react-query";
import { searchFoods, type FoodSearchResult } from "../../lib/api";

export interface FoodPickerPick {
  id: number;
  name_it: string;
  name_en?: string;
  kcal?: number;
  protein_g?: number;
  fat_g?: number;
  carb_g?: number;
  sugar_g?: number;
  fiber_g?: number;
}

export interface FoodPickerProps {
  onPick: (food: FoodPickerPick) => void;
  placeholder?: string;
  /** Minimum characters before searching. Default 2. */
  minChars?: number;
  /** Debounce in ms. Default 250. */
  debounceMs?: number;
  /** Maximum results. Default 10. */
  limit?: number;
  /** Optional aria-label for the combobox input. */
  ariaLabel?: string;
}

/**
 * Accessible food autocomplete. Follows the WAI-ARIA 1.2 combobox pattern:
 *   - <input role="combobox" aria-expanded aria-controls aria-activedescendant>
 *   - <ul role="listbox" id=…>
 *   - each result <li role="option" id=…>
 *
 * Keyboard: ArrowDown/Up to move, Enter to pick, Escape to close.
 */
export default function FoodPicker({
  onPick,
  placeholder,
  minChars = 2,
  debounceMs = 250,
  limit = 10,
  ariaLabel,
}: FoodPickerProps) {
  const { t } = useTranslation();
  const baseId = useId();
  const listboxId = `${baseId}-listbox`;
  const inputRef = useRef<HTMLInputElement | null>(null);

  const [query, setQuery] = useState("");
  const [debounced, setDebounced] = useState("");
  const [open, setOpen] = useState(false);
  const [active, setActive] = useState(-1);

  useEffect(() => {
    const id = window.setTimeout(() => setDebounced(query.trim()), debounceMs);
    return () => window.clearTimeout(id);
  }, [query, debounceMs]);

  const enabled = debounced.length >= minChars;
  const { data, isFetching, isError } = useQuery({
    queryKey: ["foods-search", debounced, limit],
    queryFn: ({ signal }) => searchFoods(debounced, { limit, signal }),
    enabled,
    staleTime: 30_000,
  });

  const results: FoodSearchResult[] = useMemo(() => data?.results ?? [], [data]);

  // Keep `active` in range when results change.
  useEffect(() => {
    if (results.length === 0) setActive(-1);
    else if (active >= results.length) setActive(results.length - 1);
  }, [results, active]);

  function commit(food: FoodSearchResult) {
    onPick({
      id: food.id,
      name_it: food.name_it,
      name_en: food.name_en,
      kcal: food.kcal,
      protein_g: food.protein_g,
      fat_g: food.fat_g,
      carb_g: food.carb_g,
      sugar_g: food.sugar_g,
      fiber_g: food.fiber_g,
    });
    setQuery("");
    setDebounced("");
    setOpen(false);
    setActive(-1);
  }

  function onKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setOpen(true);
      setActive((a) => (results.length ? Math.min(results.length - 1, a + 1) : -1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setActive((a) => Math.max(0, a - 1));
    } else if (e.key === "Enter") {
      if (open && active >= 0 && results[active]) {
        e.preventDefault();
        commit(results[active]);
      }
    } else if (e.key === "Escape") {
      setOpen(false);
      setActive(-1);
    }
  }

  const showDropdown =
    open && enabled && (isFetching || isError || results.length >= 0);

  const activeId = active >= 0 && results[active] ? `${baseId}-opt-${active}` : undefined;

  return (
    <div className="relative">
      <input
        ref={inputRef}
        type="search"
        role="combobox"
        aria-label={ariaLabel ?? t("foodPicker.ariaLabel")}
        aria-expanded={showDropdown}
        aria-controls={listboxId}
        aria-autocomplete="list"
        aria-activedescendant={activeId}
        autoComplete="off"
        value={query}
        placeholder={placeholder ?? t("foodPicker.placeholder")}
        onChange={(e) => {
          setQuery(e.target.value);
          setOpen(true);
        }}
        onFocus={() => setOpen(true)}
        onBlur={() => {
          // Delay so click on option can register.
          window.setTimeout(() => setOpen(false), 150);
        }}
        onKeyDown={onKeyDown}
        className="w-full border border-[color:var(--color-border)] bg-white px-3.5 py-2.5 text-sm"
      />
      {showDropdown ? (
        <ul
          id={listboxId}
          role="listbox"
          className="absolute z-20 left-0 right-0 mt-1 max-h-64 overflow-y-auto rounded-xl border border-[color:var(--color-border)] bg-white shadow-lg"
        >
          {isFetching && results.length === 0 ? (
            <li className="px-3 py-2 text-sm text-[color:var(--color-ink-muted)]" aria-live="polite">
              {t("search.loading")}
            </li>
          ) : null}
          {isError ? (
            <li role="alert" className="px-3 py-2 text-sm text-red-700">
              {t("search.error")}
            </li>
          ) : null}
          {!isFetching && !isError && results.length === 0 && enabled ? (
            <li className="px-3 py-2 text-sm text-[color:var(--color-ink-muted)]">
              {t("search.empty")}
            </li>
          ) : null}
          {results.map((food, idx) => {
            const optId = `${baseId}-opt-${idx}`;
            const isActive = idx === active;
            return (
              <li
                id={optId}
                key={food.id}
                role="option"
                aria-selected={isActive}
                onMouseDown={(e) => {
                  // Use mousedown so we commit before input blurs.
                  e.preventDefault();
                  commit(food);
                }}
                onMouseEnter={() => setActive(idx)}
                className={[
                  "px-3 py-2 text-sm cursor-pointer flex items-center justify-between gap-2",
                  isActive ? "bg-[color:var(--color-brand-50)]" : "",
                ].join(" ")}
              >
                <span className="flex items-center gap-2 min-w-0">
                  {food.thumb_url ? (
                    <img
                      src={food.thumb_url}
                      alt=""
                      loading="lazy"
                      onError={(e) => {
                        (e.currentTarget as HTMLImageElement).style.display = "none";
                      }}
                      className="w-8 h-8 object-cover rounded border border-[color:var(--color-border)] bg-white shrink-0"
                    />
                  ) : null}
                  <span className="truncate">{food.name_it}</span>
                </span>
                {typeof food.kcal === "number" ? (
                  <span className="shrink-0 text-xs text-[color:var(--color-ink-muted)]">
                    {Math.round(food.kcal)} kcal/100g
                  </span>
                ) : null}
              </li>
            );
          })}
        </ul>
      ) : null}
    </div>
  );
}
