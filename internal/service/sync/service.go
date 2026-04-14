package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

// Service owns the sync application logic. Thin wrapper over *sql.DB — query
// text is inline here (not sqlc generated in this plan to keep the change
// surface focused; plan-02 summary documents the contract for a later sqlc
// extraction).
type Service struct {
	db *sql.DB
}

// New constructs a Service.
func New(db *sql.DB) *Service { return &Service{db: db} }

// UpsertBatch applies a batch atomically-per-entity (one tx across all kinds).
// Ordering: recipes → recipe_items → meals → meal_items → meal_templates → favorites.
// Every row is LWW-reconciled vs current storage via updated_at_ms.
func (s *Service) UpsertBatch(ctx context.Context, userID int64, p BatchPayload) (BatchResponse, error) {
	if userID <= 0 {
		return BatchResponse{}, errors.New("sync: userID required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return BatchResponse{}, fmt.Errorf("sync: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	out := BatchResponse{Results: make([]RowStatus, 0, batchSize(p))}
	var maxCursor int64

	for _, r := range p.Recipes {
		rs, cur := s.upsertRecipe(ctx, tx, userID, r)
		out.Results = append(out.Results, rs)
		if cur > maxCursor {
			maxCursor = cur
		}
	}
	for _, ri := range p.RecipeItems {
		rs, cur := s.upsertRecipeItem(ctx, tx, userID, ri)
		out.Results = append(out.Results, rs)
		if cur > maxCursor {
			maxCursor = cur
		}
	}
	for _, m := range p.Meals {
		rs, cur := s.upsertMeal(ctx, tx, userID, m)
		out.Results = append(out.Results, rs)
		if cur > maxCursor {
			maxCursor = cur
		}
	}
	for _, mi := range p.MealItems {
		rs, cur := s.upsertMealItem(ctx, tx, userID, mi)
		out.Results = append(out.Results, rs)
		if cur > maxCursor {
			maxCursor = cur
		}
	}
	for _, mt := range p.MealTemplates {
		rs, cur := s.upsertMealTemplate(ctx, tx, userID, mt)
		out.Results = append(out.Results, rs)
		if cur > maxCursor {
			maxCursor = cur
		}
	}
	for _, fv := range p.Favorites {
		rs, cur := s.upsertFavorite(ctx, tx, userID, fv)
		out.Results = append(out.Results, rs)
		if cur > maxCursor {
			maxCursor = cur
		}
	}
	if err := tx.Commit(); err != nil {
		return BatchResponse{}, fmt.Errorf("sync: commit: %w", err)
	}
	out.ServerCursor = maxCursor
	return out, nil
}

// Pull returns delta of all entity types with server_updated_at_ms > since.
// LIMIT 500 per entity (T-04.02-07 DoS mitigation); client loops with advancing cursor.
func (s *Service) Pull(ctx context.Context, userID int64, since int64) (PullResponse, error) {
	if userID <= 0 {
		return PullResponse{}, errors.New("sync: userID required")
	}
	var out PullResponse
	var maxCursor = since

	if meals, cur, err := s.pullMeals(ctx, userID, since); err != nil {
		return PullResponse{}, err
	} else {
		out.Meals = meals
		if cur > maxCursor {
			maxCursor = cur
		}
	}
	if items, cur, err := s.pullMealItems(ctx, userID, since); err != nil {
		return PullResponse{}, err
	} else {
		out.MealItems = items
		if cur > maxCursor {
			maxCursor = cur
		}
	}
	if recipes, cur, err := s.pullRecipes(ctx, userID, since); err != nil {
		return PullResponse{}, err
	} else {
		out.Recipes = recipes
		if cur > maxCursor {
			maxCursor = cur
		}
	}
	if rItems, cur, err := s.pullRecipeItems(ctx, userID, since); err != nil {
		return PullResponse{}, err
	} else {
		out.RecipeItems = rItems
		if cur > maxCursor {
			maxCursor = cur
		}
	}
	if tmpls, cur, err := s.pullMealTemplates(ctx, userID, since); err != nil {
		return PullResponse{}, err
	} else {
		out.MealTemplates = tmpls
		if cur > maxCursor {
			maxCursor = cur
		}
	}
	if favs, cur, err := s.pullFavorites(ctx, userID, since); err != nil {
		return PullResponse{}, err
	} else {
		out.Favorites = favs
		if cur > maxCursor {
			maxCursor = cur
		}
	}
	out.ServerCursor = maxCursor
	return out, nil
}

// ---------- meal ----------

func (s *Service) upsertMeal(ctx context.Context, tx *sql.Tx, userID int64, d MealDTO) (RowStatus, int64) {
	var existingUpdated sql.NullInt64
	var existingID sql.NullInt64
	err := tx.QueryRowContext(ctx,
		`SELECT id, updated_at_ms FROM meals WHERE user_id=$1 AND client_id=$2 FOR UPDATE`,
		userID, d.ClientID).Scan(&existingID, &existingUpdated)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return RowStatus{Kind: KindMeal, ClientID: d.ClientID, Status: StatusError, Reason: "select_failed"}, 0
	}
	if existingUpdated.Valid && d.UpdatedAtMs < existingUpdated.Int64 {
		return RowStatus{Kind: KindMeal, ClientID: d.ClientID, Status: StatusConflictSkipped, Reason: "older_timestamp", ServerID: existingID.Int64}, 0
	}
	// LWW-newer OR equal (idempotent) OR new row → upsert.
	var id int64
	err = tx.QueryRowContext(ctx, `
INSERT INTO meals (user_id, client_id, occurred_at, slot, notes, updated_at_ms, deleted_at_ms)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (user_id, client_id) DO UPDATE SET
  occurred_at   = CASE WHEN EXCLUDED.updated_at_ms >= meals.updated_at_ms THEN EXCLUDED.occurred_at   ELSE meals.occurred_at   END,
  slot          = CASE WHEN EXCLUDED.updated_at_ms >= meals.updated_at_ms THEN EXCLUDED.slot          ELSE meals.slot          END,
  notes         = CASE WHEN EXCLUDED.updated_at_ms >= meals.updated_at_ms THEN EXCLUDED.notes         ELSE meals.notes         END,
  deleted_at_ms = CASE WHEN EXCLUDED.updated_at_ms >= meals.updated_at_ms THEN EXCLUDED.deleted_at_ms ELSE meals.deleted_at_ms END,
  updated_at_ms = CASE WHEN EXCLUDED.updated_at_ms >= meals.updated_at_ms THEN EXCLUDED.updated_at_ms ELSE meals.updated_at_ms END
RETURNING id`,
		userID, d.ClientID, d.OccurredAt, d.Slot, d.Notes, d.UpdatedAtMs, d.DeletedAtMs).Scan(&id)
	if err != nil {
		return RowStatus{Kind: KindMeal, ClientID: d.ClientID, Status: StatusError, Reason: "upsert_failed"}, 0
	}
	if id == 0 && existingID.Valid {
		id = existingID.Int64
	}
	cur := currentCursorMs(ctx, tx, "meals", userID, d.ClientID)
	return RowStatus{Kind: KindMeal, ClientID: d.ClientID, Status: StatusApplied, ServerID: id}, cur
}

// ---------- meal_item ----------

func (s *Service) upsertMealItem(ctx context.Context, tx *sql.Tx, userID int64, d MealItemDTO) (RowStatus, int64) {
	// Resolve parent meal_id from (user_id, meal_client_id).
	var mealID int64
	err := tx.QueryRowContext(ctx,
		`SELECT id FROM meals WHERE user_id=$1 AND client_id=$2`, userID, d.MealClientID).Scan(&mealID)
	if err != nil {
		return RowStatus{Kind: KindMealItem, ClientID: d.ClientID, Status: StatusError, Reason: "parent_not_found"}, 0
	}
	// Resolve recipe_id (optional).
	var recipeID sql.NullInt64
	if d.RecipeClientID != nil && *d.RecipeClientID != "" {
		if err := tx.QueryRowContext(ctx,
			`SELECT id FROM recipes WHERE user_id=$1 AND client_id=$2`, userID, *d.RecipeClientID).Scan(&recipeID); err != nil {
			// recipe not found → keep NULL (historical scenario); item.food_id may still satisfy CHECK.
			recipeID = sql.NullInt64{}
		}
	}
	// LWW check.
	var existingUpdated sql.NullInt64
	var existingID sql.NullInt64
	_ = tx.QueryRowContext(ctx,
		`SELECT id, updated_at_ms FROM meal_items WHERE meal_id=$1 AND client_id=$2 FOR UPDATE`,
		mealID, d.ClientID).Scan(&existingID, &existingUpdated)
	if existingUpdated.Valid && d.UpdatedAtMs < existingUpdated.Int64 {
		return RowStatus{Kind: KindMealItem, ClientID: d.ClientID, Status: StatusConflictSkipped, Reason: "older_timestamp", ServerID: existingID.Int64}, 0
	}
	var id int64
	err = tx.QueryRowContext(ctx, `
INSERT INTO meal_items (meal_id, client_id, food_id, recipe_id, name_snapshot, grams,
  kcal, protein_g, fat_g, carb_g, sugar_g, fiber_g, updated_at_ms, deleted_at_ms)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
ON CONFLICT (meal_id, client_id) DO UPDATE SET
  food_id       = CASE WHEN EXCLUDED.updated_at_ms >= meal_items.updated_at_ms THEN EXCLUDED.food_id       ELSE meal_items.food_id       END,
  recipe_id     = CASE WHEN EXCLUDED.updated_at_ms >= meal_items.updated_at_ms THEN EXCLUDED.recipe_id     ELSE meal_items.recipe_id     END,
  name_snapshot = CASE WHEN EXCLUDED.updated_at_ms >= meal_items.updated_at_ms THEN EXCLUDED.name_snapshot ELSE meal_items.name_snapshot END,
  grams         = CASE WHEN EXCLUDED.updated_at_ms >= meal_items.updated_at_ms THEN EXCLUDED.grams         ELSE meal_items.grams         END,
  kcal          = CASE WHEN EXCLUDED.updated_at_ms >= meal_items.updated_at_ms THEN EXCLUDED.kcal          ELSE meal_items.kcal          END,
  protein_g     = CASE WHEN EXCLUDED.updated_at_ms >= meal_items.updated_at_ms THEN EXCLUDED.protein_g     ELSE meal_items.protein_g     END,
  fat_g         = CASE WHEN EXCLUDED.updated_at_ms >= meal_items.updated_at_ms THEN EXCLUDED.fat_g         ELSE meal_items.fat_g         END,
  carb_g        = CASE WHEN EXCLUDED.updated_at_ms >= meal_items.updated_at_ms THEN EXCLUDED.carb_g        ELSE meal_items.carb_g        END,
  sugar_g       = CASE WHEN EXCLUDED.updated_at_ms >= meal_items.updated_at_ms THEN EXCLUDED.sugar_g       ELSE meal_items.sugar_g       END,
  fiber_g       = CASE WHEN EXCLUDED.updated_at_ms >= meal_items.updated_at_ms THEN EXCLUDED.fiber_g       ELSE meal_items.fiber_g       END,
  deleted_at_ms = CASE WHEN EXCLUDED.updated_at_ms >= meal_items.updated_at_ms THEN EXCLUDED.deleted_at_ms ELSE meal_items.deleted_at_ms END,
  updated_at_ms = CASE WHEN EXCLUDED.updated_at_ms >= meal_items.updated_at_ms THEN EXCLUDED.updated_at_ms ELSE meal_items.updated_at_ms END
RETURNING id`,
		mealID, d.ClientID, d.FoodID, recipeID, d.NameSnapshot, d.Grams,
		d.Kcal, d.ProteinG, d.FatG, d.CarbG, d.SugarG, d.FiberG, d.UpdatedAtMs, d.DeletedAtMs).Scan(&id)
	if err != nil {
		return RowStatus{Kind: KindMealItem, ClientID: d.ClientID, Status: StatusError, Reason: "upsert_failed"}, 0
	}
	if id == 0 && existingID.Valid {
		id = existingID.Int64
	}
	var cur int64
	_ = tx.QueryRowContext(ctx,
		`SELECT (EXTRACT(EPOCH FROM server_updated_at)*1000)::BIGINT FROM meal_items WHERE meal_id=$1 AND client_id=$2`,
		mealID, d.ClientID).Scan(&cur)
	return RowStatus{Kind: KindMealItem, ClientID: d.ClientID, Status: StatusApplied, ServerID: id}, cur
}

// ---------- recipe ----------

func (s *Service) upsertRecipe(ctx context.Context, tx *sql.Tx, userID int64, d RecipeDTO) (RowStatus, int64) {
	var existingUpdated sql.NullInt64
	var existingID sql.NullInt64
	err := tx.QueryRowContext(ctx,
		`SELECT id, updated_at_ms FROM recipes WHERE user_id=$1 AND client_id=$2 FOR UPDATE`,
		userID, d.ClientID).Scan(&existingID, &existingUpdated)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return RowStatus{Kind: KindRecipe, ClientID: d.ClientID, Status: StatusError, Reason: "select_failed"}, 0
	}
	if existingUpdated.Valid && d.UpdatedAtMs < existingUpdated.Int64 {
		return RowStatus{Kind: KindRecipe, ClientID: d.ClientID, Status: StatusConflictSkipped, Reason: "older_timestamp", ServerID: existingID.Int64}, 0
	}
	var id int64
	err = tx.QueryRowContext(ctx, `
INSERT INTO recipes (user_id, client_id, name, servings, total_grams, notes, updated_at_ms, deleted_at_ms)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (user_id, client_id) DO UPDATE SET
  name          = CASE WHEN EXCLUDED.updated_at_ms >= recipes.updated_at_ms THEN EXCLUDED.name          ELSE recipes.name          END,
  servings      = CASE WHEN EXCLUDED.updated_at_ms >= recipes.updated_at_ms THEN EXCLUDED.servings      ELSE recipes.servings      END,
  total_grams   = CASE WHEN EXCLUDED.updated_at_ms >= recipes.updated_at_ms THEN EXCLUDED.total_grams   ELSE recipes.total_grams   END,
  notes         = CASE WHEN EXCLUDED.updated_at_ms >= recipes.updated_at_ms THEN EXCLUDED.notes         ELSE recipes.notes         END,
  deleted_at_ms = CASE WHEN EXCLUDED.updated_at_ms >= recipes.updated_at_ms THEN EXCLUDED.deleted_at_ms ELSE recipes.deleted_at_ms END,
  updated_at_ms = CASE WHEN EXCLUDED.updated_at_ms >= recipes.updated_at_ms THEN EXCLUDED.updated_at_ms ELSE recipes.updated_at_ms END
RETURNING id`,
		userID, d.ClientID, d.Name, d.Servings, d.TotalGrams, d.Notes, d.UpdatedAtMs, d.DeletedAtMs).Scan(&id)
	if err != nil {
		return RowStatus{Kind: KindRecipe, ClientID: d.ClientID, Status: StatusError, Reason: "upsert_failed"}, 0
	}
	if id == 0 && existingID.Valid {
		id = existingID.Int64
	}
	cur := currentCursorMs(ctx, tx, "recipes", userID, d.ClientID)
	return RowStatus{Kind: KindRecipe, ClientID: d.ClientID, Status: StatusApplied, ServerID: id}, cur
}

