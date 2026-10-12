import { useLiveQuery } from "dexie-react-hooks";
import { db } from "./schema";
import { computeDayTotals, type Nutrients } from "../diary/totals";
import type { LocalMeal, LocalMealItem } from "../sync/types";

/**
 * ISO "YYYY-MM-DD" (Europe/Rome local) → [startMs, endMs) in UTC ms.
 * Uses a two-step Intl.DateTimeFormat probe so DST transitions resolve
 * correctly (no hard-coded +01:00 offset).
 */
export function romeDayBounds(isoDate: string): [number, number] {
  const [y, m, d] = isoDate.split("-").map(Number);
  // Construct a UTC midnight at the desired calendar day, then measure Rome's
  // offset at that instant and subtract it to land on Rome local midnight.
  const probe = new Date(Date.UTC(y, m - 1, d, 12, 0, 0));
  const offsetMin = romeOffsetMinutes(probe);
  const startMs = Date.UTC(y, m - 1, d, 0, 0, 0) - offsetMin * 60_000;
  return [startMs, startMs + 86_400_000];
}

/** Minutes Europe/Rome is ahead of UTC at the given instant (+60 or +120). */
export function romeOffsetMinutes(at: Date): number {
  const fmt = new Intl.DateTimeFormat("en-US", {
    timeZone: "Europe/Rome",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  });
  const parts = fmt.formatToParts(at);
  const map: Record<string, string> = {};
  for (const p of parts) map[p.type] = p.value;
  const local = Date.UTC(
    Number(map.year),
    Number(map.month) - 1,
    Number(map.day),
    Number(map.hour),
    Number(map.minute),
  );
  return Math.round((local - at.getTime()) / 60_000);
}

/** Today in Europe/Rome as ISO "YYYY-MM-DD". */
export function todayRomeIso(): string {
  return new Intl.DateTimeFormat("en-CA", { timeZone: "Europe/Rome" }).format(new Date());
}

/** Convert an ms-epoch timestamp to a MariaDB TIMESTAMP string in UTC. */
export function toSqlTimestamp(ms: number): string {
  const d = new Date(ms);
  const pad = (n: number) => String(n).padStart(2, "0");
  return (
    `${d.getUTCFullYear()}-${pad(d.getUTCMonth() + 1)}-${pad(d.getUTCDate())}T` +
    `${pad(d.getUTCHours())}:${pad(d.getUTCMinutes())}:${pad(d.getUTCSeconds())}Z`
  );
}

export interface MealsForDay {
  meals: LocalMeal[];
  items: LocalMealItem[];
}

/** Live query: meals + their items for the given ISO date (Europe/Rome). */
export function useMealsForDay(isoDate: string): MealsForDay | undefined {
  return useLiveQuery(async () => {
    const [from, to] = romeDayBounds(isoDate);
    const meals = await db.meals
      .where("occurred_at")
      .between(from, to, true, false)
      .and((m) => !m.deleted_at_ms)
      .toArray();
    if (meals.length === 0) return { meals: [], items: [] };
    const mealCids = meals.map((m) => m.client_id);
    const items = await db.meal_items
      .where("meal_client_id")
      .anyOf(mealCids)
      .and((i) => !i.deleted_at_ms)
      .toArray();
    return { meals, items };
  }, [isoDate]);
}

/** Aggregated Nutrients for the given day (joined with foods cache). */
export function useDayTotals(isoDate: string): Nutrients | undefined {
  const data = useMealsForDay(isoDate);
  const foodRefs = useLiveQuery(async () => {
    if (!data) return undefined;
    const ids = Array.from(
      new Set(
        data.items
          .map((i) => i.food_id)
          .filter((v): v is number => typeof v === "number"),
      ),
    );
    if (ids.length === 0) return new Map<number, Partial<Nutrients>>();
    const foods = await db.foods.bulkGet(ids);
    const map = new Map<number, Partial<Nutrients>>();
    for (const f of foods) {
      if (f) {
        map.set(f.id, {
          kcal: f.kcal,
          carb_g: f.carb_g,
          protein_g: f.protein_g,
          fat_g: f.fat_g,
          sugar_g: f.sugar_g,
          fiber_g: f.fiber_g,
        });
      }
    }
    return map;
  }, [data]);

  if (!data || !foodRefs) return undefined;

  return computeDayTotals(
    data.items.map((i) => ({
      grams: i.grams,
      // Item row already carries snapshotted nutrients for recipe items.
      snapshot:
        i.kcal != null || i.carb_g != null
          ? {
              kcal: i.kcal ?? undefined,
              carb_g: i.carb_g ?? undefined,
              protein_g: i.protein_g ?? undefined,
              fat_g: i.fat_g ?? undefined,
              sugar_g: i.sugar_g ?? undefined,
              fiber_g: i.fiber_g ?? undefined,
            }
          : undefined,
      food: i.food_id != null ? foodRefs.get(i.food_id) : undefined,
      deleted_at_ms: i.deleted_at_ms,
    })),
  );
}

/**
 * 7-day totals starting at `weekStartIso` (Monday). Returns an array of
 * `{isoDate, totals}` objects. Each day runs its own liveQuery via
 * `useMealsForDay` — kept intentionally simple for MVP; a single composite
 * query is a TODO optimization.
 */
export function useWeekTotals(weekStartIso: string): Array<{ isoDate: string; totals: Nutrients | undefined }> {
  const days: string[] = [];
  const [y, m, d] = weekStartIso.split("-").map(Number);
  const start = Date.UTC(y, m - 1, d);
  for (let i = 0; i < 7; i++) {
    const ms = start + i * 86_400_000;
    const dt = new Date(ms);
    const pad = (n: number) => String(n).padStart(2, "0");
    days.push(`${dt.getUTCFullYear()}-${pad(dt.getUTCMonth() + 1)}-${pad(dt.getUTCDate())}`);
  }
  // eslint-disable-next-line react-hooks/rules-of-hooks
  return days.map((iso) => ({ isoDate: iso, totals: useDayTotals(iso) }));
}
