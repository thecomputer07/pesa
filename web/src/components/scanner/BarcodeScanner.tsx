import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useZxing } from "react-zxing";
import { normalizeEAN } from "../../lib/ean";

export interface BarcodeScannerProps {
  onResult: (ean13: string) => void;
  onError?: (err: Error) => void;
  enabled: boolean;
}

/**
 * BarcodeScanner wraps `react-zxing` with Pesa's permission + lifecycle rules:
 *   - rear camera preferred
 *   - torch button hidden unless MediaTrackCapabilities.torch is supported
 *   - camera tracks released on unmount AND on visibilitychange=hidden
 *   - test-only hook `window.__pesaScanResult(ean)` bypasses the decoder
 *
 * This file is the only place that imports `react-zxing`; the parent page
 * lazy-loads it so the main bundle stays slim.
 */
export default function BarcodeScanner({
  onResult,
  onError,
  enabled,
}: BarcodeScannerProps) {
  const { t } = useTranslation();
  const [torchSupported, setTorchSupported] = useState(false);
  const [torchOn, setTorchOn] = useState(false);
  const trackRef = useRef<MediaStreamTrack | null>(null);

  const { ref } = useZxing({
    paused: !enabled,
    constraints: {
      video: { facingMode: { ideal: "environment" } },
      audio: false,
    },
    timeBetweenDecodingAttempts: 100,
    onDecodeResult(result) {
      const raw = result.getText();
      const ean = normalizeEAN(raw);
      if (ean) onResult(ean);
    },
    onError(err) {
      if (onError) onError(err as Error);
    },
  });

  // Track introspection for torch capability detection.
  useEffect(() => {
    const video = ref.current as HTMLVideoElement | null;
    if (!video) return;
    const stream = video.srcObject as MediaStream | null;
    if (!stream) return;
    const track = stream.getVideoTracks()[0];
    trackRef.current = track ?? null;
    if (track && "getCapabilities" in track) {
      try {
        const caps = track.getCapabilities() as { torch?: boolean };
        setTorchSupported(Boolean(caps.torch));
      } catch {
        setTorchSupported(false);
      }
    }
  }, [ref, enabled]);

  // Release the camera indicator when tab is hidden or the component unmounts.
  useEffect(() => {
    const stopTracks = () => {
      const video = ref.current as HTMLVideoElement | null;
      const stream = video?.srcObject as MediaStream | null;
      stream?.getTracks().forEach((tr) => tr.stop());
    };
    const onVis = () => {
      if (document.visibilityState === "hidden") stopTracks();
    };
    document.addEventListener("visibilitychange", onVis);
    return () => {
      document.removeEventListener("visibilitychange", onVis);
      stopTracks();
    };
  }, [ref]);

  // Test hook: only attached when vitest/playwright set MODE=test.
  useEffect(() => {
    if (import.meta.env.MODE !== "test" && import.meta.env.MODE !== "development")
      return;
    (window as unknown as { __pesaScanResult?: (ean: string) => void }).__pesaScanResult =
      (ean: string) => onResult(ean);
    return () => {
      delete (window as unknown as { __pesaScanResult?: (ean: string) => void })
        .__pesaScanResult;
    };
  }, [onResult]);

  const toggleTorch = async () => {
    const track = trackRef.current;
    if (!track || !torchSupported) return;
    try {
      await track.applyConstraints({
        advanced: [{ torch: !torchOn } as MediaTrackConstraintSet],
      });
      setTorchOn((v) => !v);
    } catch {
      /* ignore */
    }
  };

  return (
    <div className="relative">
      <video
        ref={ref as React.RefObject<HTMLVideoElement>}
        className="aspect-video w-full rounded-lg bg-black"
        muted
        playsInline
      />
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-4 rounded-lg border-2 border-white/70"
      />
      {torchSupported && (
        <button
          type="button"
          onClick={toggleTorch}
          className="absolute bottom-3 right-3 rounded-full bg-black/60 px-3 py-1 text-sm text-white"
        >
          {torchOn ? t("scanner.torchOff") : t("scanner.torchOn")}
        </button>
      )}
    </div>
  );
}
