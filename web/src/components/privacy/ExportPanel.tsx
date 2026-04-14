import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";

import { getExportJob, requestExport, type ExportJob } from "../../api/me";

const POLL_MS = 3000;

export default function ExportPanel() {
  const { t } = useTranslation();
  const [jobId, setJobId] = useState<number | null>(null);
  const [job, setJob] = useState<ExportJob | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const timerRef = useRef<number | null>(null);

  useEffect(() => () => {
    if (timerRef.current) window.clearInterval(timerRef.current);
  }, []);

  async function onRequest() {
    setLoading(true);
    setError(null);
    try {
      const res = await requestExport();
      setJobId(res.job_id);
      if (timerRef.current) window.clearInterval(timerRef.current);
      timerRef.current = window.setInterval(() => {
        void poll(res.job_id);
      }, POLL_MS);
      void poll(res.job_id);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  }

  async function poll(id: number) {
    try {
      const j = await getExportJob(id);
      setJob(j);
      if (j.status === "ready" || j.status === "failed" || j.status === "expired") {
        if (timerRef.current) {
          window.clearInterval(timerRef.current);
          timerRef.current = null;
        }
      }
    } catch (e) {
      setError((e as Error).message);
    }
  }

  const pending = job?.status === "pending" || job?.status === "running";
  return (
    <section aria-labelledby="export-title" className="space-y-3">
      <h2 id="export-title" className="text-lg font-semibold">
        {t("privacy.export.title")}
      </h2>
      <p className="text-sm opacity-80">{t("privacy.export.intro")}</p>
      <button
        type="button"
        onClick={onRequest}
        disabled={loading || pending}
        className="px-3 py-2 rounded-md bg-[color:var(--color-brand-600)] text-white text-sm disabled:opacity-50"
      >
        {t("privacy.export.requestCta")}
      </button>
      {job && (
        <p role="status" aria-live="polite" className="text-sm">
          {t(`privacy.export.status.${job.status}`)}
          {job.bytes !== undefined && ` (${Math.round(job.bytes / 1024)} KB)`}
        </p>
      )}
      {job?.status === "ready" && job.download_url && (
        <a
          href={job.download_url}
          target="_blank"
          rel="noreferrer"
          className="inline-block px-3 py-2 rounded-md border text-sm"
          data-testid="download-export"
        >
          {t("privacy.export.downloadCta")}
        </a>
      )}
      {error && <p className="text-sm text-red-600">{error}</p>}
      {jobId !== null && (
        <p className="text-xs opacity-60">
          {t("privacy.export.jobIdLabel")}: {jobId}
        </p>
      )}
    </section>
  );
}
