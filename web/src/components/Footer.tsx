import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import ScopeFooterNote from "./ScopeFooterNote";

export default function Footer() {
  const { t, i18n } = useTranslation();
  const year = new Date().getFullYear();

  return (
    <footer
      role="contentinfo"
      className="bg-[color:var(--color-surface)] border-t border-[color:var(--color-border)] mt-auto"
    >
      <div className="max-w-5xl mx-auto px-4 py-6 text-sm text-[color:var(--color-ink-muted)] space-y-4">
        <ScopeFooterNote />

        <nav aria-label={t("a11y.footerNav")} className="flex flex-wrap gap-x-4 gap-y-2">
          <Link to="/scopo">{t("nav.scope")}</Link>
          <Link to="/chi-siamo">{t("nav.about")}</Link>
          <Link to="/accessibilita">{t("nav.accessibility")}</Link>
          <a href="/privacy">{t("nav.privacy")}</a>
          <a href="/terms">{t("nav.terms")}</a>
          <a href="/cookies">{t("nav.cookies")}</a>
          <a href="/.well-known/security.txt">security.txt</a>
        </nav>

        <nav aria-label={t("footer.sourcesAria")} className="flex flex-wrap gap-x-4 gap-y-2">
          <span className="font-medium">{t("footer.sources")}:</span>
          <Link to="/fonti-dati">{t("footer.sourcesData")}</Link>
          <Link to="/fonti-immagini">{t("footer.sourcesImages")}</Link>
        </nav>

        <div className="flex flex-wrap items-center justify-between gap-2 pt-2 border-t border-[color:var(--color-border)]">
          <span>
            © {year} {t("footer.rights")}
          </span>
          <div className="flex items-center gap-2">
            <label htmlFor="lang-select" className="sr-only">
              {t("footer.language")}
            </label>
            <select
              id="lang-select"
              className="bg-transparent border border-[color:var(--color-border)] rounded px-2 py-1"
              value={i18n.language.startsWith("en") ? "en" : "it"}
              onChange={(e) => void i18n.changeLanguage(e.target.value)}
            >
              <option value="it">Italiano</option>
              <option value="en">English</option>
            </select>
          </div>
        </div>
      </div>
    </footer>
  );
}
