import { useState } from "react";
import { useTranslation } from "react-i18next";
import MealTemplateList from "../components/favorites/MealTemplateList";
import MealTemplateBuilder from "../components/favorites/MealTemplateBuilder";
import { useCurrentUserId } from "../lib/auth/useCurrentUserId";

export default function MealTemplatesPage() {
  const { t } = useTranslation();
  const userId = useCurrentUserId();
  const [showBuilder, setShowBuilder] = useState(false);
  if (userId == null) return null;

  return (
    <main className="space-y-4">
      <header className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">{t("templates.title")}</h1>
        <button
          type="button"
          onClick={() => setShowBuilder((v) => !v)}
          className="px-3 py-1.5 text-sm bg-[color:var(--color-brand-600)] text-white rounded-md"
        >
          {showBuilder ? t("common.cancel") : `+ ${t("templates.new")}`}
        </button>
      </header>
      {showBuilder ? <MealTemplateBuilder userId={userId} /> : <MealTemplateList userId={userId} />}
    </main>
  );
}
