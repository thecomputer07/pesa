import { useCallback } from "react";
import { useTranslation } from "react-i18next";

function localeFor(lang: string): string {
  return lang.startsWith("en") ? "en-GB" : "it-IT";
}

/** Formatter for grams, e.g. "120,5 g" (IT) or "120.5 g" (EN). */
export function useFormatGrams() {
  const { i18n } = useTranslation();
  const loc = localeFor(i18n.language);
  return useCallback(
    (value: number | null | undefined, digits = 1): string => {
      if (value == null || !Number.isFinite(value)) return "—";
      const nf = new Intl.NumberFormat(loc, {
        maximumFractionDigits: digits,
        minimumFractionDigits: 0,
      });
      return `${nf.format(value)} g`;
    },
    [loc],
  );
}

/** Formatter for kcal, integer. */
export function useFormatKcal() {
  const { i18n } = useTranslation();
  const loc = localeFor(i18n.language);
  return useCallback(
    (value: number | null | undefined): string => {
      if (value == null || !Number.isFinite(value)) return "—";
      return `${new Intl.NumberFormat(loc, { maximumFractionDigits: 0 }).format(value)} kcal`;
    },
    [loc],
  );
}

/** Formatter for a plain number (e.g. percentages). */
export function useFormatNumber() {
  const { i18n } = useTranslation();
  const loc = localeFor(i18n.language);
  return useCallback(
    (value: number | null | undefined, digits = 1): string => {
      if (value == null || !Number.isFinite(value)) return "—";
      return new Intl.NumberFormat(loc, {
        maximumFractionDigits: digits,
        minimumFractionDigits: 0,
      }).format(value);
    },
    [loc],
  );
}

/**
 * Format ISO date as dd/mm/yyyy (IT) or dd/mm/yyyy (EN — UK style for consistency).
 */
export function formatDate(iso: string | null | undefined, lang = "it"): string {
  if (!iso) return "—";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "—";
  return new Intl.DateTimeFormat(localeFor(lang), {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(d);
}

/** Parse an Italian-style decimal ("120,5") or standard ("120.5"). */
export function parseDecimal(raw: string): number | null {
  const s = raw.trim().replace(/\s+/g, "").replace(",", ".");
  if (s === "" || !/^\d+(\.\d+)?$/.test(s)) return null;
  const n = Number(s);
  return Number.isFinite(n) ? n : null;
}
