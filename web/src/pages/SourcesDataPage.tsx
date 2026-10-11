import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";

export interface Source {
  code: string;
  name: string;
  license: string;
  license_url?: string;
  attribution: string;
  source_url?: string;
  scope?: string;
  contact?: string;
  notes?: string;
}

interface Response {
  sources: Source[];
}

/**
 * SourcesDataPage renders the attribution entries for data sources
 * (CREA, BDA-IEO, OFF). It is deliberately server-driven: the DB is the
 * source of truth so changing attribution requires a migration.
 */
export default function SourcesDataPage() {
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
        (s) => s.scope === "data" || s.scope === "data+images",
      ),
    [data],
  );

  if (error) return <p role="alert">{t("sources.loadError")}</p>;
  if (!data) return <p role="status">{t("sources.loading")}</p>;

  return (
    <section>
      <h1 className="mb-3 text-3xl md:text-4xl font-semibold tracking-[-0.03em]">{t("sources.dataPage.title")}</h1>
      <p className="mb-6 text-sm opacity-80">{t("sources.dataPage.intro")}</p>

      {rows.map((s) => (
        <section
          key={s.code}
          className="mb-4 surface p-5 sm:p-6"
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
            {s.source_url && (
              <div>
                <dt className="inline font-medium">{t("sources.common.sourceUrl")}: </dt>
                <dd className="inline">
                  <a href={s.source_url} target="_blank" rel="noopener noreferrer">
                    {s.source_url}
                  </a>
                </dd>
              </div>
            )}
            {s.contact && (
              <div>
                <dt className="inline font-medium">{t("sources.common.contact")}: </dt>
                <dd className="inline">
                  <a href={`mailto:${s.contact}`}>{s.contact}</a>
                </dd>
              </div>
            )}
            {s.notes && <p className="mt-2 opacity-80">{s.notes}</p>}
          </dl>
        </section>
      ))}
    </section>
  );
}
