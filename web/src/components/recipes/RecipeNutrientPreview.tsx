import { useTranslation } from "react-i18next";
import type { RecipeAggregate } from "../../lib/recipes/aggregate";
import { useFormatGrams, useFormatKcal } from "../../lib/format";

export interface RecipeNutrientPreviewProps {
  aggregate: RecipeAggregate;
}

export default function RecipeNutrientPreview({ aggregate }: RecipeNutrientPreviewProps) {
  const { t } = useTranslation();
  const fmtG = useFormatGrams();
  const fmtKcal = useFormatKcal();

  const rows: Array<[string, typeof aggregate.total]> = [
    [t("recipes.preview.total"), aggregate.total],
    [t("recipes.preview.perServing"), aggregate.perServing],
    [t("recipes.preview.per100g"), aggregate.per100g],
  ];

  return (
    <table className="w-full text-sm border-collapse">
      <thead>
        <tr className="text-left text-[color:var(--color-ink-muted)] text-xs uppercase">
          <th className="py-1"></th>
          <th className="py-1">kcal</th>
          <th className="py-1">{t("detail.carb")}</th>
          <th className="py-1">{t("detail.protein")}</th>
          <th className="py-1">{t("detail.fat")}</th>
        </tr>
      </thead>
      <tbody>
        {rows.map(([label, n]) => (
          <tr key={label} className="border-t border-[color:var(--color-border)]">
            <th scope="row" className="py-1 font-medium">
              {label}
            </th>
            <td className="py-1">{fmtKcal(n.kcal)}</td>
            <td className="py-1">{fmtG(n.carb_g)}</td>
            <td className="py-1">{fmtG(n.protein_g)}</td>
            <td className="py-1">{fmtG(n.fat_g)}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
