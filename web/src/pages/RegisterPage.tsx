import { useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { Trans, useTranslation } from "react-i18next";
import { useAuth } from "../lib/auth/AuthContext";
import { ApiError } from "../lib/api";

const MIN_PASSWORD = 10;

export default function RegisterPage() {
  const { t } = useTranslation();
  const { register } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [consent, setConsent] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    if (password.length < MIN_PASSWORD) {
      setError(t("auth.register.weakPassword", { min: MIN_PASSWORD }));
      return;
    }
    if (password !== confirm) {
      setError(t("auth.register.mismatch"));
      return;
    }
    if (!consent) {
      setError(t("auth.register.consentRequired"));
      return;
    }
    setSubmitting(true);
    try {
      await register(email.trim(), password, consent);
      setSuccess(true);
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.type === "email_already_used") setError(t("auth.register.emailUsed"));
        else if (err.type === "password_too_short")
          setError(t("auth.register.weakPassword", { min: MIN_PASSWORD }));
        else setError(err.detail ?? t("auth.register.error"));
      } else {
        setError(t("auth.register.error"));
      }
    } finally {
      setSubmitting(false);
    }
  }

  if (success) {
    return (
      <main className="max-w-md mx-auto space-y-4" aria-labelledby="register-title">
        <h1 id="register-title" className="text-2xl font-bold">
          {t("auth.register.title")}
        </h1>
        <p role="status" className="text-sm">
          {t("auth.register.success")}
        </p>
        <p className="text-xs text-[color:var(--color-ink-muted)]">
          {t("auth.register.devHint")}
        </p>
        <p className="text-sm">
          <Link to="/accedi" className="underline">
            {t("auth.register.hasAccount")}
          </Link>
        </p>
      </main>
    );
  }

  return (
    <main className="max-w-md mx-auto space-y-4" aria-labelledby="register-title">
      <h1 id="register-title" className="text-2xl font-bold">
        {t("auth.register.title")}
      </h1>
      <form onSubmit={onSubmit} className="space-y-3" noValidate>
        <label className="block">
          <span className="text-sm font-medium">{t("auth.register.emailLabel")}</span>
          <input
            type="email"
            autoComplete="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="mt-1 w-full border border-[color:var(--color-border)] rounded-md px-2 py-1.5"
          />
        </label>
        <label className="block">
          <span className="text-sm font-medium">
            {t("auth.register.passwordLabel", { min: MIN_PASSWORD })}
          </span>
          <input
            type="password"
            autoComplete="new-password"
            required
            minLength={MIN_PASSWORD}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="mt-1 w-full border border-[color:var(--color-border)] rounded-md px-2 py-1.5"
          />
        </label>
        <label className="block">
          <span className="text-sm font-medium">{t("auth.register.confirmLabel")}</span>
          <input
            type="password"
            autoComplete="new-password"
            required
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
            className="mt-1 w-full border border-[color:var(--color-border)] rounded-md px-2 py-1.5"
          />
        </label>
        <label className="flex items-start gap-2 text-sm">
          <input
            type="checkbox"
            required
            checked={consent}
            onChange={(e) => setConsent(e.target.checked)}
            className="mt-1"
          />
          <span>
            <Trans
              i18nKey="auth.register.privacyConsent"
              components={{
                scope: <Link to="/scopo" className="underline" />,
                privacy: <Link to="/privacy" className="underline" />,
              }}
            />
          </span>
        </label>
        {error ? (
          <p role="alert" className="text-sm text-red-700">
            {error}
          </p>
        ) : null}
        <button
          type="submit"
          disabled={submitting}
          className="px-3 py-1.5 text-sm bg-[color:var(--color-brand-600)] text-white rounded-md disabled:opacity-60"
        >
          {t("auth.register.submit")}
        </button>
      </form>
      <p className="text-sm">
        <Link to="/accedi" className="underline">
          {t("auth.register.hasAccount")}
        </Link>
      </p>
    </main>
  );
}
