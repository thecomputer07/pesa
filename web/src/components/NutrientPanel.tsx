import { useTranslation } from "react-i18next";
import type { PortionValues } from "../lib/calc";
import { useFormatGrams, useFormatKcal } from "../lib/format";

export interface NutrientPanelProps {
  grams: number;
  portion: PortionValues;
  per100: PortionValues;
}

/** Two-column read-out: per-portion (prominent) + per-100g (contextual). */
export default function NutrientPanel({ grams, portion, per100 }: NutrientPanelProps) {
  const { t } = useTranslation();
  const fmtG = useFormatGrams();
  const fmtKcal = useFormatKcal();

  const rows: { key: keyof typeof portion.values; label: string; fmt: (n?: number) => string }[] = [
    { key: "carb_g", label: t("detail.carb"), fmt: fmtG },
    { key: "sugar_g", label: t("detail.sugar"), fmt: fmtG },
    { key: "fiber_g", label: t("detail.fiber"), fmt: fmtG },
    { key: "kcal", label: t("detail.kcal"), fmt: fmtKcal },
    { key: "protein_g", label: t("detail.protein"), fmt: fmtG },
    { key: "fat_g", label: t("detail.fat"), fmt: fmtG },
    { key: "water_g", label: t("detail.water"), fmt: fmtG },
    { key: "salt_g", label: t("detail.salt"), fmt: fmtG },
  ];

  return (
    <section
      aria-label={t("detail.perPortion", { grams })}
      className="bg-white border border-[color:var(--color-border)] rounded-[color:var(--radius-card)] overflow-hidden"
    >
      <header className="grid grid-cols-[1fr_auto_auto] gap-4 px-4 py-2 text-xs uppercase tracking-wide text-[color:var(--color-ink-muted)] bg-[color:var(--color-surface-muted)] border-b border-[color:var(--color-border)]">
        <span></span>
        <span className="text-right">{t("detail.perPortion", { grams })}</span>
        <span className="text-right">{t("detail.per100")}</span>
      </header>
      <dl className="divide-y divide-[color:var(--color-border)]">
        {rows.map(({ key, label, fmt }) => (
          <div key={key} className="grid grid-cols-[1fr_auto_auto] gap-4 px-4 py-2 items-baseline">
            <dt className="text-sm">{label}</dt>
            <dd
              className={`text-right ${key === "carb_g" ? "font-semibold text-base" : ""}`}
              aria-live={key === "carb_g" ? "polite" : undefined}
            >
              {fmt(portion.values[key])}
            </dd>
            <dd className="text-right text-sm text-[color:var(--color-ink-muted)]">
              {fmt(per100.values[key])}
            </dd>
          </div>
        ))}
      </dl>
    </section>
  );
}
