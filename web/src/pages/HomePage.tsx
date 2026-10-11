import { Link, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useTransition } from "react";
import SearchBox from "../components/SearchBox";
import Icon, { type IconName } from "../components/ui/Icon";

// Quick queries: common foods that exist in the catalog.
const SUGGESTIONS = ["pasta", "pane", "riso", "pizza", "mela", "sushi"];

export default function HomePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [, startTransition] = useTransition();

  function go(query: string) {
    const q = query.trim();
    if (!q) return;
    startTransition(() => {
      navigate(`/cerca?q=${encodeURIComponent(q)}`);
    });
  }

  return (
    <div className="space-y-16 md:space-y-24">
      <section className="grid grid-cols-[minmax(0,1fr)] gap-10 lg:grid-cols-[minmax(0,1.15fr)_minmax(0,0.85fr)] lg:gap-14 items-center">
        <div className="rise">
          <p className="eyebrow !text-[color:var(--color-brand-600)]">{t("home.eyebrow")}</p>
          <h1 className="mt-4 text-[2.5rem] leading-[1.05] sm:text-5xl md:text-[3.75rem] font-semibold tracking-[-0.035em] text-[color:var(--color-ink)]">
            {t("home.title")}
          </h1>
          <p className="mt-5 max-w-[52ch] text-lg leading-relaxed text-[color:var(--color-ink-muted)]">
            {t("home.tagline")}
          </p>

          <div className="mt-8 max-w-xl">
            <SearchBox size="lg" onSubmit={go} />
          </div>

          <div className="mt-5 flex flex-wrap items-center gap-2">
            <span className="text-sm text-[color:var(--color-ink-faint)] mr-1">{t("home.tryLabel")}</span>
            {SUGGESTIONS.map((s) => (
              <button key={s} type="button" onClick={() => go(s)} className="chip !min-h-9">
                {s}
              </button>
            ))}
          </div>

          <Link
            to="/scan"
            className="mt-8 inline-flex items-center gap-2 text-sm font-medium no-underline text-[color:var(--color-ink)] hover:text-[color:var(--color-brand-600)]"
          >
            <span className="grid place-items-center w-9 h-9 rounded-xl bg-white shadow-[0_0_0_1px_var(--color-border)]">
              <Icon name="scan" size={18} />
            </span>
            {t("home.scanCta")}
            <Icon name="arrow" size={16} />
          </Link>
        </div>

        <ExampleDisplay />
      </section>

      <section aria-labelledby="home-features" className="rise rise-3">
        <h2 id="home-features" className="sr-only">
          {t("nav.scope")}
        </h2>
        <ul className="grid gap-px overflow-hidden rounded-[var(--radius-card)] bg-[color:var(--color-border)] shadow-[0_0_0_1px_var(--color-border)] md:grid-cols-3">
          <Feature icon="scale" title={t("home.features.grams.title")} body={t("home.features.grams.body")} />
          <Feature icon="wifiOff" title={t("home.features.offline.title")} body={t("home.features.offline.body")} />
          <Feature icon="book" title={t("home.features.sources.title")} body={t("home.features.sources.body")} />
        </ul>
        <p className="mt-6 text-sm text-[color:var(--color-ink-muted)]">
          {t("home.learnMore")} <Link to="/scopo">{t("nav.scope")}</Link>.
        </p>
      </section>
    </div>
  );
}

/** A real catalog calculation, shown like the read-out of a kitchen scale. */
function ExampleDisplay() {
  const { t } = useTranslation();
  return (
    <figure className="rise rise-2 relative mx-auto w-full max-w-md lg:max-w-none" aria-label={t("home.example.label")}>
      <div aria-hidden="true" className="absolute inset-x-0 -inset-y-6 lg:-inset-6 -z-10 rounded-[2.5rem] bg-[radial-gradient(closest-side,rgb(49_87_224/0.18),transparent)] blur-2xl" />
      <div className="relative overflow-hidden rounded-[1.75rem] bg-[color:var(--color-ink)] p-6 sm:p-8 text-white shadow-[0_30px_60px_-30px_rgb(20_30_80/0.6)]">
        <div aria-hidden="true" className="absolute inset-0 bg-[radial-gradient(120%_80%_at_100%_0%,rgb(49_87_224/0.45),transparent_60%)]" />
        <div className="relative">
          <figcaption className="text-xs font-medium tracking-wide text-white/55">{t("home.example.label")}</figcaption>
          <div className="mt-6 flex items-center justify-between gap-4">
            <p className="text-xl font-semibold tracking-[-0.02em]">{t("home.example.food")}</p>
            <span className="rounded-full bg-white/10 px-3 py-1 text-sm text-white/80 ring-1 ring-white/15 tabular">
              {t("home.example.portion")}
            </span>
          </div>

          <p className="mt-10 flex items-baseline gap-2">
            <span className="tabular text-[4.5rem] sm:text-[5.5rem] leading-none font-semibold tracking-[-0.05em] text-[color:var(--color-carb-100)]">
              58,2
            </span>
            <span className="text-2xl font-medium text-[color:var(--color-carb-100)]/80">g</span>
          </p>
          <p className="mt-2 text-white/70">{t("home.example.carbs")}</p>

          <div className="mt-8 h-2 rounded-full bg-white/10 overflow-hidden" aria-hidden="true">
            <div className="h-full w-[72.7%] rounded-full bg-gradient-to-r from-[color:var(--color-carb-500)] to-[color:var(--color-carb-100)]" />
          </div>
          <p className="mt-3 text-xs text-white/50 tabular">{t("home.example.per100")}</p>
        </div>
      </div>
    </figure>
  );
}

function Feature({ icon, title, body }: { icon: IconName; title: string; body: string }) {
  return (
    <li className="bg-white p-6 md:p-7">
      <span className="grid place-items-center w-10 h-10 rounded-xl bg-[color:var(--color-brand-50)] text-[color:var(--color-brand-600)]">
        <Icon name={icon} size={20} />
      </span>
      <h3 className="mt-5 font-semibold text-[color:var(--color-ink)]">{title}</h3>
      <p className="mt-1.5 text-sm leading-relaxed text-[color:var(--color-ink-muted)]">{body}</p>
    </li>
  );
}
