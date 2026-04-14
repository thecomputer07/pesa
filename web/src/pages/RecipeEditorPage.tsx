import { useParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import RecipeEditor from "../components/recipes/RecipeEditor";
import { useCurrentUserId } from "../lib/auth/useCurrentUserId";

export default function RecipeEditorPage() {
  const { t } = useTranslation();
  const { clientId } = useParams<{ clientId: string }>();
  const userId = useCurrentUserId();
  if (userId == null) return null; // RequireAuth gate will redirect
  return (
    <main className="space-y-4">
      <h1 className="text-2xl font-bold">{clientId ? t("recipes.edit") : t("recipes.new")}</h1>
      <RecipeEditor clientId={clientId} userId={userId} />
    </main>
  );
}
