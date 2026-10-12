package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/thecomputer07/pesa/internal/http/middleware"
	"github.com/thecomputer07/pesa/internal/http/problem"
	"github.com/thecomputer07/pesa/internal/repo"
	"github.com/thecomputer07/pesa/internal/service/cache"
)

// Foods groups the /api/v1/foods endpoints with their DB handle.
type Foods struct {
	DB *sql.DB
	// Cache keeps public responses (anonymous search, public food detail).
	// Never the user's own foods. Noop without Redis.
	Cache cache.Cache
}

// Cache TTLs: the public catalog changes with imports and scanned products.
const (
	searchCacheTTL = 5 * time.Minute
	detailCacheTTL = time.Hour
)

// NewFoods is a tiny constructor used from router.go.
func NewFoods(db *sql.DB) *Foods { return &Foods{DB: db, Cache: cache.Noop{}} }

type foodsSourceDTO struct {
	Code       string  `json:"code"`
	License    string  `json:"license"`
	LicenseURL *string `json:"license_url,omitempty"`
}

type foodsSearchResultDTO struct {
	ID       int64          `json:"id"`
	NameIT   string         `json:"name_it"`
	NameEN   *string        `json:"name_en,omitempty"`
	Category *string        `json:"category,omitempty"`
	State    string         `json:"state"`
	Kcal     *float64       `json:"kcal,omitempty"`
	CarbG    *float64       `json:"carb_g,omitempty"`
	ProteinG *float64       `json:"protein_g,omitempty"`
	FatG     *float64       `json:"fat_g,omitempty"`
	SugarG   *float64       `json:"sugar_g,omitempty"`
	FiberG   *float64       `json:"fiber_g,omitempty"`
	ThumbURL *string        `json:"thumb_url,omitempty"`
	Source   foodsSourceDTO `json:"source"`
	// Set on /foods/mine?recipes=1 rows that stand for one of the user's recipes.
	RecipeClientID *string `json:"recipe_client_id,omitempty"`
}

type foodsSearchResponseDTO struct {
	Results []foodsSearchResultDTO `json:"results"`
	Total   int                    `json:"total"`
	Page    int                    `json:"page"`
	Limit   int                    `json:"limit"`
}

// Search handles GET /api/v1/foods/search?q=&limit=&page=.
func (h *Foods) Search(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		problem.Write(w, r, http.StatusBadRequest, "invalid-query", "q parameter is required")
		return
	}
	if len(q) > 200 {
		problem.Write(w, r, http.StatusBadRequest, "invalid-query", "q exceeds 200 chars")
		return
	}

	limit := parseIntDefault(r.URL.Query().Get("limit"), 20)
	if limit < 1 {
		limit = 1
	}
	if limit > 50 {
		limit = 50
	}
	page := parseIntDefault(r.URL.Query().Get("page"), 1)
	if page < 1 {
		page = 1
	}

	// Optional source filter: "", "crea", "ciqual", "usda", "off", "mine" (own custom foods).
	source := r.URL.Query().Get("source")
	switch source {
	case "", "crea", "ciqual", "usda", "off", "mine":
	default:
		problem.Write(w, r, http.StatusBadRequest, "invalid-query", "unknown source")
		return
	}

	userID := middleware.UserIDFrom(r.Context())
	// Anonymous results are the same for everyone: shared cache.
	cacheKey := ""
	if userID == 0 && source != "mine" {
		cacheKey = fmt.Sprintf("search:%s:%d:%d:%s", source, limit, page, strings.Join(strings.Fields(strings.ToLower(q)), " "))
		if body, ok := h.Cache.Get(r.Context(), cacheKey); ok {
			writeSearchBody(w, userID, body)
			return
		}
	}
	results, total, err := repo.SearchFoodsFiltered(r.Context(), h.DB, userID, q, source, limit, page)
	if err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "internal-error", "search failed")
		return
	}

	resp := foodsSearchResponseDTO{
		Results: make([]foodsSearchResultDTO, 0, len(results)),
		Total:   total,
		Page:    page,
		Limit:   limit,
	}
	for _, f := range results {
		resp.Results = append(resp.Results, toSearchDTO(f))
	}
	body, err := json.Marshal(resp)
	if err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "internal-error", "search failed")
		return
	}
	if cacheKey != "" {
		h.Cache.Set(r.Context(), cacheKey, body, searchCacheTTL)
	}
	writeSearchBody(w, userID, body)
}

