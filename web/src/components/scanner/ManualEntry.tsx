import { useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { normalizeEAN } from "../../lib/ean";

export interface ManualEntryProps {
  onSubmit: (ean13: string) => void;
  autoFocus?: boolean;
}

/** ManualEntry is the keyboard fallback for the camera scanner. */
export default function ManualEntry({ onSubmit, autoFocus }: ManualEntryProps) {
  const { t } = useTranslation();
  const [value, setValue] = useState("");
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    const ean = normalizeEAN(value);
    if (!ean) {
      setError(t("scanner.manualError"));
      return;
    }
    setError(null);
    onSubmit(ean);
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="mt-4 flex flex-col gap-2"
      aria-label={t("scanner.manualLabel")}
    >
      <label htmlFor="scanner-manual-ean" className="text-sm font-medium">
        {t("scanner.manualLabel")}
      </label>
      <div className="flex gap-2">
        <input
          id="scanner-manual-ean"
          type="text"
          inputMode="numeric"
          pattern="\d*"
          autoComplete="off"
          autoFocus={autoFocus}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          className="flex-1 rounded-md border border-[color:var(--color-border)] bg-transparent px-3 py-2"
          aria-invalid={error ? "true" : "false"}
          aria-describedby={error ? "scanner-manual-err" : undefined}
        />
        <button
          type="submit"
          className="rounded-md bg-[color:var(--color-brand-600)] px-4 py-2 text-white"
        >
          {t("scanner.manualSubmit")}
        </button>
      </div>
      {error && (
        <p id="scanner-manual-err" role="alert" className="text-sm text-red-600">
          {error}
        </p>
      )}
    </form>
  );
}
