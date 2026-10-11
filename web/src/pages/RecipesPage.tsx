import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import RecipeList from "../components/recipes/RecipeList";

export default function RecipesPage() {
  const { t } = useTranslation();
  return (
    <section className="space-y-4">
      <header className="flex justify-between items-center">
        <h1 className="text-2xl md:text-3xl font-semibold tracking-[-0.03em]">{t("recipes.title")}</h1>
        <Link
          to="/ricette/nuova"
          className="btn-primary text-sm"
        >
          + {t("recipes.new")}
        </Link>
      </header>
      <RecipeList />
    </section>
  );
}