// writeSearchBody: logged in, results include the user's own foods: no browser
// cache, or a food just created or deleted would show up late. Anonymous: 30 s.
func writeSearchBody(w http.ResponseWriter, userID int64, body []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if userID > 0 {
		w.Header().Set("Cache-Control", "private, no-store")
	} else {
		w.Header().Set("Cache-Control", "private, max-age=30, must-revalidate")
	}
	w.Header().Set("Vary", "Cookie")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(append(body, '\n'))
}

// Mine handles GET /api/v1/foods/mine — lista dei custom food dell'utente.
func (h *Foods) Mine(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFrom(r.Context())
	if userID <= 0 {
		problem.Write(w, r, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	// recipes=1 (offline copy for local search) also returns the foods that
	// stand for recipes; the "my foods" page leaves them to the recipes page.
	withRecipes := r.URL.Query().Get("recipes") == "1"
	rows, err := h.DB.QueryContext(r.Context(), `
		SELECT f.id, f.name_it, f.name_en, f.category, f.state,
		       f.kcal, f.carb_g, f.protein_g, f.fat_g, f.sugar_g, f.fiber_g,
		       s.code, s.license, s.license_url,
		       (SELECT url FROM food_images WHERE food_id = f.id ORDER BY id LIMIT 1) AS thumb_url,
		       rc.client_id
		  FROM foods f
		  JOIN food_sources s ON s.id = f.source_id
		  LEFT JOIN recipes rc ON rc.id = f.recipe_id
		 WHERE f.user_id = $1
		   AND ($2 OR f.recipe_id IS NULL)
		   AND f.superseded_by IS NULL
		 ORDER BY f.id DESC`, userID, withRecipes)
	if err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	defer func() { _ = rows.Close() }()

	out := make([]foodsSearchResultDTO, 0)
	for rows.Next() {
		var f repo.Food
		if err := rows.Scan(
			&f.ID, &f.NameIT, &f.NameEN, &f.Category, &f.State,
			&f.Kcal, &f.CarbG, &f.ProteinG, &f.FatG, &f.SugarG, &f.FiberG,
			&f.SourceCode, &f.License, &f.LicenseURL, &f.ThumbURL, &f.RecipeClientID,
		); err != nil {
			problem.Write(w, r, http.StatusInternalServerError, "internal", err.Error())
			return
		}
		dto := toSearchDTO(f)
		dto.RecipeClientID = nullStrPtr(f.RecipeClientID)
		out = append(out, dto)
	}
	resp := foodsSearchResponseDTO{Results: out, Total: len(out), Page: 1, Limit: len(out)}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Vary", "Cookie")
	_ = json.NewEncoder(w).Encode(resp)
}

// customFoodInput è il body JSON accettato da POST /api/v1/foods/custom.
type customFoodInput struct {
	NameIT    string   `json:"name_it"`
	Category  string   `json:"category"`
	State     string   `json:"state"` // "raw" | "cooked"
	EdiblePct *float64 `json:"edible_pct"`
	Kcal      *float64 `json:"kcal"`
	ProteinG  *float64 `json:"protein_g"`
	FatG      *float64 `json:"fat_g"`
	CarbG     *float64 `json:"carb_g"`
	SugarG    *float64 `json:"sugar_g"`
	FiberG    *float64 `json:"fiber_g"`
	SaltG     *float64 `json:"salt_g"`
}

// Create handles POST /api/v1/foods/custom — crea un alimento personalizzato
// legato all'utente (source "manual", user_id = current user).
func (h *Foods) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFrom(r.Context())
	if userID <= 0 {
		problem.Write(w, r, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	var in customFoodInput
	if err := json.NewDecoder(io.LimitReader(r.Body, 4*1024)).Decode(&in); err != nil {
		problem.Write(w, r, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	in.NameIT = strings.TrimSpace(in.NameIT)
	if in.NameIT == "" || len(in.NameIT) > 200 {
		problem.Write(w, r, http.StatusBadRequest, "invalid-name", "nome richiesto (max 200 char)")
		return
	}
	state := strings.ToLower(strings.TrimSpace(in.State))
	if state != "raw" && state != "cooked" {
		state = "raw"
	}
	edible := 100.0
	if in.EdiblePct != nil && *in.EdiblePct > 0 && *in.EdiblePct <= 100 {
		edible = *in.EdiblePct
	}
	if in.CarbG == nil {
		problem.Write(w, r, http.StatusBadRequest, "carb-required", "carboidrati (carb_g) richiesti")
		return
	}

	var sourceID int64
	if err := h.DB.QueryRowContext(r.Context(),
		`SELECT id FROM food_sources WHERE code = 'manual'`).Scan(&sourceID); err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "source-missing", "manual source not found")
		return
	}

	var newID int64
	err := h.DB.QueryRowContext(r.Context(), `
		INSERT INTO foods (source_id, user_id, name_it, category, state, edible_pct,
		                   kcal, protein_g, fat_g, carb_g, sugar_g, fiber_g, salt_g,
		                   last_imported_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, CURRENT_TIMESTAMP)
		RETURNING id`,
		sourceID, userID, in.NameIT, nullString(in.Category), state, edible,
		nullFloatInput(in.Kcal), nullFloatInput(in.ProteinG), nullFloatInput(in.FatG),
		in.CarbG, nullFloatInput(in.SugarG), nullFloatInput(in.FiberG), nullFloatInput(in.SaltG),
	).Scan(&newID)
	if err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "insert-failed", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": newID})
}

// Update handles PATCH /api/v1/foods/custom/{id} — aggiorna un custom food
// dell'utente corrente. Rifiuta se l'alimento non è suo.
func (h *Foods) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFrom(r.Context())
	if userID <= 0 {
		problem.Write(w, r, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		problem.Write(w, r, http.StatusBadRequest, "invalid-id", "id must be positive integer")
		return
	}

	// Verifica proprietà
	var ownerID sql.NullInt64
	if err := h.DB.QueryRowContext(r.Context(),
		`SELECT user_id FROM foods WHERE id = $1`, id).Scan(&ownerID); err != nil {
		if err == sql.ErrNoRows {
			problem.Write(w, r, http.StatusNotFound, "not-found", "food not found")
			return
		}
		problem.Write(w, r, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	if !ownerID.Valid || ownerID.Int64 != userID {
		problem.Write(w, r, http.StatusForbidden, "forbidden", "not your food")
		return
	}
	if isRecipeFood(r.Context(), h.DB, id) {
		problem.Write(w, r, http.StatusConflict, "recipe-food", "Questo alimento è una ricetta: modificala dalla pagina Ricette.")
		return
	}

	var in customFoodInput
	if err := json.NewDecoder(io.LimitReader(r.Body, 4*1024)).Decode(&in); err != nil {
		problem.Write(w, r, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	in.NameIT = strings.TrimSpace(in.NameIT)
	if in.NameIT == "" || len(in.NameIT) > 200 {
		problem.Write(w, r, http.StatusBadRequest, "invalid-name", "nome richiesto (max 200 char)")
		return
	}
	state := strings.ToLower(strings.TrimSpace(in.State))
	if state != "raw" && state != "cooked" {
		state = "raw"
	}
	edible := 100.0
	if in.EdiblePct != nil && *in.EdiblePct > 0 && *in.EdiblePct <= 100 {
		edible = *in.EdiblePct
	}
	if in.CarbG == nil {
		problem.Write(w, r, http.StatusBadRequest, "carb-required", "carboidrati (carb_g) richiesti")
		return
	}

	_, err = h.DB.ExecContext(r.Context(), `
		UPDATE foods SET name_it=$1, category=$2, state=$3, edible_pct=$4,
		                 kcal=$5, protein_g=$6, fat_g=$7, carb_g=$8,
		                 sugar_g=$9, fiber_g=$10, salt_g=$11,
		                 updated_at=CURRENT_TIMESTAMP
		 WHERE id=$12 AND user_id=$13`,
		in.NameIT, nullString(in.Category), state, edible,
		nullFloatInput(in.Kcal), nullFloatInput(in.ProteinG), nullFloatInput(in.FatG),
		in.CarbG, nullFloatInput(in.SugarG), nullFloatInput(in.FiberG), nullFloatInput(in.SaltG),
		id, userID)
	if err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "update-failed", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Delete handles DELETE /api/v1/foods/custom/{id}.
func (h *Foods) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFrom(r.Context())
	if userID <= 0 {
		problem.Write(w, r, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		problem.Write(w, r, http.StatusBadRequest, "invalid-id", "id must be positive integer")
		return
	}
	if isRecipeFood(r.Context(), h.DB, id) {
		problem.Write(w, r, http.StatusConflict, "recipe-food", "Questo alimento è una ricetta: eliminala dalla pagina Ricette.")
		return
	}
	res, err := h.DB.ExecContext(r.Context(),
		`DELETE FROM foods WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		// meal_items / recipe_items reference foods without cascade: a food
		// still used in the diary or a recipe cannot go (23503 = FK violation).
		if strings.Contains(err.Error(), "23503") {
			problem.Write(w, r, http.StatusConflict, "food-in-use",
				"Questo alimento è usato nel diario o in una ricetta: toglilo da lì prima di eliminarlo.")
			return
		}
		problem.Write(w, r, http.StatusInternalServerError, "delete-failed", "delete failed")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		problem.Write(w, r, http.StatusForbidden, "forbidden", "not your food or not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// isRecipeFood reports whether the food stands for a recipe (migration 00036).
func isRecipeFood(ctx context.Context, db *sql.DB, id int64) bool {
	var n int
	_ = db.QueryRowContext(ctx, `SELECT count(*) FROM foods WHERE id = $1 AND recipe_id IS NOT NULL`, id).Scan(&n)
	return n > 0
}

func nullString(s string) interface{} {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return s
}

func nullFloatInput(p *float64) interface{} {
	if p == nil {
		return nil
	}
	return *p
}

// Detail handles GET /api/v1/foods/{id}.
func (h *Foods) Detail(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		problem.Write(w, r, http.StatusBadRequest, "invalid-id", "id must be a positive integer")
		return
	}
	// Only public foods are ever stored under this key (see below).
	cacheKey := "food:" + strconv.FormatInt(id, 10)
	if body, ok := h.Cache.Get(r.Context(), cacheKey); ok {
		writePublicJSON(w, body)
		return
	}
	food, imgs, err := repo.GetFoodByID(r.Context(), h.DB, id)
	if err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "internal-error", "detail failed")
		return
	}
	// A custom food is private: only its owner sees it (404 for anyone else,
	// so its existence doesn't leak either), and never via a shared cache.
	userID := middleware.UserIDFrom(r.Context())
	if food == nil || (food.UserID.Valid && food.UserID.Int64 != userID) {
		problem.Write(w, r, http.StatusNotFound, "food-not-found", "no food with that id")
		return
	}
	siblings, _ := repo.LoadSiblings(r.Context(), h.DB, food.ID, food.ParentFoodID)
	dto := toDetailDTO(food, imgs, siblings)
	if food.UserID.Valid {
		dto.IsMine = true
		if food.RecipeClientID.Valid {
			dto.RecipeClientID = strings.TrimSpace(food.RecipeClientID.String)
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "private, no-store")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(dto)
		return
	}
	body, err := json.Marshal(dto)
	if err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "internal-error", "detail failed")
		return
	}
	h.Cache.Set(r.Context(), cacheKey, body, detailCacheTTL)
	writePublicJSON(w, body)
}

// writePublicJSON writes catalog data that is the same for everyone.
func writePublicJSON(w http.ResponseWriter, body []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// Public catalog data: browsers 5 min, Cloudflare 1 h.
	w.Header().Set("Cache-Control", "public, max-age=300, s-maxage=3600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(append(body, '\n'))
}

// --- helpers ---

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func nullStrPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}
func nullFloatPtr(nf sql.NullFloat64) *float64 {
	if !nf.Valid {
		return nil
	}
	v := nf.Float64
	return &v
}

func toSearchDTO(f repo.Food) foodsSearchResultDTO {
	return foodsSearchResultDTO{
		ID:       f.ID,
		NameIT:   f.NameIT,
		NameEN:   nullStrPtr(f.NameEN),
		Category: nullStrPtr(f.Category),
		State:    f.State,
		Kcal:     nullFloatPtr(f.Kcal),
		CarbG:    nullFloatPtr(f.CarbG),
		ProteinG: nullFloatPtr(f.ProteinG),
		FatG:     nullFloatPtr(f.FatG),
		SugarG:   nullFloatPtr(f.SugarG),
		FiberG:   nullFloatPtr(f.FiberG),
		ThumbURL: nullStrPtr(f.ThumbURL),
		Source: foodsSourceDTO{
			Code:       f.SourceCode,
			License:    f.License,
			LicenseURL: nullStrPtr(f.LicenseURL),
		},
	}
}

type foodImageDTO struct {
	URL         string  `json:"url"`
	Attribution *string `json:"attribution,omitempty"`
	License     *string `json:"license,omitempty"`
	IsExternal  bool    `json:"is_external"`
}

type foodDetailDTO struct {
	ID             int64            `json:"id"`
	NameIT         string           `json:"name_it"`
	NameEN         *string          `json:"name_en,omitempty"`
	Category       *string          `json:"category,omitempty"`
	State          string           `json:"state"`
	EdiblePct      float64          `json:"edible_pct"`
	Kcal           *float64         `json:"kcal,omitempty"`
	ProteinG       *float64         `json:"protein_g,omitempty"`
	FatG           *float64         `json:"fat_g,omitempty"`
	CarbG          *float64         `json:"carb_g,omitempty"`
	SugarG         *float64         `json:"sugar_g,omitempty"`
	FiberG         *float64         `json:"fiber_g,omitempty"`
	WaterG         *float64         `json:"water_g,omitempty"`
	AshG           *float64         `json:"ash_g,omitempty"`
	AlcoholG       *float64         `json:"alcohol_g,omitempty"`
	SaltG          *float64         `json:"salt_g,omitempty"`
	Attribution    string           `json:"attribution"`
	License        string           `json:"license"`
	LicenseURL     *string          `json:"license_url,omitempty"`
	SourceURL      *string          `json:"source_url,omitempty"`
	LastImportedAt *string          `json:"last_imported_at,omitempty"`
	Images         []foodImageDTO   `json:"images"`
	Siblings       []foodSiblingDTO `json:"siblings"`
	// IsMine: the current user's own custom food (can edit / delete).
	IsMine bool `json:"is_mine,omitempty"`
	// RecipeClientID: the food stands for this recipe; edit the recipe instead.
	RecipeClientID string `json:"recipe_client_id,omitempty"`
}

type foodSiblingDTO struct {
	ID     int64  `json:"id"`
	NameIT string `json:"name_it"`
	State  string `json:"state"`
}

func toDetailDTO(f *repo.Food, imgs []repo.FoodImage, siblings []repo.FoodSibling) foodDetailDTO {
	var lastImp *string
	if f.LastImportedAt.Valid {
		s := f.LastImportedAt.Time.UTC().Format("2006-01-02T15:04:05Z")
		lastImp = &s
	}
	out := foodDetailDTO{
		ID:             f.ID,
		NameIT:         f.NameIT,
		NameEN:         nullStrPtr(f.NameEN),
		Category:       nullStrPtr(f.Category),
		State:          f.State,
		EdiblePct:      f.EdiblePct,
		Kcal:           nullFloatPtr(f.Kcal),
		ProteinG:       nullFloatPtr(f.ProteinG),
		FatG:           nullFloatPtr(f.FatG),
		CarbG:          nullFloatPtr(f.CarbG),
		SugarG:         nullFloatPtr(f.SugarG),
		FiberG:         nullFloatPtr(f.FiberG),
		WaterG:         nullFloatPtr(f.WaterG),
		AshG:           nullFloatPtr(f.AshG),
		AlcoholG:       nullFloatPtr(f.AlcoholG),
		SaltG:          nullFloatPtr(f.SaltG),
		Attribution:    f.AttributionText,
		License:        f.License,
		LicenseURL:     nullStrPtr(f.LicenseURL),
		SourceURL:      nullStrPtr(f.SourceURL),
		LastImportedAt: lastImp,
		Images:         make([]foodImageDTO, 0, len(imgs)),
		Siblings:       make([]foodSiblingDTO, 0, len(siblings)),
	}
	for _, im := range imgs {
		out.Images = append(out.Images, foodImageDTO{
			URL:         im.URL,
			Attribution: nullStrPtr(im.Attribution),
			License:     nullStrPtr(im.License),
			IsExternal:  im.IsExternal,
		})
	}
	for _, s := range siblings {
		out.Siblings = append(out.Siblings, foodSiblingDTO{ID: s.ID, NameIT: s.NameIT, State: s.State})
	}
	return out
}
