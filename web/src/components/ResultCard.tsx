import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import type { FoodSearchResult } from "../lib/api";
import { useFormatGrams, useFormatKcal } from "../lib/format";
import SourceBadge from "./SourceBadge";
import FavoriteToggle from "./favorites/FavoriteToggle";

export interface ResultCardProps {
  food: FoodSearchResult;
}

/** One search result row: what it is on the left, its carbs on the right. */
export default function ResultCard({ food }: ResultCardProps) {
  const { t } = useTranslation();
  const fmtG = useFormatGrams();
  const fmtKcal = useFormatKcal();
  // "Crudo" on Nutella or yogurt is noise: label only cooked items.
  const isCooked = food.state === "cooked" || food.state === "cotto";

  return (
    <li className="group relative flex items-center gap-2 pr-2 transition-colors hover:bg-[color:var(--color-canvas)]/70 focus-within:bg-[color:var(--color-canvas)]/70">
      {/* The link is the whole row; the favourite button sits beside it, never inside. */}
      <Link
        to={`/alimento/${food.id}`}
        className="flex-1 min-w-0 flex items-center gap-4 p-3 sm:p-4 no-underline text-[color:var(--color-ink)] hover:text-[color:var(--color-ink)]"
      >
        <Thumb url={food.thumb_url} name={food.name_it} />

        <div className="min-w-0 flex-1">
          <h3 className="font-medium leading-snug line-clamp-2 sm:truncate">{food.name_it}</h3>
          <p className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-[color:var(--color-ink-muted)]">
            <SourceBadge code={food.source.code} license={food.source.license} compact />
            {food.category ? <span className="truncate max-w-[24ch]">{food.category}</span> : null}
            {isCooked ? (
              <span className="rounded-md bg-[color:var(--color-carb-50)] px-1.5 py-0.5 font-medium text-[color:var(--color-carb-700)]">
                {t("search.stateCotto")}
              </span>
            ) : null}
            <span className="tabular hidden sm:inline">{fmtKcal(food.kcal)}</span>
          </p>
        </div>

        <div className="shrink-0 text-right">
          <p className="tabular text-lg sm:text-xl font-semibold tracking-[-0.02em]">{fmtG(food.carb_g)}</p>
          <p className="text-[11px] text-[color:var(--color-ink-faint)]">
            {t("detail.carb").toLowerCase()} · 100 g
          </p>
        </div>
      </Link>
      <FavoriteToggle foodId={food.id} />
    </li>
  );
}

function Thumb({ url, name }: { url?: string | null; name: string }) {
  const initial = name.replace(/^[^\p{L}]+/u, "").slice(0, 1).toUpperCase() || "·";
  return (
    <span className="relative grid place-items-center w-14 h-14 shrink-0 overflow-hidden rounded-2xl bg-[color:var(--color-surface-sunken)] text-lg font-semibold text-[color:var(--color-ink-faint)]">
      <span aria-hidden="true">{initial}</span>
      {url ? (
        <img
          src={url}
          alt=""
          loading="lazy"
          onError={(e) => {
            (e.currentTarget as HTMLImageElement).style.display = "none";
          }}
          className="absolute inset-0 w-full h-full object-cover"
        />
      ) : null}
    </span>
  );
}
