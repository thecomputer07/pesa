import { useState } from "react";
import { useTranslation } from "react-i18next";

import { patchEmail } from "../../api/me";

export default function EmailChangePanel() {
  const { t } = useTranslation();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [msg, setMsg] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    setMsg(null);
    try {
      await patchEmail({ email, current_password: password });
      setMsg(t("privacy.email.sentToast"));
      setEmail("");
      setPassword("");
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <section aria-labelledby="email-change-title" className="space-y-3">
      <h2 id="email-change-title" className="text-lg font-semibold">
        {t("privacy.email.title")}
      </h2>
      <form onSubmit={onSubmit} className="space-y-2">
        <label className="block text-sm font-medium">
          {t("privacy.email.newEmailLabel")}
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="mt-1 w-full border rounded-md px-2 py-1"
            required
          />
        </label>
        <label className="block text-sm font-medium">
          {t("privacy.email.passwordLabel")}
          <input
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="mt-1 w-full border rounded-md px-2 py-1"
            required
          />
        </label>
        <button
          type="submit"
          disabled={busy}
          className="px-3 py-2 rounded-md bg-[color:var(--color-brand-600)] text-white text-sm disabled:opacity-50"
        >
          {t("privacy.email.submitCta")}
        </button>
        {msg && <p role="status" className="text-sm text-green-700">{msg}</p>}
        {error && <p className="text-sm text-red-600">{error}</p>}
      </form>
    </section>
  );
}
