import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { setUpdateHandler } from "../lib/sw-update";

interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>;
}

/**
 * Dual-purpose prompt: (a) "add to home screen" install banner, (b) SW update
 * available banner. Both are dismissible and politely wired.
 */
export default function InstallPrompt() {
  const { t } = useTranslation();
  const [installEvt, setInstallEvt] = useState<BeforeInstallPromptEvent | null>(null);
  const [acceptUpdate, setAcceptUpdate] = useState<(() => void) | null>(null);

  useEffect(() => {
    const onBeforeInstall = (e: Event) => {
      e.preventDefault();
      setInstallEvt(e as BeforeInstallPromptEvent);
    };
    window.addEventListener("beforeinstallprompt", onBeforeInstall);

    setUpdateHandler((accept) => setAcceptUpdate(() => accept));

    return () => window.removeEventListener("beforeinstallprompt", onBeforeInstall);
  }, []);

  if (acceptUpdate) {
    return (
      <div
        role="alertdialog"
        aria-labelledby="pesa-update-title"
        className="fixed bottom-4 inset-x-4 md:inset-x-auto md:right-4 md:w-80 bg-white border border-[color:var(--color-border)] rounded-lg shadow-lg p-4 z-50"
      >
        <p id="pesa-update-title" className="font-medium">
          {t("install.updateAvailable")}
        </p>
        <div className="mt-3 flex gap-2 justify-end">
          <button
            type="button"
            onClick={() => setAcceptUpdate(null)}
            className="px-3 py-1.5 text-sm rounded border border-[color:var(--color-border)]"
          >
            {t("install.updateLater")}
          </button>
          <button
            type="button"
            onClick={() => {
              acceptUpdate();
              setAcceptUpdate(null);
            }}
            className="px-3 py-1.5 text-sm rounded bg-[color:var(--color-brand-600)] text-white"
          >
            {t("install.updateNow")}
          </button>
        </div>
      </div>
    );
  }

  if (installEvt) {
    return (
      <div
        role="dialog"
        aria-labelledby="pesa-install-title"
        className="fixed bottom-4 inset-x-4 md:inset-x-auto md:right-4 md:w-80 bg-white border border-[color:var(--color-border)] rounded-lg shadow-lg p-4 z-50"
      >
        <p id="pesa-install-title" className="font-medium">
          {t("install.prompt")}
        </p>
        <div className="mt-3 flex gap-2 justify-end">
          <button
            type="button"
            onClick={() => setInstallEvt(null)}
            className="px-3 py-1.5 text-sm rounded border border-[color:var(--color-border)]"
          >
            {t("install.notNow")}
          </button>
          <button
            type="button"
            onClick={async () => {
              await installEvt.prompt();
              await installEvt.userChoice;
              setInstallEvt(null);
            }}
            className="px-3 py-1.5 text-sm rounded bg-[color:var(--color-brand-600)] text-white"
          >
            {t("install.install")}
          </button>
        </div>
      </div>
    );
  }

  return null;
}
