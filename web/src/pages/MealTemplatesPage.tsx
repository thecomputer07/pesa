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
    <section className="space-y-4">
      <header className="flex justify-between items-center">
        <h1 className="text-2xl md:text-3xl font-semibold tracking-[-0.03em]">{t("templates.title")}</h1>
        <button
          type="button"
          onClick={() => setShowBuilder((v) => !v)}
          className="btn-primary text-sm"
        >
          {showBuilder ? t("common.cancel") : `+ ${t("templates.new")}`}
        </button>
      </header>
      {showBuilder ? <MealTemplateBuilder userId={userId} /> : <MealTemplateList userId={userId} />}
    </section>
  );
}
