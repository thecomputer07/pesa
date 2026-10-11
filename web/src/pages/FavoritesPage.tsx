import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useFavorites } from "../lib/favorites/favorites";
import { useCurrentUserId } from "../lib/auth/useCurrentUserId";
import { apiFetch } from "../lib/api";

export default function FavoritesPage() {
  const { t } = useTranslation();
  const userId = useCurrentUserId();
  const favs = useFavorites(userId ?? 0);
  const [liveNames, setLiveNames] = useState<Record<number, { name_it: string; category?: string }>>({});

  // Per i food non ancora in Dexie (es. custom foods creati lato server),
  // fetchiamo live il nome via /api/v1/foods/{id}.
  useEffect(() => {
    if (!favs) return;
    const missing = favs.filter((f) => !f.food).map((f) => f.food_id);
    const todo = missing.filter((id) => !(id in liveNames));
    if (todo.length === 0) return;
    let alive = true;
    void Promise.all(
      todo.map(async (id) => {
        try {
          const res = await apiFetch(`/foods/${id}`);
          if (!res.ok) return null;
          const data = (await res.json()) as { name_it: string; category?: string };
          return [id, data] as const;
        } catch {
          return null;
        }
      }),
    ).then((pairs) => {
      if (!alive) return;
      const next: typeof liveNames = {};
      for (const p of pairs) {
        if (p) next[p[0]] = { name_it: p[1].name_it, category: p[1].category };
      }
      if (Object.keys(next).length > 0) {
        setLiveNames((prev) => ({ ...prev, ...next }));
      }
    });
    return () => {
      alive = false;
    };
  }, [favs, liveNames]);

  if (!favs) {
    return (
      <section>
        <p role="status" className="py-6 text-sm opacity-70">
          {t("a11y.loading")}
        </p>
      </section>
    );
  }

  return (
    <section className="space-y-4">
      <h1 className="text-2xl md:text-3xl font-semibold tracking-[-0.03em]">{t("favorites.title")}</h1>
      {favs.length === 0 ? (
        <div className="space-y-2">
          <p className="text-sm text-[color:var(--color-ink-muted)]">{t("favorites.empty")}</p>
          <Link to="/cerca" className="text-sm underline">
            {t("favorites.cta")}
          </Link>
        </div>
      ) : (
        <ul className="grid gap-2 md:grid-cols-2">
          {favs.map((f) => (
            <li
              key={f.client_id}
              className="border border-[color:var(--color-border)] rounded-xl bg-white"
            >
              <Link
                to={`/alimento/${f.food_id}`}
                className="block p-3 no-underline text-[color:var(--color-ink)]"
              >
                <h3 className="font-semibold truncate">
                  {f.food?.name_it ?? liveNames[f.food_id]?.name_it ?? `Alimento #${f.food_id}`}
                </h3>
                {(f.food?.category ?? liveNames[f.food_id]?.category) ? (
                  <p className="text-xs text-[color:var(--color-ink-muted)] mt-1">
                    {f.food?.category ?? liveNames[f.food_id]?.category}
                  </p>
                ) : null}
              </Link>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
