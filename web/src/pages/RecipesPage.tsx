import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import RecipeList from "../components/recipes/RecipeList";

export default function RecipesPage() {
  const { t } = useTranslation();
  return (
    <main className="space-y-4">
      <header className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">{t("recipes.title")}</h1>
        <Link
          to="/ricette/nuova"
          className="px-3 py-1.5 text-sm bg-[color:var(--color-brand-600)] text-white rounded-md"
        >
          + {t("recipes.new")}
        </Link>
      </header>
      <RecipeList />
    </main>
  );
}
