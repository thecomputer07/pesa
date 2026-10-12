import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useMealsForDay } from "../../lib/db/queries";
import type { LocalMeal, LocalMealItem, Slot } from "../../lib/sync/types";
import { useFormatGrams } from "../../lib/format";

const SLOT_ORDER: Slot[] = ["colazione", "pranzo", "cena", "spuntino", "altro"];

export interface DayTimelineProps {
  isoDate: string;
  onEdit: (meal: LocalMeal) => void;
}

export default function DayTimeline({ isoDate, onEdit }: DayTimelineProps) {
  const { t } = useTranslation();
  const data = useMealsForDay(isoDate);
  const grouped = useMemo(() => groupBySlot(data?.meals ?? [], data?.items ?? []), [data]);
  const fmtG = useFormatGrams();

  if (!data) {
    return (
      <p role="status" className="py-6 text-sm opacity-70">
        {t("a11y.loading")}
      </p>
    );
  }
  if (data.meals.length === 0) {
    return (
      <p className="py-6 text-sm text-[color:var(--color-ink-muted)]">{t("diary.empty")}</p>
    );
  }

  return (
    <div className="space-y-4">
      {SLOT_ORDER.map((slot) => {
        const section = grouped[slot];
        if (!section || section.length === 0) return null;
        return (
          <section key={slot} aria-labelledby={`slot-${slot}`}>
            <h2 id={`slot-${slot}`} className="text-sm font-semibold mb-2">
              {t(`diary.slots.${slot}`)}
            </h2>
            <ul className="space-y-2">
              {section.map(({ meal, items }) => (
                <li
                  key={meal.client_id}
                  className="border border-[color:var(--color-border)] rounded-xl bg-white"
                >
                  <button
                    type="button"
                    onClick={() => onEdit(meal)}
                    className="block w-full text-left p-3"
                  >
                    <div className="flex justify-between text-xs text-[color:var(--color-ink-muted)]">
                      <span>{formatRomeTime(meal.occurred_at)}</span>
                      <span>
                        {t("detail.carb")}:{" "}
                        {fmtG(items.reduce((s, i) => s + (i.carb_g ?? 0) * (i.grams / 100), 0))}
                      </span>
                    </div>
                    <p className="text-sm truncate">
                      {items.slice(0, 2).map((i) => i.name_snapshot).join(", ") ||
                        t("diary.empty")}
                      {items.length > 2 ? ` +${items.length - 2}` : ""}
                    </p>
                  </button>
                </li>
              ))}
            </ul>
          </section>
        );
      })}
    </div>
  );
}

function groupBySlot(meals: LocalMeal[], items: LocalMealItem[]) {
  const out: Partial<Record<Slot, Array<{ meal: LocalMeal; items: LocalMealItem[] }>>> = {};
  const byMeal = new Map<string, LocalMealItem[]>();
  for (const i of items) {
    const arr = byMeal.get(i.meal_client_id) ?? [];
    arr.push(i);
    byMeal.set(i.meal_client_id, arr);
  }
  const sorted = [...meals].sort((a, b) => a.occurred_at - b.occurred_at);
  for (const m of sorted) {
    const arr = out[m.slot] ?? [];
    arr.push({ meal: m, items: byMeal.get(m.client_id) ?? [] });
    out[m.slot] = arr;
  }
  return out;
}

function formatRomeTime(ms: number): string {
  return new Intl.DateTimeFormat("it-IT", {
    timeZone: "Europe/Rome",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(ms));
}
