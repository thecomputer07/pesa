import { useTranslation } from "react-i18next";
import type { PortionValues } from "../lib/calc";
import { useFormatGrams, useFormatKcal } from "../lib/format";

export interface NutrientPanelProps {
  grams: number;
  portion: PortionValues;
  per100: PortionValues;
}

/** Full read-out: per-portion (prominent) + per-100 g (contextual). */
export default function NutrientPanel({ grams, portion, per100 }: NutrientPanelProps) {
  const { t } = useTranslation();
  const fmtG = useFormatGrams();
  const fmtKcal = useFormatKcal();

  const rows: { key: keyof typeof portion.values; label: string; fmt: (n?: number) => string; sub?: boolean }[] = [
    { key: "carb_g", label: t("detail.carb"), fmt: fmtG },
    { key: "sugar_g", label: t("detail.sugar"), fmt: fmtG, sub: true },
    { key: "fiber_g", label: t("detail.fiber"), fmt: fmtG },
    { key: "kcal", label: t("detail.kcal"), fmt: fmtKcal },
    { key: "protein_g", label: t("detail.protein"), fmt: fmtG },
    { key: "fat_g", label: t("detail.fat"), fmt: fmtG },
    { key: "water_g", label: t("detail.water"), fmt: fmtG },
    { key: "salt_g", label: t("detail.salt"), fmt: fmtG },
  ];

  return (
    <section aria-label={t("detail.perPortion", { grams })} className="surface overflow-hidden">
      <header className="grid grid-cols-[1fr_6.5rem_5.5rem] sm:grid-cols-[1fr_8rem_7rem] items-end gap-2 px-5 pt-5 pb-3">
        <h2 className="text-base font-semibold">{t("detail.nutrients")}</h2>
        <span className="text-right text-xs font-semibold text-[color:var(--color-ink)] tabular">
          <span className="sr-only">{t("detail.perPortion", { grams })}</span>
          <span aria-hidden="true">{t("detail.portionCol", { grams })}</span>
        </span>
        <span className="text-right text-xs font-medium text-[color:var(--color-ink-faint)] tabular">
          <span className="sr-only">{t("detail.per100")}</span>
          <span aria-hidden="true">{t("detail.per100Col")}</span>
        </span>
      </header>
      <dl className="px-2 pb-2">
        {rows.map(({ key, label, fmt, sub }) => {
          const carb = key === "carb_g";
          return (
            <div
              key={key}
              className={[
                "grid grid-cols-[1fr_6.5rem_5.5rem] sm:grid-cols-[1fr_8rem_7rem] items-baseline gap-2 px-3 py-2.5 rounded-xl",
                carb ? "bg-[color:var(--color-carb-50)]" : "border-t border-[color:var(--color-border)] first:border-0",
              ].join(" ")}
            >
              <dt className={`text-sm ${sub ? "pl-4 text-[color:var(--color-ink-muted)]" : ""} ${carb ? "font-semibold" : ""}`}>
                {label}
              </dt>
              <dd
                className={`text-right tabular ${carb ? "font-semibold text-[color:var(--color-carb-700)]" : "font-medium"}`}
                aria-live={carb ? "polite" : undefined}
              >
                {fmt(portion.values[key])}
              </dd>
              <dd className="text-right text-sm tabular text-[color:var(--color-ink-faint)]">{fmt(per100.values[key])}</dd>
            </div>
          );
        })}
      </dl>
    </section>
  );
}
