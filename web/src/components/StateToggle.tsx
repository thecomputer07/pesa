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
 * Required toggle between raw/cooked. When only one state is available,
 * the other button is disabled but still visible to keep context.
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
          "px-3 py-2 text-sm font-medium border",
          active
            ? "bg-[color:var(--color-brand-600)] text-white border-[color:var(--color-brand-600)]"
            : "bg-white text-[color:var(--color-ink)] border-[color:var(--color-border)]",
          !available ? "opacity-50 cursor-not-allowed" : "",
          state === "crudo" ? "rounded-l-md" : "rounded-r-md",
        ].join(" ")}
      >
        {state === "crudo" ? t("detail.stateCrudo") : t("detail.stateCotto")}
      </button>
    );
  };

  return (
    <div className="flex flex-col gap-1">
      <span id={`${groupId}-label`} className="text-sm font-medium">
        {t("detail.stateLabel")}
      </span>
      <div role="radiogroup" aria-labelledby={`${groupId}-label`} className="inline-flex">
        {btn("crudo")}
        {btn("cotto")}
      </div>
    </div>
  );
}
