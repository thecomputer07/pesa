import { useId } from "react";
import { useTranslation } from "react-i18next";

export type FoodState = "crudo" | "cotto";

export interface StateToggleProps {
  value: FoodState;
  onChange: (v: FoodState) => void;
  availableStates?: FoodState[];
  disabled?: boolean;
}

/**
 * Segmented raw/cooked switch. When only one state is available, the other
 * option is disabled but still visible to keep context.
 */
export default function StateToggle({
  value,
  onChange,
  availableStates = ["crudo", "cotto"],
  disabled = false,
}: StateToggleProps) {
  const { t } = useTranslation();
  const groupId = useId();

  const btn = (state: FoodState) => {
    const available = availableStates.includes(state);
    const active = value === state;
    return (
      <button
        key={state}
        type="button"
        role="radio"
        aria-checked={active}
        disabled={disabled || !available}
        onClick={() => available && onChange(state)}
        className={[
          "flex-1 px-4 !min-h-10 rounded-xl text-sm font-semibold",
          active
            ? "bg-white text-[color:var(--color-ink)] shadow-[0_0_0_1px_rgb(22_24_40/0.06),var(--shadow-soft)]"
            : "text-[color:var(--color-ink-muted)] hover:text-[color:var(--color-ink)]",
          !available ? "opacity-40" : "",
        ].join(" ")}
      >
        {state === "crudo" ? t("detail.stateCrudo") : t("detail.stateCotto")}
      </button>
    );
  };

  return (
    <div className="flex flex-col gap-3">
      <span id={`${groupId}-label`} className="eyebrow">
        {t("detail.stateLabel")}
      </span>
      <div
        role="radiogroup"
        aria-labelledby={`${groupId}-label`}
        className="flex gap-1 rounded-2xl bg-[color:var(--color-surface-sunken)] p-1"
      >
        {btn("crudo")}
        {btn("cotto")}
      </div>
    </div>
  );
}