// ---------- recipe_item ----------

func (s *Service) upsertRecipeItem(ctx context.Context, tx *sql.Tx, userID int64, d RecipeItemDTO) (RowStatus, int64) {
	var recipeID int64
	err := tx.QueryRowContext(ctx,
		`SELECT id FROM recipes WHERE user_id=$1 AND client_id=$2`, userID, d.RecipeClientID).Scan(&recipeID)
	if err != nil {
		return RowStatus{Kind: KindRecipeItem, ClientID: d.ClientID, Status: StatusError, Reason: "parent_not_found"}, 0
	}
	var existingUpdated sql.NullInt64
	var existingID sql.NullInt64
	_ = tx.QueryRowContext(ctx,
		`SELECT id, updated_at_ms FROM recipe_items WHERE recipe_id=$1 AND client_id=$2 FOR UPDATE`,
		recipeID, d.ClientID).Scan(&existingID, &existingUpdated)
	if existingUpdated.Valid && d.UpdatedAtMs < existingUpdated.Int64 {
		return RowStatus{Kind: KindRecipeItem, ClientID: d.ClientID, Status: StatusConflictSkipped, Reason: "older_timestamp", ServerID: existingID.Int64}, 0
	}
	snapBytes, err := json.Marshal(d.SnapshotJSON)
	if err != nil {
		return RowStatus{Kind: KindRecipeItem, ClientID: d.ClientID, Status: StatusError, Reason: "invalid_snapshot"}, 0
	}
	var id int64
	err = tx.QueryRowContext(ctx, `
INSERT INTO recipe_items (recipe_id, client_id, food_id, name_snapshot, grams, snapshot_json, updated_at_ms, deleted_at_ms)
VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8)
ON CONFLICT (recipe_id, client_id) DO UPDATE SET
  food_id       = CASE WHEN EXCLUDED.updated_at_ms >= recipe_items.updated_at_ms THEN EXCLUDED.food_id       ELSE recipe_items.food_id       END,
  name_snapshot = CASE WHEN EXCLUDED.updated_at_ms >= recipe_items.updated_at_ms THEN EXCLUDED.name_snapshot ELSE recipe_items.name_snapshot END,
  grams         = CASE WHEN EXCLUDED.updated_at_ms >= recipe_items.updated_at_ms THEN EXCLUDED.grams         ELSE recipe_items.grams         END,
  snapshot_json = CASE WHEN EXCLUDED.updated_at_ms >= recipe_items.updated_at_ms THEN EXCLUDED.snapshot_json ELSE recipe_items.snapshot_json END,
  deleted_at_ms = CASE WHEN EXCLUDED.updated_at_ms >= recipe_items.updated_at_ms THEN EXCLUDED.deleted_at_ms ELSE recipe_items.deleted_at_ms END,
  updated_at_ms = CASE WHEN EXCLUDED.updated_at_ms >= recipe_items.updated_at_ms THEN EXCLUDED.updated_at_ms ELSE recipe_items.updated_at_ms END
RETURNING id`,
		recipeID, d.ClientID, d.FoodID, d.NameSnapshot, d.Grams, string(snapBytes), d.UpdatedAtMs, d.DeletedAtMs).Scan(&id)
	if err != nil {
		return RowStatus{Kind: KindRecipeItem, ClientID: d.ClientID, Status: StatusError, Reason: "upsert_failed"}, 0
	}
	if id == 0 && existingID.Valid {
		id = existingID.Int64
	}
	var cur int64
	_ = tx.QueryRowContext(ctx,
		`SELECT (EXTRACT(EPOCH FROM server_updated_at)*1000)::BIGINT FROM recipe_items WHERE recipe_id=$1 AND client_id=$2`,
		recipeID, d.ClientID).Scan(&cur)
	return RowStatus{Kind: KindRecipeItem, ClientID: d.ClientID, Status: StatusApplied, ServerID: id}, cur
}

