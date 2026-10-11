import { useTranslation } from "react-i18next";
import type { FoodDetail } from "../lib/api";
import SourceBadge from "./SourceBadge";
import { useFormatGrams } from "../lib/format";

export interface ConflictPanelProps {
  candidates: FoodDetail[];
  selectedId: number;
  onPick: (id: number) => void;
}

/** Side-by-side comparison when CREA vs BDA disagree by >10% on carbs. */
export default function ConflictPanel({ candidates, selectedId, onPick }: ConflictPanelProps) {
  const { t } = useTranslation();
  const fmt = useFormatGrams();

  return (
    <section
      role="region"
      aria-labelledby="conflict-title"
      className="bg-amber-50 border border-amber-200 rounded-[color:var(--radius-card)] p-4"
    >
      <h2 id="conflict-title" className="font-semibold text-amber-900">
        {t("detail.conflictTitle")}
      </h2>
      <p className="mt-1 text-sm text-amber-900/80">{t("detail.conflictBody")}</p>

      <div className="mt-3 grid gap-3 md:grid-cols-2">
        {candidates.map((c) => {
          const selected = c.id === selectedId;
          return (
            <div
              key={`${c.attribution}-${c.id}`}
              className={[
                "bg-white rounded-xl border p-3 flex flex-col gap-2",
                selected ? "border-[color:var(--color-brand-600)]" : "border-[color:var(--color-border)]",
              ].join(" ")}
            >
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">{c.name_it}</span>
                <SourceBadge
                  code={sourceCodeOf(c)}
                  license={c.license}
                  licenseUrl={c.license_url}
                />
              </div>
              <p className="text-sm">
                <span className="opacity-70">{t("detail.carb")}:</span>{" "}
                <span className="font-semibold">{fmt(c.carb_g)}</span>
              </p>
              <button
                type="button"
                onClick={() => onPick(c.id)}
                aria-pressed={selected}
                className={[
                  "mt-auto px-3 py-1.5 rounded text-sm",
                  selected
                    ? "bg-[color:var(--color-brand-600)] text-white"
                    : "border border-[color:var(--color-border)]",
                ].join(" ")}
              >
                {t("detail.conflictPick")}
              </button>
            </div>
          );
        })}
      </div>
    </section>
  );
}

function sourceCodeOf(f: FoodDetail): string {
  const a = f.attribution.toUpperCase();
  if (a.includes("CREA")) return "CREA";
  if (a.includes("CIQUAL")) return "CIQUAL";
  if (a.includes("AGRICULTURE") || a.includes("FOODDATA")) return "USDA";
  if (a.includes("BDA") || a.includes("IEO")) return "BDA";
  if (a.includes("OPEN FOOD")) return "OFF";
  return "—";
}
