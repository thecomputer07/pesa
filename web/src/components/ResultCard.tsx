import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import type { FoodSearchResult } from "../lib/api";
import { useFormatGrams, useFormatKcal } from "../lib/format";
import SourceBadge from "./SourceBadge";
import FavoriteToggle from "./favorites/FavoriteToggle";

export interface ResultCardProps {
  food: FoodSearchResult;
}

export default function ResultCard({ food }: ResultCardProps) {
  const { t } = useTranslation();
  const fmtG = useFormatGrams();
  const fmtKcal = useFormatKcal();
  const isCooked = food.state === "cooked" || food.state === "cotto";
  // "Crudo" on Nutella or yogurt is noise: label only cooked items.

  return (
    <li className="border border-[color:var(--color-border)] rounded-[color:var(--radius-card)] bg-white hover:border-[color:var(--color-brand-500)] focus-within:border-[color:var(--color-brand-500)] transition-colors">
      <Link
        to={`/alimento/${food.id}`}
        className="block p-4 no-underline text-[color:var(--color-ink)]"
      >
        <div className="flex items-start justify-between gap-3">
          <div className="flex items-start gap-3 min-w-0">
            {food.thumb_url ? (
              <img
                src={food.thumb_url}
                alt=""
                loading="lazy"
                onError={(e) => {
                  (e.currentTarget as HTMLImageElement).style.display = "none";
                }}
                className="w-14 h-14 object-cover rounded-md border border-[color:var(--color-border)] bg-white shrink-0"
              />
            ) : null}
            <div className="min-w-0">
              <h3 className="font-semibold truncate">{food.name_it}</h3>
              {food.category ? (
                <p className="text-xs text-[color:var(--color-ink-muted)] mt-0.5">{food.category}</p>
              ) : null}
            </div>
          </div>
          <div className="flex items-center gap-1">
            <FavoriteToggle foodId={food.id} className="text-lg leading-none px-1" />
            <SourceBadge
              code={food.source.code}
              license={food.source.license}
              licenseUrl={food.source.license_url}
            />
          </div>
        </div>

        <dl className="mt-3 grid grid-cols-4 gap-2 text-sm">
          <Kv label={t("detail.carb")} value={fmtG(food.carb_g)} emphasise />
          <Kv label={t("detail.kcal")} value={fmtKcal(food.kcal)} />
          <Kv label={t("detail.fiber")} value={fmtG(food.fiber_g)} />
          {isCooked ? <Kv label={t("detail.stateLabel")} value={t("search.stateCotto")} /> : null}
        </dl>
      </Link>
    </li>
  );
}

function Kv({
  label,
  value,
  emphasise = false,
}: {
  label: string;
  value: string;
  emphasise?: boolean;
}) {
  return (
    <div>
      <dt className="text-[10px] uppercase tracking-wide text-[color:var(--color-ink-muted)]">
        {label}
      </dt>
      <dd className={emphasise ? "font-semibold" : ""}>{value}</dd>
    </div>
  );
}