// ---------- meal_template ----------

func (s *Service) upsertMealTemplate(ctx context.Context, tx *sql.Tx, userID int64, d MealTemplateDTO) (RowStatus, int64) {
	var existingUpdated sql.NullInt64
	var existingID sql.NullInt64
	err := tx.QueryRowContext(ctx,
		`SELECT id, updated_at_ms FROM meal_templates WHERE user_id=$1 AND client_id=$2 FOR UPDATE`,
		userID, d.ClientID).Scan(&existingID, &existingUpdated)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return RowStatus{Kind: KindMealTemplate, ClientID: d.ClientID, Status: StatusError, Reason: "select_failed"}, 0
	}
	if existingUpdated.Valid && d.UpdatedAtMs < existingUpdated.Int64 {
		return RowStatus{Kind: KindMealTemplate, ClientID: d.ClientID, Status: StatusConflictSkipped, Reason: "older_timestamp", ServerID: existingID.Int64}, 0
	}
	payload, err := json.Marshal(d.PayloadJSON)
	if err != nil {
		return RowStatus{Kind: KindMealTemplate, ClientID: d.ClientID, Status: StatusError, Reason: "invalid_payload"}, 0
	}
	var id int64
	err = tx.QueryRowContext(ctx, `
INSERT INTO meal_templates (user_id, client_id, name, payload_json, updated_at_ms, deleted_at_ms)
VALUES ($1, $2, $3, $4::jsonb, $5, $6)
ON CONFLICT (user_id, client_id) DO UPDATE SET
  name          = CASE WHEN EXCLUDED.updated_at_ms >= meal_templates.updated_at_ms THEN EXCLUDED.name          ELSE meal_templates.name          END,
  payload_json  = CASE WHEN EXCLUDED.updated_at_ms >= meal_templates.updated_at_ms THEN EXCLUDED.payload_json  ELSE meal_templates.payload_json  END,
  deleted_at_ms = CASE WHEN EXCLUDED.updated_at_ms >= meal_templates.updated_at_ms THEN EXCLUDED.deleted_at_ms ELSE meal_templates.deleted_at_ms END,
  updated_at_ms = CASE WHEN EXCLUDED.updated_at_ms >= meal_templates.updated_at_ms THEN EXCLUDED.updated_at_ms ELSE meal_templates.updated_at_ms END
RETURNING id`,
		userID, d.ClientID, d.Name, string(payload), d.UpdatedAtMs, d.DeletedAtMs).Scan(&id)
	if err != nil {
		return RowStatus{Kind: KindMealTemplate, ClientID: d.ClientID, Status: StatusError, Reason: "upsert_failed"}, 0
	}
	if id == 0 && existingID.Valid {
		id = existingID.Int64
	}
	cur := currentCursorMs(ctx, tx, "meal_templates", userID, d.ClientID)
	return RowStatus{Kind: KindMealTemplate, ClientID: d.ClientID, Status: StatusApplied, ServerID: id}, cur
}

