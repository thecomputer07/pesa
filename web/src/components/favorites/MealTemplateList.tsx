import { useTranslation } from "react-i18next";
import { deleteTemplate, useTemplates } from "../../lib/favorites/templates";

export interface MealTemplateListProps {
  userId: number;
  onPick?: (clientId: string) => void;
}

export default function MealTemplateList({ userId, onPick }: MealTemplateListProps) {
  const { t } = useTranslation();
  const templates = useTemplates(userId);

  if (!templates) {
    return (
      <p role="status" className="py-6 text-sm opacity-70">
        {t("a11y.loading")}
      </p>
    );
  }
  if (templates.length === 0) {
    return <p className="py-6 text-sm text-[color:var(--color-ink-muted)]">{t("templates.empty")}</p>;
  }

  return (
    <ul className="space-y-2">
      {templates.map((tpl) => {
        const items = safeParse(tpl.payload_json);
        return (
          <li
            key={tpl.client_id}
            className="border border-[color:var(--color-border)] rounded-xl bg-white p-3"
          >
            <div className="flex justify-between items-start gap-2">
              <div className="min-w-0">
                <h3 className="font-semibold truncate">{tpl.name}</h3>
                <p className="text-xs text-[color:var(--color-ink-muted)]">
                  {items.length} {items.length === 1 ? "elemento" : "elementi"}
                </p>
              </div>
              <div className="flex gap-2">
                {onPick ? (
                  <button
                    type="button"
                    onClick={() => onPick(tpl.client_id)}
                    className="text-xs px-2 py-1 bg-[color:var(--color-brand-600)] text-white rounded"
                  >
                    {t("common.confirm")}
                  </button>
                ) : null}
                <button
                  type="button"
                  onClick={() => void deleteTemplate(tpl.client_id)}
                  className="text-xs px-2 py-1 text-red-700"
                >
                  {t("templates.delete")}
                </button>
              </div>
            </div>
          </li>
        );
      })}
    </ul>
  );
}

function safeParse(s: string): unknown[] {
  try {
    const v = JSON.parse(s);
    return Array.isArray(v) ? v : [];
  } catch {
    return [];
  }
}
