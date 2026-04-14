import { useTranslation } from "react-i18next";
import { toggleFavorite, useIsFavorited } from "../../lib/favorites/favorites";
import { flushOutbox } from "../../lib/sync/worker";
import { useCurrentUserId } from "../../lib/auth/useCurrentUserId";

export interface FavoriteToggleProps {
  foodId: number;
  userId?: number;
  className?: string;
}

export default function FavoriteToggle({ foodId, userId, className }: FavoriteToggleProps) {
  const { t } = useTranslation();
  const authUid = useCurrentUserId();
  const uid = userId ?? authUid ?? 0;
  const isFav = useIsFavorited(foodId, uid);

  function onClick() {
    if (!uid) return; // not authenticated — silently no-op
    void toggleFavorite(foodId, uid).then(() => flushOutbox());
  }

  if (!uid) return null;

  return (
    <button
      type="button"
      aria-pressed={isFav === true}
      aria-label={isFav ? t("favorites.remove") : t("favorites.add")}
      onClick={onClick}
      className={
        className ??
        "text-xl leading-none px-2 py-1 rounded hover:bg-[color:var(--color-surface-hover)]"
      }
    >
      {isFav ? "★" : "☆"}
    </button>
  );
}
