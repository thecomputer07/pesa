import { useTranslation } from "react-i18next";

export interface SourceBadgeProps {
  code: string;
  license?: string;
  licenseUrl?: string;
  className?: string;
}

const STYLE_BY_CODE: Record<string, string> = {
  CREA: "bg-emerald-50 text-emerald-800 border-emerald-200",
  CIQUAL: "bg-indigo-50 text-indigo-800 border-indigo-200",
  BDA: "bg-amber-50 text-amber-800 border-amber-200",
  OFF: "bg-sky-50 text-sky-800 border-sky-200",
};

/** Small pill identifying the origin dataset of a nutritional value. */
export default function SourceBadge({
  code,
  license,
  licenseUrl,
  className = "",
}: SourceBadgeProps) {
  const { t } = useTranslation();
  const style = STYLE_BY_CODE[code] ?? "bg-slate-50 text-slate-700 border-slate-200";
  const label = `${t("a11y.sourceBadge")}: ${code}${license ? ` (${license})` : ""}`;

  const inner = (
    <span
      className={`inline-flex items-center gap-1 text-xs font-medium border rounded-full px-2 py-0.5 ${style} ${className}`}
      aria-label={label}
      title={label}
    >
      {code}
      {license ? <span className="opacity-70 font-normal">· {license}</span> : null}
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
