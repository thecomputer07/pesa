import { useTranslation } from "react-i18next";
import type { PermissionState } from "./permissionMachine";

export interface PermissionPromptProps {
  state: PermissionState;
  onRequest: () => void;
  onReset: () => void;
}

/**
 * PermissionPrompt renders a user-gesture button for requesting camera access
 * and copy covering the denied/unavailable dead-ends. It is the only entry
 * point to the camera (iOS Safari standalone requires an explicit gesture).
 */
export default function PermissionPrompt({
  state,
  onRequest,
  onReset,
}: PermissionPromptProps) {
  const { t } = useTranslation();

  if (state === "granted") return null;

  return (
    <div className="rounded-lg border border-[color:var(--color-border)] p-4 text-sm">
      {state === "idle" && (
        <>
          <p>{t("scanner.prompt")}</p>
          <button
            type="button"
            onClick={onRequest}
            className="mt-3 rounded-md bg-[color:var(--color-brand-600)] px-4 py-2 text-white"
          >
            {t("scanner.promptCta")}
          </button>
        </>
      )}
      {state === "prompting" && <p>{t("scanner.prompting")}</p>}
      {state === "denied" && (
        <>
          <p>{t("scanner.denied")}</p>
          <button
            type="button"
            onClick={onReset}
            className="mt-3 rounded-md border border-[color:var(--color-border)] px-4 py-2"
          >
            {t("scanner.retry")}
          </button>
        </>
      )}
      {state === "unavailable" && <p>{t("scanner.unavailable")}</p>}
    </div>
  );
}
