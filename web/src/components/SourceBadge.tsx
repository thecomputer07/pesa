import { useTranslation } from "react-i18next";

export interface SourceBadgeProps {
  code: string;
  license?: string;
  licenseUrl?: string;
  className?: string;
  /** Hide the licence text (still in the accessible label and tooltip). */
  compact?: boolean;
}

// One quiet style per source: a coloured dot, never a loud pill.
const DOT_BY_CODE: Record<string, string> = {
  CREA: "bg-emerald-500",
  CIQUAL: "bg-indigo-500",
  BDA: "bg-amber-500",
  OFF: "bg-sky-500",
};

/** Small pill identifying the origin dataset of a nutritional value. */
export default function SourceBadge({
  code,
  license,
  licenseUrl,
  className = "",
  compact = false,
}: SourceBadgeProps) {
  const { t } = useTranslation();
  const shown = code === "off" || code === "OFF" ? "OFF" : code.toUpperCase();
  const dot = DOT_BY_CODE[shown] ?? "bg-slate-400";
  const label = `${t("a11y.sourceBadge")}: ${shown}${license ? ` (${license})` : ""}`;

  const inner = (
    <span
      className={`inline-flex items-center gap-1.5 text-[11px] font-semibold tracking-wide rounded-md px-1.5 py-0.5 bg-[color:var(--color-surface-sunken)] text-[color:var(--color-ink-muted)] ${className}`}
      aria-label={label}
      title={label}
    >
      <span aria-hidden="true" className={`w-1.5 h-1.5 rounded-full ${dot}`} />
      {shown}
      {license && !compact ? <span className="font-normal opacity-80">· {license}</span> : null}
    </span>
  );

  if (licenseUrl) {
    return (
      <a href={licenseUrl} target="_blank" rel="noopener noreferrer" className="no-underline">
        {inner}
      </a>
    );
  }
  return inner;
}
