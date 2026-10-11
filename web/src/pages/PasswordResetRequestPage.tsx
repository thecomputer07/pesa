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
    <section className="max-w-md mx-auto space-y-5 surface p-6 sm:p-8 mt-2 md:mt-6" aria-labelledby="pwreset-title">
      <h1 id="pwreset-title" className="text-2xl md:text-3xl font-semibold tracking-[-0.03em]">
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
              className="mt-1 w-full border border-[color:var(--color-border)] bg-white px-3.5 py-2.5"
            />
          </label>
          <button
            type="submit"
            disabled={submitting}
            className="btn-primary text-sm disabled:opacity-60"
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
    </section>
  );
}
