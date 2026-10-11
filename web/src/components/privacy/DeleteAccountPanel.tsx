import { useState } from "react";
import { useTranslation } from "react-i18next";

import { deleteAccount } from "../../api/me";

type Step = "idle" | "password" | "confirm";

export default function DeleteAccountPanel() {
  const { t } = useTranslation();
  const [step, setStep] = useState<Step>("idle");
  const [password, setPassword] = useState("");
  const [confirmText, setConfirmText] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const confirmOk = confirmText === "ELIMINA";

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await deleteAccount({
        current_password: password,
        confirmation_text: confirmText,
      });
      window.location.assign("/?deleted=1");
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <section aria-labelledby="delete-title" className="space-y-3">
      <h2 id="delete-title" className="text-lg font-semibold">
        {t("privacy.delete.title")}
      </h2>
      <p className="text-sm opacity-80">{t("privacy.delete.intro")}</p>
      {step === "idle" && (
        <button
          type="button"
          onClick={() => setStep("password")}
          className="px-3 py-2 rounded-xl border text-sm"
          data-testid="delete-start"
        >
          {t("privacy.delete.startCta")}
        </button>
      )}
      {step === "password" && (
        <form
          onSubmit={(e) => {
            e.preventDefault();
            if (password.length > 0) setStep("confirm");
          }}
          className="space-y-2"
        >
          <label className="block text-sm font-medium">
            {t("privacy.delete.passwordLabel")}
            <input
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="mt-1 w-full border rounded-xl px-2 py-1"
              data-testid="delete-password"
              required
            />
          </label>
          <button
            type="submit"
            disabled={password.length === 0}
            className="px-3 py-2 rounded-xl bg-[color:var(--color-brand-600)] text-white text-sm disabled:opacity-50"
          >
            {t("privacy.delete.continueCta")}
          </button>
        </form>
      )}
      {step === "confirm" && (
        <form onSubmit={onSubmit} className="space-y-2">
          <p className="text-sm">{t("privacy.delete.typeConfirmHint")}</p>
          <label className="block text-sm font-medium">
            {t("privacy.delete.confirmLabel")}
            <input
              type="text"
              value={confirmText}
              onChange={(e) => setConfirmText(e.target.value)}
              autoComplete="off"
              className="mt-1 w-full border rounded-xl px-2 py-1"
              data-testid="delete-confirm-input"
              aria-invalid={!confirmOk}
              required
            />
          </label>
          <button
            type="submit"
            disabled={!confirmOk || busy}
            className="px-3 py-2 rounded-xl bg-red-600 text-white text-sm disabled:opacity-50"
            data-testid="delete-submit"
          >
            {t("privacy.delete.submitCta")}
          </button>
          {error && <p className="text-sm text-red-600">{error}</p>}
        </form>
      )}
    </section>
  );
}
