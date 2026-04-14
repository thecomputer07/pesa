import { useTranslation } from "react-i18next";
import type { FoodDetail } from "../lib/api";
import { formatDate } from "../lib/format";

export interface AttributionFooterProps {
  food: FoodDetail;
  lang?: string;
}

/** Per-detail attribution block: source name, license, last-imported, original URL. */
export default function AttributionFooter({ food, lang = "it" }: AttributionFooterProps) {
  const { t } = useTranslation();
  return (
    <footer className="text-xs text-[color:var(--color-ink-muted)] border-t border-[color:var(--color-border)] pt-3 space-y-1">
      <p>
        <span className="font-medium">{t("detail.attribution")}:</span> {food.attribution}
      </p>
      <p>
        <span className="font-medium">{t("detail.license")}:</span>{" "}
        {food.license_url ? (
          <a href={food.license_url} target="_blank" rel="noopener noreferrer">
            {food.license}
          </a>
        ) : (
          food.license
        )}
      </p>
      {food.last_imported_at ? (
        <p>
          <span className="font-medium">{t("detail.lastImported")}:</span>{" "}
          {formatDate(food.last_imported_at, lang)}
        </p>
      ) : null}
      {food.source_url ? (
        <p>
          <a href={food.source_url} target="_blank" rel="noopener noreferrer">
            {t("detail.viewOriginal")} ↗
          </a>
        </p>
      ) : null}
    </footer>
  );
}
