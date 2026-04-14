import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import type { Source } from "./SourcesDataPage";

interface Response {
  sources: Source[];
}

/**
 * SourcesImagesPage lists image-source attributions (OFF hosted under ODbL,
 * Wikimedia per-file) and clarifies that CREA/BDA images are NOT hosted.
 */
export default function SourcesImagesPage() {
  const { t } = useTranslation();
  const [data, setData] = useState<Source[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetch("/api/v1/sources")
      .then((r) => {
        if (!r.ok) throw new Error(String(r.status));
        return r.json() as Promise<Response>;
      })
      .then((j) => {
        if (!cancelled) setData(j.sources);
      })
      .catch((e) => {
        if (!cancelled) setError(String(e));
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const rows = useMemo(
    () =>
      (data ?? []).filter(
        (s) => s.scope === "images" || s.scope === "data+images",
      ),
    [data],
  );

  if (error) return <p role="alert">{t("sources.loadError")}</p>;
  if (!data) return <p role="status">{t("sources.loading")}</p>;

  return (
    <section>
      <h1 className="mb-3 text-2xl font-semibold">{t("sources.imagesPage.title")}</h1>
      <p className="mb-4 text-sm opacity-80">{t("sources.imagesPage.intro")}</p>

      <div className="mb-6 rounded-md border border-amber-500/40 bg-amber-50/40 p-3 text-sm">
        <strong>{t("sources.imagesPage.creaBdaNote")}</strong>
      </div>

      {rows.map((s) => (
        <section
          key={s.code}
          className="mb-6 rounded-lg border border-[color:var(--color-border)] p-4"
        >
          <h2 className="text-lg font-semibold">{s.name}</h2>
          <dl className="mt-2 grid gap-1 text-sm">
            <div>
              <dt className="inline font-medium">{t("sources.common.license")}: </dt>
              <dd className="inline">{s.license}</dd>
            </div>
            {s.license_url && (
              <div>
                <dt className="inline font-medium">{t("sources.common.licenseUrl")}: </dt>
                <dd className="inline">
                  <a href={s.license_url} target="_blank" rel="noopener noreferrer">
                    {s.license_url}
                  </a>
                </dd>
              </div>
            )}
            <div>
              <dt className="inline font-medium">{t("sources.common.attribution")}: </dt>
              <dd className="inline">{s.attribution}</dd>
            </div>
            {s.notes && <p className="mt-2 opacity-80">{s.notes}</p>}
          </dl>
        </section>
      ))}
    </section>
  );
}