// ---------- favorite ----------

func (s *Service) upsertFavorite(ctx context.Context, tx *sql.Tx, userID int64, d FavoriteDTO) (RowStatus, int64) {
	// Detect (user_id, food_id) collision with different client_id BEFORE insert.
	var existingClientID string
	var existingDeleted sql.NullInt64
	var existingID int64
	err := tx.QueryRowContext(ctx,
		`SELECT id, client_id, deleted_at_ms FROM favorites WHERE user_id=$1 AND food_id=$2 FOR UPDATE`,
		userID, d.FoodID).Scan(&existingID, &existingClientID, &existingDeleted)
	if err == nil && existingClientID != d.ClientID && !existingDeleted.Valid {
		return RowStatus{Kind: KindFavorite, ClientID: d.ClientID, Status: StatusConflictSkipped, Reason: "duplicate_food", ServerID: existingID}, 0
	}
	// LWW on (user_id, client_id).
	var existingUpdated sql.NullInt64
	var rowID sql.NullInt64
	_ = tx.QueryRowContext(ctx,
		`SELECT id, updated_at_ms FROM favorites WHERE user_id=$1 AND client_id=$2 FOR UPDATE`,
		userID, d.ClientID).Scan(&rowID, &existingUpdated)
	if existingUpdated.Valid && d.UpdatedAtMs < existingUpdated.Int64 {
		return RowStatus{Kind: KindFavorite, ClientID: d.ClientID, Status: StatusConflictSkipped, Reason: "older_timestamp", ServerID: rowID.Int64}, 0
	}
	var id int64
	err = tx.QueryRowContext(ctx, `
INSERT INTO favorites (user_id, client_id, food_id, updated_at_ms, deleted_at_ms)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, client_id) DO UPDATE SET
  food_id       = CASE WHEN EXCLUDED.updated_at_ms >= favorites.updated_at_ms THEN EXCLUDED.food_id       ELSE favorites.food_id       END,
  deleted_at_ms = CASE WHEN EXCLUDED.updated_at_ms >= favorites.updated_at_ms THEN EXCLUDED.deleted_at_ms ELSE favorites.deleted_at_ms END,
  updated_at_ms = CASE WHEN EXCLUDED.updated_at_ms >= favorites.updated_at_ms THEN EXCLUDED.updated_at_ms ELSE favorites.updated_at_ms END
RETURNING id`,
		userID, d.ClientID, d.FoodID, d.UpdatedAtMs, d.DeletedAtMs).Scan(&id)
	if err != nil {
		return RowStatus{Kind: KindFavorite, ClientID: d.ClientID, Status: StatusError, Reason: "upsert_failed"}, 0
	}
	if id == 0 && rowID.Valid {
		id = rowID.Int64
	}
	cur := currentCursorMs(ctx, tx, "favorites", userID, d.ClientID)
	return RowStatus{Kind: KindFavorite, ClientID: d.ClientID, Status: StatusApplied, ServerID: id}, cur
}

