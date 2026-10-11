import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useLiveQuery } from "dexie-react-hooks";
import { db } from "../../lib/db/schema";
import { recipePer100g } from "../../lib/recipes/aggregate";
import type { Nutrients } from "../../lib/diary/totals";

export interface RecipePick {
  client_id: string;
  name: string;
  per100: Partial<Nutrients>;
}

interface RecipePickerProps {
  onPick: (recipe: RecipePick) => void;
}

/**
 * Combobox per scegliere una propria ricetta come ingrediente di un pasto.
 * Calcola le macros per 100g (snapshot della ricetta) usando aggregateRecipe.
 */
export default function RecipePicker({ onPick }: RecipePickerProps) {
  const { t } = useTranslation();
  const [q, setQ] = useState("");

  const recipes = useLiveQuery(async () => {
    const all = await db.recipes.toArray();
    return all.filter((r) => !r.deleted_at_ms);
  }, []);

  const filtered = useMemo(() => {
    if (!recipes) return [];
    const qq = q.trim().toLowerCase();
    if (!qq) return recipes;
    return recipes.filter((r) => r.name.toLowerCase().includes(qq));
  }, [recipes, q]);

  async function handlePick(clientId: string, name: string) {
    const items = await db.recipe_items.where({ recipe_client_id: clientId }).toArray();
    const active = items.filter((i) => !i.deleted_at_ms);
    const recipe = recipes?.find((r) => r.client_id === clientId);
    const totalGrams = recipe?.total_grams ?? undefined;
    const per100 = recipePer100g(
      active.map((i) => ({
        grams: i.grams,
        snapshot: (i.snapshot_json ?? {}) as Partial<Nutrients>,
      })),
      totalGrams ?? undefined,
    );
    onPick({ client_id: clientId, name, per100 });
  }

  if (!recipes) {
    return <p className="text-sm opacity-70">{t("a11y.loading")}</p>;
  }
  if (recipes.length === 0) {
    return (
      <p className="text-sm text-[color:var(--color-ink-muted)]">
        Nessuna ricetta. Creane una dalla pagina <em>Ricette</em>.
      </p>
    );
  }

  return (
    <div className="space-y-2">
      <input
        type="search"
        value={q}
        onChange={(e) => setQ(e.target.value)}
        placeholder="Cerca tra le tue ricette…"
        className="w-full border border-[color:var(--color-border)] bg-white px-3.5 py-2.5 text-sm"
      />
      <ul className="max-h-48 overflow-y-auto rounded-xl border border-[color:var(--color-border)] bg-white">
        {filtered.length === 0 ? (
          <li className="px-3 py-2 text-sm text-[color:var(--color-ink-muted)]">
            Nessun risultato.
          </li>
        ) : (
          filtered.map((r) => (
            <li key={r.client_id}>
              <button
                type="button"
                onClick={() => handlePick(r.client_id, r.name)}
                className="w-full text-left px-3 py-2 text-sm hover:bg-[color:var(--color-brand-50)]"
              >
                <span className="truncate block">{r.name}</span>
                <span className="text-xs text-[color:var(--color-ink-muted)]">
                  {r.servings} {Number(r.servings) === 1 ? "porzione" : "porzioni"}
                  {r.total_grams ? ` · ${r.total_grams} g totali` : ""}
                </span>
              </button>
            </li>
          ))
        )}
      </ul>
    </div>
  );
}
