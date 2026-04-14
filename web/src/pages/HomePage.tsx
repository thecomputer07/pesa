import { Link, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useState, useTransition } from "react";

export default function HomePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [q, setQ] = useState("");
  const [isPending, startTransition] = useTransition();

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const query = q.trim();
    if (!query) return;
    startTransition(() => {
      navigate(`/cerca?q=${encodeURIComponent(query)}`);
    });
  }

  return (
    <div className="space-y-8">
      <section className="bg-[color:var(--color-surface)] rounded-[color:var(--radius-card)] border border-[color:var(--color-border)] p-6 md:p-8">
        <h1 className="text-3xl md:text-4xl font-bold text-[color:var(--color-ink)]">
          {t("home.title")}
        </h1>
        <p className="mt-3 text-[color:var(--color-ink-muted)] text-lg">{t("home.tagline")}</p>

        <form onSubmit={onSubmit} className="mt-6 flex gap-2" role="search">
          <label htmlFor="home-search" className="sr-only">
            {t("search.label")}
          </label>
          <input
            id="home-search"
            type="search"
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder={t("search.placeholder")}
            className="flex-1 px-3 py-2 rounded-md border border-[color:var(--color-border)] bg-white"
            autoComplete="off"
          />
          <button
            type="submit"
            disabled={isPending}
            className="px-4 py-2 rounded-md bg-[color:var(--color-brand-600)] hover:bg-[color:var(--color-brand-700)] text-white font-medium"
          >
            {t("search.submit")}
          </button>
        </form>
      </section>

      <section className="grid gap-4 md:grid-cols-3">
        <FeatureCard title={t("home.features.grams.title")} body={t("home.features.grams.body")} />
        <FeatureCard
          title={t("home.features.offline.title")}
          body={t("home.features.offline.body")}
        />
        <FeatureCard
          title={t("home.features.sources.title")}
          body={t("home.features.sources.body")}
        />
      </section>

      <section className="text-sm text-[color:var(--color-ink-muted)]">
        <p>
          {t("home.learnMore")} <Link to="/scopo">{t("nav.scope")}</Link>.
        </p>
      </section>
    </div>
  );
}

function FeatureCard({ title, body }: { title: string; body: string }) {
  return (
    <div className="p-4 bg-[color:var(--color-surface)] border border-[color:var(--color-border)] rounded-[color:var(--radius-card)]">
      <h2 className="font-semibold text-[color:var(--color-ink)]">{title}</h2>
      <p className="mt-1 text-sm text-[color:var(--color-ink-muted)]">{body}</p>
    </div>
  );
}
