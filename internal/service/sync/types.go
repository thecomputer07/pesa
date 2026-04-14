// Package sync implements the /api/v1/sync server contract: batched idempotent
// upsert + cursor-based delta pull for Phase 4 synced entities.
//
// Invariants:
//   - Every DTO carries (ClientID, UpdatedAtMs, DeletedAtMs).
//   - MealItem / RecipeItem reference the PARENT via *_client_id, never server id
//     (parent may not yet exist server-side when child arrives; resolver does the join).
//   - Foods are NEVER accepted from clients (server-wins catalog).
package sync

// Entity kinds as reported in RowStatus.Kind.
const (
	KindMeal         = "meal"
	KindMealItem     = "meal_item"
	KindRecipe       = "recipe"
	KindRecipeItem   = "recipe_item"
	KindMealTemplate = "meal_template"
	KindFavorite     = "favorite"
)

// RowStatus values.
const (
	StatusApplied         = "applied"
	StatusConflictSkipped = "conflict_skipped"
	StatusError           = "error"
)

// Common LWW fields embedded in every DTO.
type base struct {
	ClientID    string `json:"client_id"     validate:"required,uuid"`
	UpdatedAtMs int64  `json:"updated_at_ms" validate:"required"`
	DeletedAtMs *int64 `json:"deleted_at_ms,omitempty"`
}

// MealDTO is the on-the-wire shape for meals.
type MealDTO struct {
	ClientID    string `json:"client_id"     validate:"required,uuid"`
	UpdatedAtMs int64  `json:"updated_at_ms" validate:"required"`
	DeletedAtMs *int64 `json:"deleted_at_ms,omitempty"`

	// "2006-01-02T15:04:05Z" UTC (server converts to TIMESTAMP).
	OccurredAt string  `json:"occurred_at" validate:"required"`
	Slot       string  `json:"slot"        validate:"required,oneof=colazione pranzo cena spuntino altro"`
	Notes      *string `json:"notes,omitempty"`
}

// MealItemDTO uses MealClientID to resolve the parent at apply time.
type MealItemDTO struct {
	ClientID    string `json:"client_id"     validate:"required,uuid"`
	UpdatedAtMs int64  `json:"updated_at_ms" validate:"required"`
	DeletedAtMs *int64 `json:"deleted_at_ms,omitempty"`

	MealClientID  string   `json:"meal_client_id" validate:"required,uuid"`
	FoodID        *int64   `json:"food_id,omitempty"`
	RecipeClientID *string `json:"recipe_client_id,omitempty"`
	NameSnapshot  string   `json:"name_snapshot" validate:"required,max=255"`
	Grams         float64  `json:"grams"         validate:"required,gt=0"`
	Kcal          *float64 `json:"kcal,omitempty"`
	ProteinG      *float64 `json:"protein_g,omitempty"`
	FatG          *float64 `json:"fat_g,omitempty"`
	CarbG         *float64 `json:"carb_g,omitempty"`
	SugarG        *float64 `json:"sugar_g,omitempty"`
	FiberG        *float64 `json:"fiber_g,omitempty"`
}

// RecipeDTO shape.
type RecipeDTO struct {
	ClientID    string `json:"client_id"     validate:"required,uuid"`
	UpdatedAtMs int64  `json:"updated_at_ms" validate:"required"`
	DeletedAtMs *int64 `json:"deleted_at_ms,omitempty"`

	Name       string   `json:"name"              validate:"required,max=255"`
	Servings   float64  `json:"servings"          validate:"required,gt=0"`
	TotalGrams *float64 `json:"total_grams,omitempty"`
	Notes      *string  `json:"notes,omitempty"`
}

// RecipeItemDTO snapshot_json is a JSON object stored verbatim.
type RecipeItemDTO struct {
	ClientID    string `json:"client_id"     validate:"required,uuid"`
	UpdatedAtMs int64  `json:"updated_at_ms" validate:"required"`
	DeletedAtMs *int64 `json:"deleted_at_ms,omitempty"`

	RecipeClientID string  `json:"recipe_client_id" validate:"required,uuid"`
	FoodID         *int64  `json:"food_id,omitempty"`
	NameSnapshot   string  `json:"name_snapshot"    validate:"required,max=255"`
	Grams          float64 `json:"grams"            validate:"required,gt=0"`
	SnapshotJSON   any     `json:"snapshot_json"    validate:"required"`
}

// MealTemplateDTO — payload_json stored verbatim.
type MealTemplateDTO struct {
	ClientID    string `json:"client_id"     validate:"required,uuid"`
	UpdatedAtMs int64  `json:"updated_at_ms" validate:"required"`
	DeletedAtMs *int64 `json:"deleted_at_ms,omitempty"`

	Name        string `json:"name"         validate:"required,max=255"`
	PayloadJSON any    `json:"payload_json" validate:"required"`
}

// FavoriteDTO — food_id references public catalog.
type FavoriteDTO struct {
	ClientID    string `json:"client_id"     validate:"required,uuid"`
	UpdatedAtMs int64  `json:"updated_at_ms" validate:"required"`
	DeletedAtMs *int64 `json:"deleted_at_ms,omitempty"`

	FoodID int64 `json:"food_id" validate:"required,gt=0"`
}

// BatchPayload is the POST body shape.
type BatchPayload struct {
	Meals         []MealDTO         `json:"meals,omitempty"          validate:"max=50,dive"`
	MealItems     []MealItemDTO     `json:"meal_items,omitempty"     validate:"max=200,dive"`
	Recipes       []RecipeDTO       `json:"recipes,omitempty"        validate:"max=50,dive"`
	RecipeItems   []RecipeItemDTO   `json:"recipe_items,omitempty"   validate:"max=200,dive"`
	MealTemplates []MealTemplateDTO `json:"meal_templates,omitempty" validate:"max=50,dive"`
	Favorites     []FavoriteDTO     `json:"favorites,omitempty"      validate:"max=200,dive"`
}

// RowStatus is the per-row outcome in the batch response.
type RowStatus struct {
	Kind     string `json:"kind"`
	ClientID string `json:"client_id"`
	Status   string `json:"status"`
	Reason   string `json:"reason,omitempty"`
	ServerID int64  `json:"server_id,omitempty"`
}

// BatchResponse is the POST response shape.
type BatchResponse struct {
	Results      []RowStatus `json:"results"`
	ServerCursor int64       `json:"server_cursor"`
}

// PullResponse piggybacks BatchPayload with a cursor.
type PullResponse struct {
	BatchPayload
	ServerCursor int64 `json:"server_cursor"`
}

// used to silence unused warnings on the shared base (kept for docs / future inlining).
var _ = base{}
