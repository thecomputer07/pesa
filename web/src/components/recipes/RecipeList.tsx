import { useState } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useLiveQuery } from "dexie-react-hooks";
import { db } from "../../lib/db/schema";

export default function RecipeList() {
  const { t } = useTranslation();
  const [q, setQ] = useState("");
  const recipes = useLiveQuery(async () => {
    const all = await db.recipes.toArray();
    const filtered = all.filter((r) => !r.deleted_at_ms);
    if (!q.trim()) return filtered;
    const qq = q.toLowerCase();
    return filtered.filter((r) => r.name.toLowerCase().includes(qq));
  }, [q]);

  if (!recipes) {
    return (
      <p role="status" className="py-6 text-sm opacity-70">
        {t("a11y.loading")}
      </p>
    );
  }
  if (recipes.length === 0) {
    return <p className="py-6 text-sm text-[color:var(--color-ink-muted)]">{t("recipes.empty")}</p>;
  }
  return (
    <div className="space-y-3">
      <input
        value={q}
        onChange={(e) => setQ(e.target.value)}
        placeholder={t("search.placeholder")}
        aria-label={t("search.label")}
        className="w-full border border-[color:var(--color-border)] rounded-xl px-3 py-2"
      />
      <ul className="grid gap-2 md:grid-cols-2">
        {recipes.map((r) => (
          <li
            key={r.client_id}
            className="border border-[color:var(--color-border)] rounded-xl bg-white"
          >
            <Link to={`/ricette/${r.client_id}`} className="block p-3 no-underline text-[color:var(--color-ink)]">
              <h3 className="font-semibold truncate">{r.name}</h3>
              <p className="text-xs text-[color:var(--color-ink-muted)] mt-1">
                {r.servings} {Number(r.servings) === 1 ? "porzione" : "porzioni"}
              </p>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
