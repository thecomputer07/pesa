import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { todayRomeIso } from "../lib/db/queries";
import { db } from "../lib/db/schema";
import type { LocalMeal, LocalMealItem } from "../lib/sync/types";
import WeekStrip from "../components/diary/WeekStrip";
import DayTotals from "../components/diary/DayTotals";
import DayTimeline from "../components/diary/DayTimeline";
import MealEditorSheet from "../components/diary/MealEditorSheet";
import { useCurrentUserId } from "../lib/auth/useCurrentUserId";

export default function DiaryPage() {
  const { t } = useTranslation();
  const { date } = useParams<{ date?: string }>();
  const navigate = useNavigate();
  const userId = useCurrentUserId();
  const iso = date ?? todayRomeIso();
  const [editor, setEditor] = useState<{
    open: boolean;
    meal?: LocalMeal;
    items?: LocalMealItem[];
    mode: "create" | "edit";
  }>({ open: false, mode: "create" });
  const [toastMsg, setToastMsg] = useState<string | null>(null);

  useEffect(() => {
    if (!date) {
      navigate(`/diario/${todayRomeIso()}`, { replace: true });
    }
  }, [date, navigate]);

  useEffect(() => {
    function onConflict() {
      setToastMsg(t("sync.conflict"));
      window.setTimeout(() => setToastMsg(null), 4000);
    }
    window.addEventListener("pesa:conflict", onConflict as EventListener);
    return () => window.removeEventListener("pesa:conflict", onConflict as EventListener);
  }, [t]);

  async function openEdit(meal: LocalMeal) {
    const items = await db.meal_items
      .where("meal_client_id")
      .equals(meal.client_id)
      .and((i) => !i.deleted_at_ms)
      .toArray();
    setEditor({ open: true, mode: "edit", meal, items });
  }

  return (
    <main className="space-y-3" aria-labelledby="diary-title">
      <h1 id="diary-title" className="sr-only">
        {t("diary.title")}
      </h1>

      <WeekStrip isoDate={iso} />
      <DayTotals isoDate={iso} />
      <DayTimeline isoDate={iso} onEdit={openEdit} />

      <button
        type="button"
        aria-label={t("diary.addMeal")}
        onClick={() => setEditor({ open: true, mode: "create" })}
        className="fixed bottom-6 right-6 w-14 h-14 rounded-full bg-[color:var(--color-brand-600)] text-white text-2xl shadow-lg"
      >
        +
      </button>

      {userId != null && editor.open ? (
        <MealEditorSheet
          open={editor.open}
          mode={editor.mode}
          isoDate={iso}
          userId={userId}
          initialMeal={editor.meal}
          initialItems={editor.items}
          onClose={() => setEditor({ open: false, mode: "create" })}
        />
      ) : null}

      {toastMsg ? (
        <div
          role="status"
          aria-live="polite"
          className="fixed bottom-24 left-1/2 -translate-x-1/2 bg-black text-white px-4 py-2 rounded-md text-sm"
        >
          {toastMsg}
        </div>
      ) : null}
    </main>
  );
}
