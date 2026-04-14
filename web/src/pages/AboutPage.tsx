import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";

export default function AboutPage() {
  const { t } = useTranslation();
  return (
    <article className="prose max-w-none space-y-4">
      <h1 className="text-3xl font-bold">{t("about.title")}</h1>
      <p className="text-[color:var(--color-ink-muted)]">{t("about.intro")}</p>
      <p>
        {t("about.scopeLead")} <Link to="/scopo">{t("nav.scope")}</Link>.
      </p>
      <p>
        {t("about.a11yLead")} <Link to="/accessibilita">{t("nav.accessibility")}</Link>.
      </p>
      <p className="text-sm text-[color:var(--color-ink-muted)]">
        {t("about.openSource", { license: "AGPL v3" })}
      </p>
    </article>
  );
}
