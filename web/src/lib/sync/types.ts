// Wire-level DTOs — MUST mirror internal/service/sync/types.go exactly.
// See .planning/phases/04-diary-sync/04-plan-02-SUMMARY.md for the contract.

export type Slot = "colazione" | "pranzo" | "cena" | "spuntino" | "altro";

export interface MealDTO {
  client_id: string;
  updated_at_ms: number;
  deleted_at_ms?: number | null;
  occurred_at: string; // "YYYY-MM-DDTHH:mm:ssZ" UTC
  slot: Slot;
  notes?: string | null;
}

export interface MealItemDTO {
  client_id: string;
  updated_at_ms: number;
  deleted_at_ms?: number | null;
  meal_client_id: string;
  food_id?: number | null;
  recipe_client_id?: string | null;
  name_snapshot: string;
  grams: number;
  kcal?: number | null;
  protein_g?: number | null;
  fat_g?: number | null;
  carb_g?: number | null;
  sugar_g?: number | null;
  fiber_g?: number | null;
}

export interface RecipeDTO {
  client_id: string;
  updated_at_ms: number;
  deleted_at_ms?: number | null;
  name: string;
  servings: number;
  total_grams?: number | null;
  notes?: string | null;
}

export interface RecipeItemSnapshot {
  kcal?: number;
  carb_g?: number;
  protein_g?: number;
  fat_g?: number;
  sugar_g?: number;
  fiber_g?: number;
  source_food_name?: string;
}

export interface RecipeItemDTO {
  client_id: string;
  updated_at_ms: number;
  deleted_at_ms?: number | null;
  recipe_client_id: string;
  food_id?: number | null;
  name_snapshot: string;
  grams: number;
  snapshot_json: RecipeItemSnapshot;
}

export interface MealTemplateDTO {
  client_id: string;
  updated_at_ms: number;
  deleted_at_ms?: number | null;
  name: string;
  payload_json: unknown;
}

export interface FavoriteDTO {
  client_id: string;
  updated_at_ms: number;
  deleted_at_ms?: number | null;
  food_id: number;
}

export interface BatchPayload {
  meals?: MealDTO[];
  meal_items?: MealItemDTO[];
  recipes?: RecipeDTO[];
  recipe_items?: RecipeItemDTO[];
  meal_templates?: MealTemplateDTO[];
  favorites?: FavoriteDTO[];
}

export type RowStatusKind =
  | "applied"
  | "conflict_skipped"
  | "error";

export interface RowStatus {
  kind: OutboxKind;
  client_id: string;
  status: RowStatusKind;
  reason?: string;
  server_id?: number;
}

export interface BatchResponse {
  results: RowStatus[];
  server_cursor: number;
}

export type PullResponse = BatchPayload & { server_cursor: number };

// ---- Client-only types ----

export type OutboxKind =
  | "meal"
  | "meal_item"
  | "recipe"
  | "recipe_item"
  | "meal_template"
  | "favorite";

export interface OutboxRow {
  id?: number;
  kind: OutboxKind;
  client_id: string;
  payload_json: string;
  created_at: number;
  next_attempt_at: number;
  attempts: number;
  last_error?: string;
  blocked?: boolean;
}

// ---- Local Dexie shapes ----

export interface LocalMeal {
  client_id: string;
  user_id: number;
  occurred_at: number; // ms UTC
  slot: Slot;
  notes?: string;
  updated_at_ms: number;
  deleted_at_ms?: number | null;
  server_id?: number;
}

export interface LocalMealItem {
  client_id: string;
  meal_client_id: string;
  food_id?: number | null;
  recipe_id?: number | null; // local resolved server id if known
  recipe_client_id?: string | null;
  name_snapshot: string;
  grams: number;
  kcal?: number | null;
  protein_g?: number | null;
  fat_g?: number | null;
  carb_g?: number | null;
  sugar_g?: number | null;
  fiber_g?: number | null;
  updated_at_ms: number;
  deleted_at_ms?: number | null;
}

export interface LocalRecipe {
  client_id: string;
  user_id: number;
  name: string;
  servings: number;
  total_grams?: number | null;
  notes?: string | null;
  updated_at_ms: number;
  deleted_at_ms?: number | null;
}

export interface LocalRecipeItem {
  client_id: string;
  recipe_client_id: string;
  food_id?: number | null;
  name_snapshot: string;
  grams: number;
  snapshot_json: RecipeItemSnapshot;
  updated_at_ms: number;
  deleted_at_ms?: number | null;
}

export interface LocalMealTemplate {
  client_id: string;
  user_id: number;
  name: string;
  payload_json: string; // JSON-stringified for indexing simplicity
  updated_at_ms: number;
  deleted_at_ms?: number | null;
}

export interface LocalFavorite {
  client_id: string;
  user_id: number;
  food_id: number;
  updated_at_ms: number;
  deleted_at_ms?: number | null;
}
