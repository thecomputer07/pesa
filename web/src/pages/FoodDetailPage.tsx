import { useMemo, useState } from "react";
import { Link, Navigate, useParams, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { getFood, apiFetch } from "../lib/api";
import { scaleFromPer100g } from "../lib/calc";
import WeightInput from "../components/WeightInput";
import StateToggle, { type FoodState } from "../components/StateToggle";
import NutrientPanel from "../components/NutrientPanel";
import AttributionFooter from "../components/AttributionFooter";
import SourceBadge from "../components/SourceBadge";
import FavoriteToggle from "../components/favorites/FavoriteToggle";
import { useAuth } from "../lib/auth/AuthContext";

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

  const { data: food, isLoading, isError } = useQuery({
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
    return (
      <p role="alert" className="py-8 text-center text-sm text-red-700">
        {t("search.error")}
      </p>
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
    <article className="space-y-6">
      <nav aria-label="breadcrumb">
        <button
          type="button"
          onClick={() => navigate(-1)}
          className="text-sm underline bg-transparent border-0 p-0 cursor-pointer"
        >
          ← {t("detail.back")}
        </button>
      </nav>

      <header className="flex items-start justify-between gap-3">
        <div className="flex items-start gap-4">
          {primaryImage ? (
            <img
              src={primaryImage.url}
              alt=""
              loading="lazy"
              onError={(e) => {
                (e.currentTarget as HTMLImageElement).style.display = "none";
              }}
              className="w-24 h-24 md:w-32 md:h-32 object-cover rounded-md border border-[color:var(--color-border)] bg-white"
            />
          ) : null}
          <div>
            <h1 className="text-2xl md:text-3xl font-bold">{food.name_it}</h1>
            {food.category ? (
              <p className="text-sm text-[color:var(--color-ink-muted)] mt-1">{food.category}</p>
            ) : null}
          </div>
        </div>
        <div className="flex items-center gap-2">
          {!offEan ? <FavoriteToggle foodId={id} /> : null}
          <SourceBadge
            code={attrCode(food.attribution)}
            license={food.license}
            licenseUrl={food.license_url}
          />
        </div>
      </header>

      {offEan && user ? (
        <section className="rounded-md border border-[color:var(--color-border)] bg-[color:var(--color-brand-50)] p-3 text-sm">
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
                className="mt-1 w-28 px-2 py-1 border border-[color:var(--color-border)] rounded-md"
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
                className="px-3 py-1.5 bg-[color:var(--color-brand-600)] text-white rounded-md"
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

      <section className="flex flex-wrap items-end gap-6">
        <WeightInput value={grams} onChange={setGrams} autoFocus />
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
      </section>

      {computed ? (
        <NutrientPanel
          grams={grams ?? 0}
          portion={computed.portion}
          per100={computed.per100}
        />
      ) : null}

      <AttributionFooter food={food} lang={i18n.language} />
    </article>
  );
}

function attrCode(attribution: string): string {
  const a = attribution.toUpperCase();
  if (a.includes("CREA")) return "CREA";
  if (a.includes("CIQUAL")) return "CIQUAL";
  if (a.includes("BDA") || a.includes("IEO")) return "BDA";
  if (a.includes("OPEN FOOD")) return "OFF";
  return "—";
}
