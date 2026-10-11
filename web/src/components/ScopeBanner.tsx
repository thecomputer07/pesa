import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import Icon from "./ui/Icon";

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
    <aside role="note" aria-labelledby="scope-banner-title" className="max-w-6xl w-full mx-auto px-4 md:px-6 pt-4">
      <div className="flex items-start gap-3 rounded-2xl bg-white/70 px-4 py-3 text-sm shadow-[0_0_0_1px_var(--color-border)] backdrop-blur">
        <Icon name="info" size={18} className="mt-0.5 shrink-0 text-[color:var(--color-brand-600)]" />
        <p id="scope-banner-title" className="flex-1 text-[color:var(--color-ink-muted)]">
          <strong className="font-semibold text-[color:var(--color-ink)]">{t("scopeBanner.badge")}</strong>{" "}
          {t("scopeBanner.body")}{" "}
          <Link to="/scopo" className="whitespace-nowrap font-medium">
            {t("scopeBanner.learnMore")}
          </Link>
        </p>
        <button
          type="button"
          onClick={onDismiss}
          aria-label={t("scopeBanner.dismissLabel")}
          className="shrink-0 grid place-items-center w-8 !min-h-8 h-8 -mr-1 -mt-0.5 rounded-full text-[color:var(--color-ink-faint)] hover:text-[color:var(--color-ink)] hover:bg-black/[0.05]"
        >
          <Icon name="close" size={16} />
        </button>
      </div>
    </aside>
  );
}