// ---------- pull helpers ----------

func (s *Service) pullMeals(ctx context.Context, userID, since int64) ([]MealDTO, int64, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT client_id, occurred_at, slot, notes, updated_at_ms, deleted_at_ms,
       (EXTRACT(EPOCH FROM server_updated_at)*1000)::BIGINT AS cur
FROM meals
WHERE user_id=$1 AND (EXTRACT(EPOCH FROM server_updated_at)*1000)::BIGINT > $2
ORDER BY server_updated_at ASC
LIMIT 500`, userID, since)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	var out []MealDTO
	var maxCur int64
	for rows.Next() {
		var d MealDTO
		var notes sql.NullString
		var deleted sql.NullInt64
		var cur int64
		if err := rows.Scan(&d.ClientID, &d.OccurredAt, &d.Slot, &notes, &d.UpdatedAtMs, &deleted, &cur); err != nil {
			return nil, 0, err
		}
		if notes.Valid {
			v := notes.String
			d.Notes = &v
		}
		if deleted.Valid {
			v := deleted.Int64
			d.DeletedAtMs = &v
		}
		if cur > maxCur {
			maxCur = cur
		}
		out = append(out, d)
	}
	return out, maxCur, rows.Err()
}

func (s *Service) pullMealItems(ctx context.Context, userID, since int64) ([]MealItemDTO, int64, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT mi.client_id, m.client_id AS meal_client_id, mi.food_id,
       (SELECT client_id FROM recipes r WHERE r.id=mi.recipe_id) AS recipe_client_id,
       mi.name_snapshot, mi.grams, mi.kcal, mi.protein_g, mi.fat_g, mi.carb_g, mi.sugar_g, mi.fiber_g,
       mi.updated_at_ms, mi.deleted_at_ms,
       (EXTRACT(EPOCH FROM mi.server_updated_at)*1000)::BIGINT AS cur
FROM meal_items mi
JOIN meals m ON m.id = mi.meal_id
WHERE m.user_id=$1 AND (EXTRACT(EPOCH FROM mi.server_updated_at)*1000)::BIGINT > $2
ORDER BY mi.server_updated_at ASC
LIMIT 500`, userID, since)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	var out []MealItemDTO
	var maxCur int64
	for rows.Next() {
		var d MealItemDTO
		var deleted sql.NullInt64
		var foodID sql.NullInt64
		var recipeCID sql.NullString
		var kcal, p, f, c, sg, fb sql.NullFloat64
		var cur int64
		if err := rows.Scan(&d.ClientID, &d.MealClientID, &foodID, &recipeCID, &d.NameSnapshot, &d.Grams,
			&kcal, &p, &f, &c, &sg, &fb, &d.UpdatedAtMs, &deleted, &cur); err != nil {
			return nil, 0, err
		}
		if foodID.Valid {
			v := foodID.Int64
			d.FoodID = &v
		}
		if recipeCID.Valid {
			v := recipeCID.String
			d.RecipeClientID = &v
		}
		if kcal.Valid {
			v := kcal.Float64
			d.Kcal = &v
		}
		if p.Valid {
			v := p.Float64
			d.ProteinG = &v
		}
		if f.Valid {
			v := f.Float64
			d.FatG = &v
		}
		if c.Valid {
			v := c.Float64
			d.CarbG = &v
		}
		if sg.Valid {
			v := sg.Float64
			d.SugarG = &v
		}
		if fb.Valid {
			v := fb.Float64
			d.FiberG = &v
		}
		if deleted.Valid {
			v := deleted.Int64
			d.DeletedAtMs = &v
		}
		if cur > maxCur {
			maxCur = cur
		}
		out = append(out, d)
	}
	return out, maxCur, rows.Err()
}

