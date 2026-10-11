import { useEffect, useId, useState } from "react";
import { useTranslation } from "react-i18next";
import { parseDecimal } from "../lib/format";
import Icon from "./ui/Icon";

export interface WeightInputProps {
  value: number | null;
  onChange: (grams: number | null) => void;
  max?: number;
  autoFocus?: boolean;
  /** Large field with ±10 g steppers and quick-portion chips (food detail). */
  presets?: number[];
}

/**
 * Numeric input for grams. Accetta virgola italiana ("120,5") e punto ("120.5").
 * Stato stringa locale: permette stati intermedi ("100,", "0.") senza che il
 * parent "pulisca" l'input. Emette il numero solo quando il parse ha successo.
 */
export default function WeightInput({ value, onChange, max = 10_000, autoFocus, presets }: WeightInputProps) {
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

  const big = presets != null;
  const step = (delta: number) => {
    const next = Math.min(max, Math.max(0, Math.round(((value ?? 0) + delta) * 10) / 10));
    onChange(next);
  };

  const input = (
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
      onFocus={big ? (e) => e.currentTarget.select() : undefined}
      aria-describedby={`${id}-help`}
      className={
        big
          ? "tabular w-full min-w-0 bg-transparent border-0 !shadow-none text-center text-4xl font-semibold tracking-[-0.03em] py-2 outline-none"
          : "w-32 px-3 py-2 border border-[color:var(--color-border)] bg-white text-right tabular"
      }
    />
  );

  if (!big) {
    return (
      <div className="flex flex-col gap-1">
        <label htmlFor={id} className="text-sm font-medium">
          {t("detail.weightLabel")}
        </label>
        <div className="flex items-center gap-2">
          {input}
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

  const stepBtn =
    "grid place-items-center w-12 h-12 shrink-0 rounded-2xl bg-white text-[color:var(--color-ink)] shadow-[0_0_0_1px_var(--color-border),var(--shadow-soft)] hover:shadow-[0_0_0_1px_var(--color-border-strong),var(--shadow-lift)]";

  return (
    <div className="flex flex-col gap-3">
      <label htmlFor={id} className="eyebrow">
        {t("detail.weightLabel")}
      </label>
      <div className="flex items-center gap-3">
        <button type="button" className={stepBtn} onClick={() => step(-10)} aria-label={t("detail.decrease")}>
          <Icon name="minus" size={20} />
        </button>
        <div className="flex-1 flex items-baseline justify-center rounded-2xl bg-[color:var(--color-canvas)] px-3 focus-within:bg-white focus-within:shadow-[0_0_0_1px_var(--color-brand-500),0_0_0_5px_rgb(49_87_224/0.14)] transition-shadow">
          {input}
          <span aria-hidden="true" className="-ml-1 pr-2 text-xl font-medium text-[color:var(--color-ink-faint)]">
            g
          </span>
        </div>
        <button type="button" className={stepBtn} onClick={() => step(10)} aria-label={t("detail.increase")}>
          <Icon name="plus" size={20} />
        </button>
      </div>
      <p id={`${id}-help`} className="text-xs text-[color:var(--color-ink-faint)] text-center">
        {t("detail.weightHelp")}
      </p>
      <div role="group" aria-label={t("detail.quickWeights")} className="flex flex-wrap justify-center gap-2">
        {presets.map((g) => (
          <button
            key={g}
            type="button"
            aria-pressed={value === g}
            onClick={() => onChange(g)}
            className="chip tabular !min-h-9"
          >
            {g} g
          </button>
        ))}
      </div>
    </div>
  );
}
