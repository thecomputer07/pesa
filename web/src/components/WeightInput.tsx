import { useEffect, useId, useState } from "react";
import { useTranslation } from "react-i18next";
import { parseDecimal } from "../lib/format";

export interface WeightInputProps {
  value: number | null;
  onChange: (grams: number | null) => void;
  max?: number;
  autoFocus?: boolean;
}

/**
 * Numeric input for grams. Accetta virgola italiana ("120,5") e punto ("120.5").
 * Stato stringa locale: permette stati intermedi ("100,", "0.") senza che il
 * parent "pulisca" l'input. Emette il numero solo quando il parse ha successo.
 */
export default function WeightInput({ value, onChange, max = 10_000, autoFocus }: WeightInputProps) {
  const { t } = useTranslation();
  const id = useId();
  const [text, setText] = useState<string>(value == null ? "" : String(value).replace(".", ","));

  // Se il parent cambia value (es. reset), ri-sincronizza il testo.
  useEffect(() => {
    const parsed = parseDecimal(text);
    if (parsed !== value) {
      setText(value == null ? "" : String(value).replace(".", ","));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value]);

  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={id} className="text-sm font-medium">
        {t("detail.weightLabel")}
      </label>
      <div className="flex items-center gap-2">
        <input
          id={id}
          type="text"
          inputMode="decimal"
          pattern="[0-9]+([,.][0-9]+)?"
          value={text}
          autoFocus={autoFocus}
          onChange={(e) => {
            const raw = e.target.value;
            // Consenti solo cifre, virgola, punto (e vuoto)
            if (raw !== "" && !/^[0-9]*[.,]?[0-9]*$/.test(raw)) return;
            setText(raw);
            if (raw === "") {
              onChange(null);
              return;
            }
            const n = parseDecimal(raw);
            if (n == null) return; // stato intermedio "100," → non emettere
            if (n < 0 || n > max) return;
            onChange(n);
          }}
          aria-describedby={`${id}-help`}
          className="w-32 px-3 py-2 rounded-md border border-[color:var(--color-border)] bg-white text-right"
        />
        <span aria-hidden="true" className="text-sm text-[color:var(--color-ink-muted)]">
          g
        </span>
      </div>
      <p id={`${id}-help`} className="text-xs text-[color:var(--color-ink-muted)]">
        {t("detail.weightHelp")}
      </p>
    </div>
  );
}