func (s *Service) pullRecipes(ctx context.Context, userID, since int64) ([]RecipeDTO, int64, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT client_id, name, servings, total_grams, notes, updated_at_ms, deleted_at_ms,
       (EXTRACT(EPOCH FROM server_updated_at)*1000)::BIGINT AS cur
FROM recipes
WHERE user_id=$1 AND (EXTRACT(EPOCH FROM server_updated_at)*1000)::BIGINT > $2
ORDER BY server_updated_at ASC
LIMIT 500`, userID, since)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	var out []RecipeDTO
	var maxCur int64
	for rows.Next() {
		var d RecipeDTO
		var total sql.NullFloat64
		var notes sql.NullString
		var deleted sql.NullInt64
		var cur int64
		if err := rows.Scan(&d.ClientID, &d.Name, &d.Servings, &total, &notes, &d.UpdatedAtMs, &deleted, &cur); err != nil {
			return nil, 0, err
		}
		if total.Valid {
			v := total.Float64
			d.TotalGrams = &v
		}
		if notes.Valid {
			v := notes.String
			d.Notes = &v
		}
		if deleted.Valid {
			v := deleted.Int64
			d.DeletedAtMs = &v
		}
		if cur > maxCur {
			maxCur = cur
		}
		out = append(out, d)
	}
	return out, maxCur, rows.Err()
}

func (s *Service) pullRecipeItems(ctx context.Context, userID, since int64) ([]RecipeItemDTO, int64, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT ri.client_id, r.client_id AS recipe_client_id, ri.food_id, ri.name_snapshot, ri.grams,
       ri.snapshot_json, ri.updated_at_ms, ri.deleted_at_ms,
       (EXTRACT(EPOCH FROM ri.server_updated_at)*1000)::BIGINT AS cur
FROM recipe_items ri
JOIN recipes r ON r.id = ri.recipe_id
WHERE r.user_id=$1 AND (EXTRACT(EPOCH FROM ri.server_updated_at)*1000)::BIGINT > $2
ORDER BY ri.server_updated_at ASC
LIMIT 500`, userID, since)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	var out []RecipeItemDTO
	var maxCur int64
	for rows.Next() {
		var d RecipeItemDTO
		var foodID sql.NullInt64
		var deleted sql.NullInt64
		var snap []byte
		var cur int64
		if err := rows.Scan(&d.ClientID, &d.RecipeClientID, &foodID, &d.NameSnapshot, &d.Grams, &snap, &d.UpdatedAtMs, &deleted, &cur); err != nil {
			return nil, 0, err
		}
		if foodID.Valid {
			v := foodID.Int64
			d.FoodID = &v
		}
		if deleted.Valid {
			v := deleted.Int64
			d.DeletedAtMs = &v
		}
		var anyVal any
		_ = json.Unmarshal(snap, &anyVal)
		d.SnapshotJSON = anyVal
		if cur > maxCur {
			maxCur = cur
		}
		out = append(out, d)
	}
	return out, maxCur, rows.Err()
}

