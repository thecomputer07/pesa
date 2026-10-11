import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { useTranslation } from "react-i18next";

type Status = "verifying" | "ok" | "invalid" | "missing";

// Module-level cache keyed by token. React 19 StrictMode runs the effect twice
// in dev; the single-use token must be consumed exactly once, and both runs
// must receive the same result. We cache an in-flight promise per token.
const pending = new Map<string, Promise<Status>>();

async function verifyOnce(token: string): Promise<Status> {
  const existing = pending.get(token);
  if (existing) return existing;
  const p = (async (): Promise<Status> => {
    try {
      const res = await fetch(`/verify-email?token=${encodeURIComponent(token)}`, {
        method: "GET",
        redirect: "manual",
        credentials: "include",
      });
      if (res.type === "opaqueredirect" || res.status === 0 || (res.status >= 200 && res.status < 400)) {
        return "ok";
      }
      return "invalid";
    } catch {
      return "invalid";
    }
  })();
  pending.set(token, p);
  return p;
}

export default function VerifyEmailPage() {
  const { t } = useTranslation();
  const [params] = useSearchParams();
  const token = params.get("token") ?? "";
  const [status, setStatus] = useState<Status>(token ? "verifying" : "missing");

  useEffect(() => {
    if (!token) return;
    let cancelled = false;
    verifyOnce(token).then((result) => {
      if (!cancelled) setStatus(result);
    });
    return () => {
      cancelled = true;
    };
  }, [token]);

  return (
    <section className="max-w-md mx-auto space-y-5 surface p-6 sm:p-8 mt-2 md:mt-6" aria-labelledby="verify-title">
      <h1 id="verify-title" className="text-2xl md:text-3xl font-semibold tracking-[-0.03em]">
        {t("auth.verifyEmail.title")}
      </h1>
      {status === "verifying" ? (
        <p role="status" className="text-sm">
          {t("auth.verifyEmail.verifying")}
        </p>
      ) : null}
      {status === "ok" ? (
        <>
          <p role="status" className="text-sm">
            {t("auth.verifyEmail.ok")}
          </p>
          <p className="text-sm">
            <Link to="/accedi" className="underline">
              {t("auth.register.hasAccount")}
            </Link>
          </p>
        </>
      ) : null}
      {status === "invalid" ? (
        <p role="alert" className="text-sm text-red-700">
          {t("auth.verifyEmail.invalid")}
        </p>
      ) : null}
      {status === "missing" ? (
        <p role="alert" className="text-sm text-red-700">
          {t("auth.verifyEmail.missing")}
        </p>
      ) : null}
    </section>
  );
}
