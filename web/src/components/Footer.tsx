import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import ScopeFooterNote from "./ScopeFooterNote";

export default function Footer() {
  const { t, i18n } = useTranslation();
  const year = new Date().getFullYear();

  const link = "no-underline text-[color:var(--color-ink-muted)] hover:text-[color:var(--color-ink)]";

  return (
    <footer role="contentinfo" className="mt-auto border-t border-black/[0.06] bg-white/50">
      <div className="max-w-6xl mx-auto px-4 md:px-6 pt-10 pb-28 md:pb-10 text-sm grid gap-8 md:grid-cols-[1.3fr_1fr]">
        <div className="space-y-4 max-w-md">
          <div className="flex items-center gap-2.5">
            <img src="/icon.svg" alt="" aria-hidden="true" width={28} height={28} className="w-7 h-7 rounded-lg" />
            <span className="font-semibold tracking-[-0.02em] text-[color:var(--color-ink)]">Pesa</span>
          </div>
          <div className="text-[color:var(--color-ink-muted)]">
            <ScopeFooterNote />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-6">
          <nav aria-label={t("a11y.footerNav")} className="flex flex-col gap-2">
            <Link to="/scopo" className={link}>{t("nav.scope")}</Link>
            <Link to="/chi-siamo" className={link}>{t("nav.about")}</Link>
            <Link to="/accessibilita" className={link}>{t("nav.accessibility")}</Link>
            <a href="/privacy" className={link}>{t("nav.privacy")}</a>
            <a href="/terms" className={link}>{t("nav.terms")}</a>
            <a href="/cookies" className={link}>{t("nav.cookies")}</a>
            <a href="/.well-known/security.txt" className={link}>security.txt</a>
          </nav>
          <nav aria-label={t("footer.sourcesAria")} className="flex flex-col gap-2">
            <span className="eyebrow mb-1">{t("footer.sources")}</span>
            <Link to="/fonti-dati" className={link}>{t("footer.sourcesData")}</Link>
            <Link to="/fonti-immagini" className={link}>{t("footer.sourcesImages")}</Link>
          </nav>
        </div>

        <div className="md:col-span-2 flex flex-wrap items-center justify-between gap-3 pt-6 border-t border-[color:var(--color-border)] text-[color:var(--color-ink-faint)]">
          <span>
            © {year} {t("footer.rights")}
          </span>
          <div className="flex items-center gap-2">
            <label htmlFor="lang-select" className="sr-only">
              {t("footer.language")}
            </label>
            <select
              id="lang-select"
              className="bg-white border border-[color:var(--color-border)] px-3 py-1.5 text-sm text-[color:var(--color-ink)]"
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