func (s *Service) pullMealTemplates(ctx context.Context, userID, since int64) ([]MealTemplateDTO, int64, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT client_id, name, payload_json, updated_at_ms, deleted_at_ms,
       (EXTRACT(EPOCH FROM server_updated_at)*1000)::BIGINT AS cur
FROM meal_templates
WHERE user_id=$1 AND (EXTRACT(EPOCH FROM server_updated_at)*1000)::BIGINT > $2
ORDER BY server_updated_at ASC
LIMIT 500`, userID, since)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	var out []MealTemplateDTO
	var maxCur int64
	for rows.Next() {
		var d MealTemplateDTO
		var deleted sql.NullInt64
		var payload []byte
		var cur int64
		if err := rows.Scan(&d.ClientID, &d.Name, &payload, &d.UpdatedAtMs, &deleted, &cur); err != nil {
			return nil, 0, err
		}
		if deleted.Valid {
			v := deleted.Int64
			d.DeletedAtMs = &v
		}
		var anyVal any
		_ = json.Unmarshal(payload, &anyVal)
		d.PayloadJSON = anyVal
		if cur > maxCur {
			maxCur = cur
		}
		out = append(out, d)
	}
	return out, maxCur, rows.Err()
}

func (s *Service) pullFavorites(ctx context.Context, userID, since int64) ([]FavoriteDTO, int64, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT client_id, food_id, updated_at_ms, deleted_at_ms,
       (EXTRACT(EPOCH FROM server_updated_at)*1000)::BIGINT AS cur
FROM favorites
WHERE user_id=$1 AND (EXTRACT(EPOCH FROM server_updated_at)*1000)::BIGINT > $2
ORDER BY server_updated_at ASC
LIMIT 500`, userID, since)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	var out []FavoriteDTO
	var maxCur int64
	for rows.Next() {
		var d FavoriteDTO
		var deleted sql.NullInt64
		var cur int64
		if err := rows.Scan(&d.ClientID, &d.FoodID, &d.UpdatedAtMs, &deleted, &cur); err != nil {
			return nil, 0, err
		}
		if deleted.Valid {
			v := deleted.Int64
			d.DeletedAtMs = &v
		}
		if cur > maxCur {
			maxCur = cur
		}
		out = append(out, d)
	}
	return out, maxCur, rows.Err()
}

// ---------- helpers ----------

func currentCursorMs(ctx context.Context, tx *sql.Tx, table string, userID int64, clientID string) int64 {
	var cur int64
	// #nosec G201 — table is a compile-time constant; no user input.
	_ = tx.QueryRowContext(ctx,
		"SELECT (EXTRACT(EPOCH FROM server_updated_at)*1000)::BIGINT FROM "+table+" WHERE user_id=$1 AND client_id=$2",
		userID, clientID).Scan(&cur)
	return cur
}

func batchSize(p BatchPayload) int {
	return len(p.Meals) + len(p.MealItems) + len(p.Recipes) + len(p.RecipeItems) + len(p.MealTemplates) + len(p.Favorites)
}
