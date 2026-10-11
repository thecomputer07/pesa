import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { marked } from "marked";

type Slug = "privacy" | "cookie";

export function LegalPage({ slug }: { slug: Slug }) {
  const { i18n, t } = useTranslation();
  const lang = i18n.language.startsWith("en") ? "en" : "it";
  const [html, setHtml] = useState<string>("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    setHtml("");
    setError(null);
    fetch(`/legal/${slug}-${lang}.md`)
      .then((r) => {
        if (!r.ok) throw new Error(String(r.status));
        return r.text();
      })
      .then((md) => {
        if (cancelled) return;
        setHtml(marked.parse(md, { async: false }) as string);
      })
      .catch(() => {
        if (!cancelled) setError(t("legal.loadError"));
      });
    return () => {
      cancelled = true;
    };
  }, [slug, lang, t]);

  return (
    <section className="surface mx-auto max-w-3xl p-6 sm:p-10" aria-labelledby="legal-title">
      <h1 id="legal-title" className="sr-only">
        {t(`nav.${slug === "privacy" ? "privacy" : "cookies"}`)}
      </h1>
      {error ? (
        <p role="alert" className="text-red-700">
          {error}
        </p>
      ) : (
        <article
          className="prose prose-sm max-w-none [&_h1]:text-2xl [&_h1]:font-bold [&_h1]:mt-0 [&_h1]:mb-4 [&_h2]:text-xl [&_h2]:font-semibold [&_h2]:mt-6 [&_h2]:mb-2 [&_h3]:text-lg [&_h3]:font-semibold [&_h3]:mt-4 [&_h3]:mb-2 [&_p]:mb-3 [&_ul]:list-disc [&_ul]:pl-6 [&_ul]:mb-3 [&_ol]:list-decimal [&_ol]:pl-6 [&_ol]:mb-3 [&_li]:mb-1 [&_table]:w-full [&_table]:border-collapse [&_table]:text-sm [&_th]:border [&_th]:border-[color:var(--color-border)] [&_th]:p-2 [&_th]:bg-[color:var(--color-surface-muted)] [&_td]:border [&_td]:border-[color:var(--color-border)] [&_td]:p-2 [&_td]:align-top [&_a]:text-[color:var(--color-brand-600)] [&_a]:underline"
          dangerouslySetInnerHTML={{ __html: html }}
        />
      )}
    </section>
  );
}

export function PrivacyLegalPage() {
  return <LegalPage slug="privacy" />;
}

export function CookieLegalPage() {
  return <LegalPage slug="cookie" />;
}
