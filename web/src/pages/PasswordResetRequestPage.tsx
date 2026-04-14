import { useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { apiFetch } from "../lib/api";

export default function PasswordResetRequestPage() {
  const { t } = useTranslation();
  const [email, setEmail] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [done, setDone] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    try {
      await apiFetch("/auth/password-reset/request", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: email.trim() }),
      });
    } catch {
      /* swallow — enumeration-safe: always show same success */
    } finally {
      setSubmitting(false);
      setDone(true);
    }
  }

  return (
    <main className="max-w-md mx-auto space-y-4" aria-labelledby="pwreset-title">
      <h1 id="pwreset-title" className="text-2xl font-bold">
        {t("auth.passwordReset.requestTitle")}
      </h1>
      {done ? (
        <p role="status" className="text-sm">
          {t("auth.passwordReset.requestAck")}
        </p>
      ) : (
        <form onSubmit={onSubmit} className="space-y-3" noValidate>
          <label className="block">
            <span className="text-sm font-medium">{t("auth.login.emailLabel")}</span>
            <input
              type="email"
              autoComplete="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="mt-1 w-full border border-[color:var(--color-border)] rounded-md px-2 py-1.5"
            />
          </label>
          <button
            type="submit"
            disabled={submitting}
            className="px-3 py-1.5 text-sm bg-[color:var(--color-brand-600)] text-white rounded-md disabled:opacity-60"
          >
            {t("auth.passwordReset.requestSubmit")}
          </button>
        </form>
      )}
      <p className="text-sm">
        <Link to="/accedi" className="underline">
          {t("auth.register.hasAccount")}
        </Link>
      </p>
    </main>
  );
}
