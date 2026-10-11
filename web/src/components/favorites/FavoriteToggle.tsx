import { useTranslation } from "react-i18next";
import Icon from "../ui/Icon";
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
      className={[
        "grid place-items-center w-10 h-10 !min-h-10 rounded-full transition-colors",
        isFav
          ? "text-rose-500 hover:bg-rose-50"
          : "text-[color:var(--color-ink-faint)] hover:text-rose-500 hover:bg-black/[0.04]",
        className ?? "",
      ].join(" ")}
    >
      <Icon name="heart" size={20} fill={isFav ? "currentColor" : "none"} />
    </button>
  );
}
