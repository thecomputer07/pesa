import { useMemo, useState } from "react";
import { Link, Navigate, useParams, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { getFood, apiFetch, ApiError } from "../lib/api";
import ConfirmDialog from "../components/ConfirmDialog";
import { removeCustomFood } from "../lib/customFoods";
import { scaleFromPer100g } from "../lib/calc";
import WeightInput from "../components/WeightInput";
import StateToggle, { type FoodState } from "../components/StateToggle";
import NutrientPanel from "../components/NutrientPanel";
import AttributionFooter from "../components/AttributionFooter";
import SourceBadge from "../components/SourceBadge";
import FavoriteToggle from "../components/favorites/FavoriteToggle";
import { useAuth } from "../lib/auth/AuthContext";
import { useFormatGrams } from "../lib/format";
import Icon from "../components/ui/Icon";

function isValidCarb(v: string): boolean {
  const n = Number(v.replace(",", "."));
  return v.trim() !== "" && Number.isFinite(n) && n >= 0 && n <= 100;
}

function stateToFoodState(s: string): FoodState {
  return s === "cooked" || s === "cotto" ? "cotto" : "crudo";
}

export default function FoodDetailPage() {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const { user } = useAuth();
  const { id: idStr } = useParams<{ id: string }>();
  const id = Number(idStr);
  // Parametro speciale "off-{EAN}" = prodotto OFF live, non ancora in DB
  const offEan = idStr?.startsWith("off-") ? idStr.slice(4) : null;

  const [grams, setGrams] = useState<number | null>(100);
  const [savedId, setSavedId] = useState<number | null>(null);
  const [savingErr, setSavingErr] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  // Scanned product without carbohydrates on OFF: value from the label.
  const [labelCarb, setLabelCarb] = useState("");

  const [confirmDelete, setConfirmDelete] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const { data: food, isLoading, isError, error } = useQuery({
    queryKey: offEan ? ["off-barcode", offEan] : ["foods", id],
    queryFn: async ({ signal }) => {
      if (offEan) {
        const res = await apiFetch(`/foods/barcode/${encodeURIComponent(offEan)}`, { signal });
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const body = (await res.json()) as Awaited<ReturnType<typeof getFood>>;
        // In the catalog: load the real food (the render redirects to it).
        if (body.id > 0) return getFood(body.id, { signal });
        return body;
      }
      return getFood(id, { signal });
    },
    enabled: !!offEan || (Number.isFinite(id) && id > 0),
    staleTime: 10 * 60 * 1000,
  });

  const computed = useMemo(() => {
    if (!food) return null;
    const per100 = scaleFromPer100g(food, 100);
    const portion = scaleFromPer100g(food, grams ?? 0);
    return { per100, portion };
  }, [food, grams]);

  if (offEan && food && food.id > 0) {
    return <Navigate to={`/alimento/${food.id}`} replace />;
  }
  if (isLoading) {
    return (
      <p role="status" aria-live="polite" className="py-8 text-center text-sm opacity-70">
        {t("a11y.loading")}
      </p>
    );
  }
  if (isError || !food) {
    const gone = error instanceof ApiError && error.status === 404;
    return (
      <div role="alert" className="surface mx-auto max-w-md p-8 text-center">
        <p className="font-medium">{gone ? "Questo alimento non esiste più." : t("search.error")}</p>
        <Link to="/cerca" className="mt-4 inline-block text-sm">
          {t("nav.search")}
        </Link>
      </div>
    );
  }

  const currentState: FoodState = stateToFoodState(food.state);
  const siblingByState: Record<FoodState, number | undefined> = {
    crudo: undefined,
    cotto: undefined,
  };
  for (const s of food.siblings ?? []) {
    siblingByState[stateToFoodState(s.state)] = s.id;
  }
  siblingByState[currentState] = id;
  const availableStates: FoodState[] = (["crudo", "cotto"] as FoodState[]).filter(
    (s) => siblingByState[s] !== undefined,
  );
  const primaryImage = food.images?.[0];

  return (
    <article className="mx-auto max-w-5xl space-y-6 md:space-y-8">
      <nav aria-label="breadcrumb">
        <button
          type="button"
          onClick={() => navigate(-1)}
          className="inline-flex items-center gap-1.5 -ml-2 px-2 !min-h-9 rounded-full text-sm font-medium text-[color:var(--color-ink-muted)] hover:text-[color:var(--color-ink)] hover:bg-black/[0.04]"
        >
          <Icon name="back" size={16} /> {t("detail.back")}
        </button>
      </nav>

      <header className="flex items-start gap-4 sm:gap-6 rise">
        <span className="relative grid place-items-center w-20 h-20 sm:w-28 sm:h-28 shrink-0 overflow-hidden rounded-3xl bg-[color:var(--color-surface-sunken)] text-3xl font-semibold text-[color:var(--color-ink-faint)] shadow-[0_0_0_1px_rgb(22_24_40/0.06)]">
          <span aria-hidden="true">{food.name_it.slice(0, 1).toUpperCase()}</span>
          {primaryImage ? (
            <img
              src={primaryImage.url}
              alt=""
              loading="lazy"
              onError={(e) => {
                (e.currentTarget as HTMLImageElement).style.display = "none";
              }}
              className="absolute inset-0 w-full h-full object-cover"
            />
          ) : null}
        </span>
        <div className="min-w-0 flex-1 pt-1">
          <h1 className="text-2xl sm:text-4xl font-semibold tracking-[-0.03em] leading-tight">{food.name_it}</h1>
          <div className="mt-3 flex flex-wrap items-center gap-2 text-sm text-[color:var(--color-ink-muted)]">
            <SourceBadge
              code={attrCode(food.attribution)}
              license={food.license}
              licenseUrl={food.license_url}
            />
            {food.category ? <span>{food.category}</span> : null}
          </div>
        </div>
        {!offEan ? <FavoriteToggle foodId={id} /> : null}
      </header>

      {food.is_mine ? (
        <div className="flex flex-wrap items-center gap-2 -mt-2">
          <span className="text-xs text-[color:var(--color-ink-faint)] mr-1">Alimento creato da te</span>
          <Link to={`/alimento/${id}/modifica`} className="btn-quiet !py-1.5 !min-h-9 text-sm">
            Modifica
          </Link>
          <button
            type="button"
            onClick={() => setConfirmDelete(true)}
            className="px-3 !min-h-9 rounded-[var(--radius-control)] text-sm font-medium text-red-700 border border-red-200 bg-white hover:bg-red-50"
          >
            Elimina
          </button>
          {deleteError ? (
            <p role="alert" className="basis-full text-sm text-red-700">
              {deleteError}
            </p>
          ) : null}
          <ConfirmDialog
            open={confirmDelete}
            title="Eliminare l'alimento?"
            message={`"${food.name_it}" verrà eliminato. L'azione è irreversibile.`}
            confirmLabel="Elimina"
            variant="danger"
            onConfirm={async () => {
              setConfirmDelete(false);
              setDeleteError(null);
              const err = await removeCustomFood(id);
              if (err) setDeleteError(err);
              else navigate("/miei-alimenti", { replace: true });
            }}
            onCancel={() => setConfirmDelete(false)}
          />
        </div>
      ) : null}

      {offEan && user ? (
        <section className="surface p-5 text-sm">
          <p className="mb-2">
            Prodotto letto da Open Food Facts via codice a barre{" "}
            <span className="font-mono text-xs">{offEan}</span>. Per aggiungerlo ai tuoi pasti
            salvalo nei tuoi alimenti.
          </p>
          {food.carb_g == null ? (
            <label className="mb-2 block">
              <span className="block font-medium text-red-700">
                Open Food Facts non riporta i carboidrati di questo prodotto: inseriscili
                dall'etichetta (g per 100 g).
              </span>
              <input
                type="number"
                inputMode="decimal"
                min={0}
                max={100}
                step="0.1"
                value={labelCarb}
                onChange={(e) => setLabelCarb(e.target.value)}
                className="mt-1 w-28 px-3 py-2 border border-[color:var(--color-border)] bg-white"
              />
            </label>
          ) : null}
          {savedId ? (
            <p role="status" className="text-green-700">
              Salvato.{" "}
              <Link to={`/alimento/${savedId}`} className="underline">
                Apri
              </Link>
            </p>
          ) : (
            <>
              <button
                type="button"
                disabled={saving || (food.carb_g == null && !isValidCarb(labelCarb))}
                onClick={async () => {
                  setSavingErr(null);
                  setSaving(true);
                  try {
                    const res = await apiFetch(`/foods/custom`, {
                      method: "POST",
                      headers: { "Content-Type": "application/json" },
                      body: JSON.stringify({
                        name_it: food.name_it,
                        category: food.category,
                        state: "raw",
                        edible_pct: food.edible_pct,
                        kcal: food.kcal,
                        // never default to 0: a missing value comes from the label
                        carb_g: food.carb_g ?? Number(labelCarb.replace(",", ".")),
                        protein_g: food.protein_g,
                        fat_g: food.fat_g,
                        sugar_g: food.sugar_g,
                        fiber_g: food.fiber_g,
                      }),
                    });
                    if (!res.ok) {
                      const txt = await res.text().catch(() => "");
                      setSavingErr(txt || `Errore ${res.status}`);
                      return;
                    }
                    const data = (await res.json()) as { id: number };
                    setSavedId(data.id);
                  } catch (err) {
                    setSavingErr(err instanceof Error ? err.message : "Errore imprevisto");
                  } finally {
                    setSaving(false);
                  }
                }}
                className="btn-primary"
              >
                {saving ? "Salvataggio…" : "Salva nei miei alimenti"}
              </button>
              {savingErr ? (
                <p role="alert" className="mt-2 text-red-700">
                  {savingErr}
                </p>
              ) : null}
            </>
          )}
        </section>
      ) : null}

      <section className="grid gap-4 lg:grid-cols-[1fr_1.1fr] rise rise-2">
        <div className="surface p-5 sm:p-6 flex flex-col gap-7">
          <WeightInput value={grams} onChange={setGrams} autoFocus presets={[50, 80, 100, 150, 200]} />
          {/* Toggle only when a raw/cooked pair exists (see migration 00026). */}
          {availableStates.length > 1 ? (
            <StateToggle
              value={currentState}
              onChange={(next) => {
                const targetId = siblingByState[next];
                if (targetId && targetId !== id) navigate(`/alimento/${targetId}`);
              }}
              availableStates={availableStates}
            />
          ) : null}
        </div>

        {computed ? (
          <CarbDisplay
            grams={grams ?? 0}
            carb={computed.portion.values.carb_g}
            carbPer100={computed.per100.values.carb_g}
          />
        ) : null}
      </section>

      {computed ? (
        <div className="rise rise-3">
          <NutrientPanel grams={grams ?? 0} portion={computed.portion} per100={computed.per100} />
        </div>
      ) : null}

      <AttributionFooter food={food} lang={i18n.language} />
    </article>
  );
}

/** The answer, read out like the display of a kitchen scale. */
function CarbDisplay({ grams, carb, carbPer100 }: { grams: number; carb?: number; carbPer100?: number }) {
  const { t, i18n } = useTranslation();
  const fmtG = useFormatGrams();
  const value =
    carb == null || !Number.isFinite(carb)
      ? "—"
      : new Intl.NumberFormat(i18n.language.startsWith("en") ? "en-GB" : "it-IT", { maximumFractionDigits: 1 }).format(carb);
  const share = carbPer100 != null && Number.isFinite(carbPer100) ? Math.min(100, Math.max(0, carbPer100)) : 0;

  return (
    <div className="relative overflow-hidden rounded-[var(--radius-card)] bg-[color:var(--color-ink)] p-6 sm:p-8 text-white shadow-[0_30px_60px_-30px_rgb(20_30_80/0.55)] flex flex-col justify-between min-h-[16rem]">
      <div aria-hidden="true" className="absolute inset-0 bg-[radial-gradient(120%_80%_at_100%_0%,rgb(49_87_224/0.45),transparent_60%)]" />
      <p className="relative text-xs font-medium tracking-wide text-white/55">{t("detail.carb")}</p>
      <div className="relative">
        <p className="flex items-baseline gap-2">
          <span className="tabular text-[4.25rem] sm:text-[5.5rem] leading-none font-semibold tracking-[-0.05em] text-[color:var(--color-carb-100)]">
            {value}
          </span>
          <span className="text-2xl font-medium text-[color:var(--color-carb-100)]/80">g</span>
        </p>
        <p className="mt-2 text-white/70 tabular">{t("detail.resultFor", { grams: new Intl.NumberFormat(i18n.language).format(grams) })}</p>
        <div className="mt-7 h-2 rounded-full bg-white/10 overflow-hidden" aria-hidden="true">
          <div
            className="h-full rounded-full bg-gradient-to-r from-[color:var(--color-carb-500)] to-[color:var(--color-carb-100)] transition-[width] duration-500"
            style={{ width: `${share}%` }}
          />
        </div>
        <p className="mt-3 text-xs text-white/50 tabular">
          {fmtG(carbPer100)} / {t("detail.per100Col")}
        </p>
      </div>
    </div>
  );
}

function attrCode(attribution: string): string {
  const a = attribution.toUpperCase();
  if (a.includes("CREA")) return "CREA";
  if (a.includes("CIQUAL")) return "CIQUAL";
  if (a.includes("AGRICULTURE") || a.includes("FOODDATA")) return "USDA";
  if (a.includes("BDA") || a.includes("IEO")) return "BDA";
  if (a.includes("OPEN FOOD")) return "OFF";
  return "—";
}
