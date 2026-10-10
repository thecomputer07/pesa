import { lazy, Suspense, useCallback, useEffect, useReducer, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import {
  nextState,
  type PermissionEvent,
  type PermissionState,
} from "../components/scanner/permissionMachine";
import PermissionPrompt from "../components/scanner/PermissionPrompt";
import ManualEntry from "../components/scanner/ManualEntry";

// Lazy-load the camera component so @zxing/browser stays out of the main chunk.
const BarcodeScanner = lazy(() => import("../components/scanner/BarcodeScanner"));

function permissionReducer(
  s: PermissionState,
  e: PermissionEvent,
): PermissionState {
  return nextState(s, e);
}

/**
 * ScanPage is the `/scan` route. It orchestrates:
 *   1. Permission state machine (idle → prompting → granted/denied/unavailable).
 *   2. Lazy-loaded BarcodeScanner when granted.
 *   3. ManualEntry fallback always visible.
 *   4. On scan success: fetch `/api/v1/foods/barcode/:ean` and navigate or toast.
 */
export default function ScanPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [perm, dispatch] = useReducer(permissionReducer, "idle" as PermissionState);
  const [toast, setToast] = useState<string | null>(null);

  const requestCamera = useCallback(async () => {
    dispatch({ type: "REQUEST" });
    if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
      dispatch({ type: "UNAVAILABLE" });
      return;
    }
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ video: true });
      // Immediately release — react-zxing will request its own stream via the
      // ideal=environment constraint once BarcodeScanner mounts.
      stream.getTracks().forEach((tr) => tr.stop());
      dispatch({ type: "GRANTED" });
    } catch (err) {
      const name = (err as DOMException).name;
      if (name === "NotAllowedError" || name === "SecurityError") {
        dispatch({ type: "DENIED" });
      } else {
        dispatch({ type: "UNAVAILABLE" });
      }
    }
  }, []);

  const resolve = useCallback(
    async (ean: string) => {
      try {
        const resp = await fetch(`/api/v1/foods/barcode/${encodeURIComponent(ean)}`);
        if (resp.status === 200) {
          // id > 0: product is in the catalog (already there or just added)
          const body = (await resp.json().catch(() => ({}))) as { id?: number };
          navigate(body.id && body.id > 0 ? `/alimento/${body.id}` : `/alimento/off-${ean}`);
          return;
        }
        if (resp.status === 404) {
          setToast(t("scanner.notFound"));
          return;
        }
        setToast(t("scanner.genericError"));
      } catch {
        setToast(t("scanner.genericError"));
      }
    },
    [navigate, t],
  );

  // Belt-and-braces: reset when the tab is backgrounded so the camera does
  // not auto-resume when the user returns.
  useEffect(() => {
    const onVis = () => {
      if (document.visibilityState === "hidden" && perm === "granted") {
        dispatch({ type: "RESET" });
      }
    };
    document.addEventListener("visibilitychange", onVis);
    return () => document.removeEventListener("visibilitychange", onVis);
  }, [perm]);

  return (
    <section>
      <h1 className="mb-4 text-2xl font-semibold">{t("scanner.pageTitle")}</h1>
      <p className="mb-4 text-sm opacity-80">{t("scanner.pageIntro")}</p>

      <PermissionPrompt
        state={perm}
        onRequest={requestCamera}
        onReset={() => dispatch({ type: "RESET" })}
      />

      {perm === "granted" && (
        <Suspense fallback={<div className="py-8 text-center">…</div>}>
          <BarcodeScanner enabled onResult={resolve} />
        </Suspense>
      )}

      <ManualEntry onSubmit={resolve} />

      {toast && (
        <div role="status" aria-live="polite" className="mt-4 rounded-md bg-yellow-100 p-3 text-sm">
          <p>{toast}</p>
          <a href="/cerca" className="underline">
            {t("scanner.tryTextSearch")}
          </a>
        </div>
      )}
    </section>
  );
}
