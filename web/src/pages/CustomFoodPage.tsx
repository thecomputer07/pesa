import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { apiFetch } from "../lib/api";
import ConfirmDialog from "../components/ConfirmDialog";
import { removeCustomFood } from "../lib/customFoods";

type Form = {
  name_it: string;
  category: string;
  state: "raw" | "cooked";
  edible_pct: string;
  kcal: string;
  carb_g: string;
  protein_g: string;
  fat_g: string;
  sugar_g: string;
  fiber_g: string;
  salt_g: string;
};

const empty: Form = {
  name_it: "",
  category: "",
  state: "raw",
  edible_pct: "100",
  kcal: "",
  carb_g: "",
  protein_g: "",
  fat_g: "",
  sugar_g: "",
  fiber_g: "",
  salt_g: "",
};

function parseNum(s: string): number | undefined {
  const v = s.replace(",", ".").trim();
  if (v === "") return undefined;
  const n = Number(v);
  return Number.isFinite(n) ? n : undefined;
}

export default function CustomFoodPage() {
  const navigate = useNavigate();
  const { id: editId } = useParams<{ id: string }>();
  const isEdit = !!editId;
  const [form, setForm] = useState<Form>(empty);
  const [submitting, setSubmitting] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [loading, setLoading] = useState(isEdit);
  const [error, setError] = useState<string | null>(null);

  function update<K extends keyof Form>(key: K, value: Form[K]) {
    setForm((f) => ({ ...f, [key]: value }));
  }

  // In edit mode, carica valori esistenti.
  useEffect(() => {
    if (!isEdit || !editId) return;
    let alive = true;
    void (async () => {
      try {
        const res = await apiFetch(`/foods/${editId}`);
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const d = await res.json();
        if (!alive) return;
        setForm({
          name_it: d.name_it ?? "",
          category: d.category ?? "",
          state: d.state === "cooked" ? "cooked" : "raw",
          edible_pct: d.edible_pct != null ? String(d.edible_pct).replace(".", ",") : "100",
          kcal: d.kcal != null ? String(d.kcal).replace(".", ",") : "",
          carb_g: d.carb_g != null ? String(d.carb_g).replace(".", ",") : "",
          protein_g: d.protein_g != null ? String(d.protein_g).replace(".", ",") : "",
          fat_g: d.fat_g != null ? String(d.fat_g).replace(".", ",") : "",
          sugar_g: d.sugar_g != null ? String(d.sugar_g).replace(".", ",") : "",
          fiber_g: d.fiber_g != null ? String(d.fiber_g).replace(".", ",") : "",
          salt_g: d.salt_g != null ? String(d.salt_g).replace(".", ",") : "",
        });
      } catch (err) {
        setError(err instanceof Error ? err.message : "Errore caricamento");
      } finally {
        if (alive) setLoading(false);
      }
    })();
    return () => {
      alive = false;
    };
  }, [isEdit, editId]);

  const [confirmDelete, setConfirmDelete] = useState(false);

  async function doDelete() {
    if (!editId) return;
    setConfirmDelete(false);
    setDeleting(true);
    try {
      const err = await removeCustomFood(Number(editId));
      if (err) {
        setError(err);
        return;
      }
      navigate("/miei-alimenti", { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Errore imprevisto");
    } finally {
      setDeleting(false);
    }
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      const body = {
        name_it: form.name_it.trim(),
        category: form.category.trim() || undefined,
        state: form.state,
        edible_pct: parseNum(form.edible_pct),
        kcal: parseNum(form.kcal),
        carb_g: parseNum(form.carb_g),
        protein_g: parseNum(form.protein_g),
        fat_g: parseNum(form.fat_g),
        sugar_g: parseNum(form.sugar_g),
        fiber_g: parseNum(form.fiber_g),
        salt_g: parseNum(form.salt_g),
      };
      if (!body.name_it) {
        setError("Il nome è obbligatorio.");
        return;
      }
      if (body.carb_g === undefined) {
        setError("I carboidrati (g/100g) sono obbligatori.");
        return;
      }
      const url = isEdit ? `/foods/custom/${editId}` : "/foods/custom";
      const method = isEdit ? "PATCH" : "POST";
      const res = await apiFetch(url, {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      if (!res.ok && !(isEdit && res.status === 204)) {
        const txt = await res.text().catch(() => "");
        setError(txt || `Errore ${res.status}`);
        return;
      }
      if (isEdit) {
        navigate(`/alimento/${editId}`);
      } else {
        const data = (await res.json()) as { id: number };
        navigate(`/alimento/${data.id}`);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Errore imprevisto");
    } finally {
      setSubmitting(false);
    }
  }

  const input = (
    label: string,
    key: keyof Form,
    placeholder = "",
    required = false,
    type: "text" | "number" = "text",
  ) => (
    <label className="block">
      <span className="block text-sm font-medium mb-1">
        {label}
        {required ? " *" : ""}
      </span>
      <input
        type="text"
        inputMode={type === "number" ? "decimal" : undefined}
        pattern={type === "number" ? "[0-9]*[.,]?[0-9]*" : undefined}
        placeholder={placeholder}
        value={form[key]}
        onChange={(e) => update(key, e.target.value as Form[typeof key])}
        required={required}
        className="w-full border border-[color:var(--color-border)] bg-white px-3.5 py-2.5 text-sm"
      />
    </label>
  );

  if (loading) {
    return <p className="py-8 text-center text-sm opacity-70">Caricamento…</p>;
  }

  return (
    <section className="max-w-xl mx-auto space-y-6">
      <header>
        <h1 className="text-2xl md:text-3xl font-bold">
          {isEdit ? "Modifica alimento" : "Aggiungi un tuo alimento"}
        </h1>
        <p className="text-sm text-[color:var(--color-ink-muted)] mt-1">
          Valori per 100 g. {isEdit ? "Visibile solo a te." : "Rimarrà visibile solo a te."}
        </p>
      </header>

      <form onSubmit={onSubmit} className="space-y-4">
        {input("Nome", "name_it", "es. Panino della nonna", true)}
        {input("Categoria", "category", "es. Dolci")}

        <fieldset className="space-y-1">
          <legend className="text-sm font-medium">Stato</legend>
          <label className="inline-flex items-center gap-2 mr-4">
            <input
              type="radio"
              name="state"
              value="raw"
              checked={form.state === "raw"}
              onChange={() => update("state", "raw")}
            />
            Crudo
          </label>
          <label className="inline-flex items-center gap-2">
            <input
              type="radio"
              name="state"
              value="cooked"
              checked={form.state === "cooked"}
              onChange={() => update("state", "cooked")}
            />
            Cotto
          </label>
        </fieldset>

        <div className="grid grid-cols-2 gap-3">
          {input("Parte edibile (%)", "edible_pct", "100", false, "number")}
          {input("Energia (kcal/100g)", "kcal", "es. 250", false, "number")}
          {input("Carboidrati (g/100g)", "carb_g", "es. 45", true, "number")}
          {input("Zuccheri (g/100g)", "sugar_g", "", false, "number")}
          {input("Proteine (g/100g)", "protein_g", "", false, "number")}
          {input("Grassi (g/100g)", "fat_g", "", false, "number")}
          {input("Fibre (g/100g)", "fiber_g", "", false, "number")}
          {input("Sale (g/100g)", "salt_g", "", false, "number")}
        </div>

        {error ? (
          <p role="alert" className="text-sm text-red-700 bg-red-50 border border-red-200 rounded-xl px-3 py-2">
            {error}
          </p>
        ) : null}

        <div className="flex items-center gap-3 flex-wrap">
          <button
            type="submit"
            disabled={submitting}
            className="px-4 py-2 bg-[color:var(--color-brand-600)] text-white rounded-xl font-medium disabled:opacity-50"
          >
            {submitting ? "Salvataggio…" : "Salva"}
          </button>
          <button
            type="button"
            onClick={() => navigate(-1)}
            className="px-4 py-2 border border-[color:var(--color-border)] rounded-xl"
          >
            Annulla
          </button>
          {isEdit ? (
            <button
              type="button"
              onClick={() => setConfirmDelete(true)}
              disabled={deleting}
              className="ml-auto px-4 py-2 border border-red-300 text-red-700 rounded-xl disabled:opacity-50"
            >
              {deleting ? "Eliminazione…" : "Elimina"}
            </button>
          ) : null}
        </div>
      </form>
      <ConfirmDialog
        open={confirmDelete}
        title="Eliminare l'alimento?"
        message="Questa azione è irreversibile."
        confirmLabel="Elimina"
        variant="danger"
        onConfirm={doDelete}
        onCancel={() => setConfirmDelete(false)}
      />
    </section>
  );
}
