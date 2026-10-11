import { useState, type FormEvent } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useAuth } from "../lib/auth/AuthContext";
import { ApiError } from "../lib/api";

export default function LoginPage() {
  const { t } = useTranslation();
  const { login } = useAuth();
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      await login(email.trim(), password);
      const from = params.get("from");
      const safeFrom = from && from.startsWith("/") ? from : "/diario";
      navigate(safeFrom, { replace: true });
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.type === "email_not_verified") {
          setError(t("auth.login.errorNotVerified"));
        } else {
          setError(err.detail ?? t("auth.login.error"));
        }
      } else {
        setError(t("auth.login.error"));
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <section className="max-w-md mx-auto space-y-5 surface p-6 sm:p-8 mt-2 md:mt-6" aria-labelledby="login-title">
      <h1 id="login-title" className="text-2xl md:text-3xl font-semibold tracking-[-0.03em]">
        {t("auth.login.title")}
      </h1>
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
        <label className="block">
          <span className="text-sm font-medium">{t("auth.login.passwordLabel")}</span>
          <input
            type="password"
            autoComplete="current-password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
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
          {t("auth.login.submit")}
        </button>
      </form>
      <p className="text-sm">
        {t("auth.login.noAccount")}{" "}
        <Link to="/registrati" className="underline">
          {t("auth.login.signup")}
        </Link>
      </p>
      <p className="text-sm">
        <Link to="/password-reset/richiedi" className="underline">
          {t("auth.login.forgotPassword")}
        </Link>
      </p>
    </section>
  );
}
