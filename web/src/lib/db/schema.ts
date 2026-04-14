import Dexie, { type EntityTable } from "dexie";
import type {
  LocalFavorite,
  LocalMeal,
  LocalMealItem,
  LocalMealTemplate,
  LocalRecipe,
  LocalRecipeItem,
  OutboxRow,
} from "../sync/types";

/**
 * Local cache of the food catalog for offline search + detail.
 * Keeps only the fields needed for UI; the canonical source stays server-side.
 */
export interface CachedFood {
  id: number;
  name_it: string;
  name_en?: string;
  category?: string;
  state: string;
  kcal?: number;
  carb_g?: number;
  protein_g?: number;
  fat_g?: number;
  sugar_g?: number;
  fiber_g?: number;
  source_code: string;
  source_license: string;
  source_license_url?: string;
  /** lowercased, accent-folded tokens for local FTS */
  search_tokens: string;
  updated_at: string;
}

export interface SyncMeta {
  key: string;
  value: string;
}

export type PesaDB = Dexie & {
  foods: EntityTable<CachedFood, "id">;
  meta: EntityTable<SyncMeta, "key">;
  meals: EntityTable<LocalMeal, "client_id">;
  meal_items: EntityTable<LocalMealItem, "client_id">;
  recipes: EntityTable<LocalRecipe, "client_id">;
  recipe_items: EntityTable<LocalRecipeItem, "client_id">;
  meal_templates: EntityTable<LocalMealTemplate, "client_id">;
  favorites: EntityTable<LocalFavorite, "client_id">;
  outbox: EntityTable<OutboxRow, "id">;
};

const db = new Dexie("pesa") as PesaDB;

// v1: foods + meta only.
db.version(1).stores({
  foods: "id, name_it, state, category, source_code, *search_tokens_arr",
  meta: "key",
});

// v2: adds diary / recipes / templates / favorites / outbox.
// Existing foods+meta rows survive the upgrade (Dexie preserves v1 stores that
// are redeclared verbatim). Compound indexes marked `[...]`. `++id` = auto-incr.
db.version(2).stores({
  foods: "id, name_it, state, category, source_code, *search_tokens_arr",
  meta: "key",
  meals:
    "client_id, user_id, occurred_at, [user_id+occurred_at], updated_at_ms, deleted_at_ms",
  meal_items:
    "client_id, meal_client_id, food_id, recipe_id, updated_at_ms, deleted_at_ms",
  recipes: "client_id, user_id, name, updated_at_ms, deleted_at_ms",
  recipe_items:
    "client_id, recipe_client_id, food_id, updated_at_ms, deleted_at_ms",
  meal_templates:
    "client_id, user_id, updated_at_ms, deleted_at_ms",
  favorites:
    "client_id, user_id, food_id, [user_id+food_id], updated_at_ms, deleted_at_ms",
  outbox:
    "++id, kind, client_id, next_attempt_at, attempts, blocked, created_at",
});

export { db };
