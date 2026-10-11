import { useState, type FormEvent } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { apiFetch } from "../lib/api";

const MIN_PASSWORD = 10;

export default function PasswordResetConfirmPage() {
  const { t } = useTranslation();
  const [params] = useSearchParams();
  const token = params.get("token") ?? "";
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [done, setDone] = useState(false);

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
    setSubmitting(true);
    try {
      const res = await apiFetch("/auth/password-reset/confirm", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ token, new_password: password }),
      });
      if (!res.ok) {
        let detail: string | undefined;
        try {
          const body = await res.json();
          detail = body.type ?? body.detail;
        } catch {
          /* ignore */
        }
        setError(
          detail === "token_invalid"
            ? t("auth.passwordReset.invalidToken")
            : t("auth.passwordReset.error"),
        );
        return;
      }
      setDone(true);
    } catch {
      setError(t("auth.passwordReset.error"));
    } finally {
      setSubmitting(false);
    }
  }

  if (!token) {
    return (
      <section className="max-w-md mx-auto space-y-3">
        <h1 className="text-2xl md:text-3xl font-semibold tracking-[-0.03em]">{t("auth.passwordReset.confirmTitle")}</h1>
        <p role="alert" className="text-sm text-red-700">
          {t("auth.passwordReset.missingToken")}
        </p>
        <p className="text-sm">
          <Link to="/password-reset/richiedi" className="underline">
            {t("auth.passwordReset.requestTitle")}
          </Link>
        </p>
      </section>
    );
  }

  return (
    <section className="max-w-md mx-auto space-y-5 surface p-6 sm:p-8 mt-2 md:mt-6" aria-labelledby="pwconfirm-title">
      <h1 id="pwconfirm-title" className="text-2xl md:text-3xl font-semibold tracking-[-0.03em]">
        {t("auth.passwordReset.confirmTitle")}
      </h1>
      {done ? (
        <>
          <p role="status" className="text-sm">
            {t("auth.passwordReset.confirmSuccess")}
          </p>
          <p className="text-sm">
            <Link to="/accedi" className="underline">
              {t("auth.register.hasAccount")}
            </Link>
          </p>
        </>
      ) : (
        <form onSubmit={onSubmit} className="space-y-3" noValidate>
          <label className="block">
            <span className="text-sm font-medium">
              {t("auth.passwordReset.newPasswordLabel", { min: MIN_PASSWORD })}
            </span>
            <input
              type="password"
              autoComplete="new-password"
              required
              minLength={MIN_PASSWORD}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="mt-1 w-full border border-[color:var(--color-border)] bg-white px-3.5 py-2.5"
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
              className="mt-1 w-full border border-[color:var(--color-border)] bg-white px-3.5 py-2.5"
            />
          </label>
          {error ? (
            <p role="alert" className="text-sm text-red-700">
              {error}
            </p>
          ) : null}
          <button
            type="submit"
            disabled={submitting}
            className="btn-primary text-sm disabled:opacity-60"
          >
            {t("auth.passwordReset.confirmSubmit")}
          </button>
        </form>
      )}
    </section>
  );
}
