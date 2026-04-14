import { useTranslation } from "react-i18next";
import { useDayTotals } from "../../lib/db/queries";
import { useFormatGrams, useFormatKcal } from "../../lib/format";

export interface DayTotalsProps {
  isoDate: string;
}

export default function DayTotals({ isoDate }: DayTotalsProps) {
  const { t } = useTranslation();
  const totals = useDayTotals(isoDate);
  const fmtG = useFormatGrams();
  const fmtKcal = useFormatKcal();

  return (
    <section aria-label={t("diary.totals.ariaLabel")} className="flex flex-wrap gap-2 py-2">
      <Chip label={t("diary.totals.kcal")} value={fmtKcal(totals?.kcal)} emphasise />
      <Chip label={t("diary.totals.carbs")} value={fmtG(totals?.carb_g)} emphasise />
      <Chip label={t("diary.totals.protein")} value={fmtG(totals?.protein_g)} />
      <Chip label={t("diary.totals.fat")} value={fmtG(totals?.fat_g)} />
      <Chip label={t("diary.totals.fiber")} value={fmtG(totals?.fiber_g)} />
    </section>
  );
}

function Chip({
  label,
  value,
  emphasise = false,
}: {
  label: string;
  value: string;
  emphasise?: boolean;
}) {
  return (
    <div
      className={`px-3 py-1.5 rounded-full border border-[color:var(--color-border)] bg-white text-sm ${
        emphasise ? "font-semibold" : ""
      }`}
    >
      <span className="text-[10px] uppercase tracking-wide text-[color:var(--color-ink-muted)] mr-1.5">
        {label}
      </span>
      {value}
    </div>
  );
}
