import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { saveTemplate, type TemplateItem } from "../../lib/favorites/templates";

export interface MealTemplateBuilderProps {
  userId: number;
}

export default function MealTemplateBuilder({ userId }: MealTemplateBuilderProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [name, setName] = useState("");
  const [items, setItems] = useState<TemplateItem[]>([]);
  const [newName, setNewName] = useState("");
  const [newGrams, setNewGrams] = useState(100);

  function addItem() {
    const trimmed = newName.trim();
    if (!trimmed || !(newGrams > 0)) return;
    setItems((prev) => [
      ...prev,
      {
        name_snapshot: trimmed,
        grams: newGrams,
        snapshot: { source_food_name: trimmed },
      },
    ]);
    setNewName("");
    setNewGrams(100);
  }

  async function onSave() {
    if (!name.trim() || items.length === 0) return;
    await saveTemplate(name.trim(), items, userId);
    navigate("/template");
  }

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        void onSave();
      }}
      className="space-y-3"
    >
      <label className="block">
        <span className="text-sm font-medium">{t("templates.nameLabel")}</span>
        <input
          required
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="mt-1 w-full border border-[color:var(--color-border)] rounded-md px-2 py-1.5"
        />
      </label>

      <ul className="divide-y divide-[color:var(--color-border)]">
        {items.map((i, idx) => (
          <li key={idx} className="py-2 flex justify-between text-sm">
            <span>{i.name_snapshot}</span>
            <span className="text-xs text-[color:var(--color-ink-muted)]">{i.grams} g</span>
          </li>
        ))}
      </ul>

      <div className="flex gap-2">
        <input
          aria-label={t("diary.itemName")}
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          placeholder={t("diary.itemName")}
          className="flex-1 border border-[color:var(--color-border)] rounded-md px-2 py-1.5 text-sm"
        />
        <input
          aria-label={t("detail.weightLabel")}
          type="number"
          min={0}
          value={newGrams}
          onChange={(e) => setNewGrams(Number(e.target.value))}
          className="w-24 border border-[color:var(--color-border)] rounded-md px-2 py-1.5 text-sm"
        />
        <button
          type="button"
          onClick={addItem}
          className="px-3 text-sm bg-[color:var(--color-brand-600)] text-white rounded"
        >
          +
        </button>
      </div>

      <button
        type="submit"
        className="px-3 py-1.5 text-sm bg-[color:var(--color-brand-600)] text-white rounded-md"
      >
        {t("common.save")}
      </button>
    </form>
  );
}
