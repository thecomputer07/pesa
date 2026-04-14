import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";

const STORAGE_KEY = "pesa.scope_banner_dismissed_v1";

/**
 * Persistent top-of-page notice clarifying Pesa is a generic nutritional
 * carbohydrate calculator. Dismissible; persists in localStorage.
 * Copy is anti-word-scan compliant (no medical/diagnostic/therapeutic terms).
 */
export default function ScopeBanner() {
  const { t } = useTranslation();
  const [dismissed, setDismissed] = useState<boolean>(true);

  useEffect(() => {
    try {
      setDismissed(localStorage.getItem(STORAGE_KEY) === "1");
    } catch {
      setDismissed(false);
    }
  }, []);

  if (dismissed) return null;

  const onDismiss = () => {
    try {
      localStorage.setItem(STORAGE_KEY, "1");
    } catch {
      /* ignore */
    }
    setDismissed(true);
  };

  return (
    <aside
      role="note"
      aria-labelledby="scope-banner-title"
      className="bg-[color:var(--color-brand-50)] border-b border-[color:var(--color-brand-100)] text-[color:var(--color-brand-700)]"
    >
      <div className="max-w-5xl mx-auto px-4 py-2 flex items-start gap-3 text-sm">
        <p id="scope-banner-title" className="flex-1">
          <strong className="font-semibold">{t("scopeBanner.badge")}</strong>{" "}
          <span className="text-[color:var(--color-ink)]">{t("scopeBanner.body")}</span>{" "}
          <Link to="/scopo" className="whitespace-nowrap">
            {t("scopeBanner.learnMore")}
          </Link>
        </p>
        <button
          type="button"
          onClick={onDismiss}
          aria-label={t("scopeBanner.dismissLabel")}
          className="shrink-0 px-2 py-1 rounded hover:bg-white/60"
        >
          ×
        </button>
      </div>
    </aside>
  );
}
