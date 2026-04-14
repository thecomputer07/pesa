import { useTranslation } from "react-i18next";

/**
 * Persistent scope reminder in the footer. Anti-word-scan compliant:
 * uses positive phrasings from MARKETING-RULES §6.
 */
export default function ScopeFooterNote() {
  const { t } = useTranslation();
  return (
    <p className="text-xs leading-relaxed">
      <strong className="font-semibold">{t("scopeFooter.badge")}</strong>{" "}
      {t("scopeFooter.body")}
    </p>
  );
}
