import { useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";

export interface WeekStripProps {
  isoDate: string; // selected day
}

/** Week strip: Monday..Sunday buttons. Tap navigates to /diario/:iso. */
export default function WeekStrip({ isoDate }: WeekStripProps) {
  const navigate = useNavigate();
  const { i18n } = useTranslation();
  const days = useMemo(() => computeWeekDays(isoDate), [isoDate]);
  const fmt = useMemo(
    () =>
      new Intl.DateTimeFormat(i18n.language || "it", {
        weekday: "short",
        day: "2-digit",
      }),
    [i18n.language],
  );

  return (
    <nav aria-label="settimana" className="flex gap-1 py-2 overflow-x-auto">
      {days.map((iso) => {
        const active = iso === isoDate;
        const d = new Date(`${iso}T12:00:00Z`);
        return (
          <button
            key={iso}
            type="button"
            aria-current={active ? "date" : undefined}
            onClick={() => navigate(`/diario/${iso}`)}
            className={`min-w-[3.5rem] px-2 py-1.5 rounded-md text-xs border ${
              active
                ? "bg-[color:var(--color-brand-600)] text-white border-transparent"
                : "bg-white border-[color:var(--color-border)]"
            }`}
          >
            {fmt.format(d)}
          </button>
        );
      })}
    </nav>
  );
}

function computeWeekDays(isoDate: string): string[] {
  const [y, m, d] = isoDate.split("-").map(Number);
  const target = new Date(Date.UTC(y, m - 1, d));
  // Europe/Rome week: Monday = 1 … Sunday = 7. getUTCDay: Sun=0..Sat=6.
  const dow = target.getUTCDay() || 7; // Sun→7
  const monday = new Date(target.getTime() - (dow - 1) * 86_400_000);
  const out: string[] = [];
  for (let i = 0; i < 7; i++) {
    const cur = new Date(monday.getTime() + i * 86_400_000);
    const pad = (n: number) => String(n).padStart(2, "0");
    out.push(`${cur.getUTCFullYear()}-${pad(cur.getUTCMonth() + 1)}-${pad(cur.getUTCDate())}`);
  }
  return out;
}
